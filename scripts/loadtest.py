"""Async load test for the GoProject backend.

Fires 100 concurrent requests across the auth → folder → note read/write flow
to exercise the cache layer (cold writes followed by hot read replays) and
report per-endpoint latency.

Usage:
    pip install -r scripts/requirements.txt
    python scripts/loadtest.py                # hits http://localhost:8080
    BASE_URL=http://host:8080 python scripts/loadtest.py
"""

from __future__ import annotations

import asyncio
import os
import random
import statistics
import string
import time
import uuid
from collections import defaultdict
from dataclasses import dataclass, field
from typing import Any

import httpx

BASE_URL = os.getenv("BASE_URL", "http://localhost:8080").rstrip("/") + "/api"
# Backend caps /auth/login and /users/register at 5/min/IP, so we top out the
# write phase at 5 users. The read fanout then carries the rest of the run.
N_USERS = int(os.getenv("N_USERS", "5"))
WRITE_CONCURRENCY = int(os.getenv("WRITE_CONCURRENCY", "3"))
# Total target requests; the read fanout fills whatever's left after setup.
TOTAL_REQUESTS = int(os.getenv("REQUESTS", "100"))
# Cap concurrent open sockets so we don't hit ulimit -n at high request counts.
READ_CONCURRENCY = int(os.getenv("READ_CONCURRENCY", "100"))
PASSWORD = "Passw0rd!"
TIMEOUT = httpx.Timeout(connect=5.0, read=30.0, write=30.0, pool=30.0)


@dataclass
class Sample:
    endpoint: str
    method: str
    status: int
    elapsed_ms: float
    ok: bool


@dataclass
class Stats:
    samples: list[Sample] = field(default_factory=list)

    def add(self, s: Sample) -> None:
        self.samples.append(s)

    def report(self) -> None:
        if not self.samples:
            print("(no samples)")
            return

        by_endpoint: dict[str, list[Sample]] = defaultdict(list)
        for s in self.samples:
            by_endpoint[f"{s.method} {s.endpoint}"].append(s)

        total = len(self.samples)
        ok = sum(1 for s in self.samples if s.ok)
        fail = total - ok

        print()
        print(f"Total requests : {total}")
        print(f"  succeeded    : {ok}")
        print(f"  failed       : {fail}")
        print()
        header = f"{'Endpoint':<40} {'N':>4} {'OK':>4} {'p50':>8} {'p90':>8} {'p95':>8} {'p99':>8} {'max':>8}"
        print(header)
        print("-" * len(header))
        for endpoint, samples in sorted(by_endpoint.items()):
            elapsed = [s.elapsed_ms for s in samples]
            ok_count = sum(1 for s in samples if s.ok)
            print(
                f"{endpoint:<40} "
                f"{len(samples):>4} "
                f"{ok_count:>4} "
                f"{pct(elapsed, 50):>7.1f}ms "
                f"{pct(elapsed, 90):>7.1f}ms "
                f"{pct(elapsed, 95):>7.1f}ms "
                f"{pct(elapsed, 99):>7.1f}ms "
                f"{max(elapsed):>7.1f}ms"
            )


def pct(values: list[float], p: int) -> float:
    if not values:
        return 0.0
    return statistics.quantiles(values, n=100, method="inclusive")[p - 1] if len(values) > 1 else values[0]


def rand_suffix(n: int = 8) -> str:
    return "".join(random.choices(string.ascii_lowercase + string.digits, k=n))


async def timed_request(
    client: httpx.AsyncClient,
    stats: Stats,
    method: str,
    path: str,
    *,
    json: Any = None,
    headers: dict[str, str] | None = None,
) -> httpx.Response | None:
    """Issue one HTTP call, record timing in stats, and return the response.

    Network failures are caught and reported as a fake 0-status sample so a
    single broken request doesn't tank the whole run.
    """
    start = time.perf_counter()
    try:
        resp = await client.request(method, path, json=json, headers=headers)
        elapsed = (time.perf_counter() - start) * 1000
        stats.add(Sample(path, method, resp.status_code, elapsed, resp.is_success))
        return resp
    except (httpx.RequestError, httpx.HTTPError) as exc:
        elapsed = (time.perf_counter() - start) * 1000
        stats.add(Sample(path, method, 0, elapsed, False))
        print(f"[error] {method} {path}: {exc!r}")
        return None


async def register_user(client: httpx.AsyncClient, stats: Stats, idx: int) -> dict[str, str]:
    suffix = rand_suffix()
    user = {
        "username": f"loadtest_{idx}_{suffix}",
        "email": f"loadtest_{idx}_{suffix}@example.com",
        "password": PASSWORD,
        "role": "member",
    }
    await timed_request(client, stats, "POST", "/users/register", json=user)
    return user


async def login(client: httpx.AsyncClient, stats: Stats, user: dict[str, str]) -> str | None:
    resp = await timed_request(
        client, stats, "POST", "/auth/login",
        json={"email": user["email"], "password": user["password"]},
    )
    if resp is None or not resp.is_success:
        return None
    return resp.json().get("token")


async def create_folder(client: httpx.AsyncClient, stats: Stats, token: str) -> int | None:
    resp = await timed_request(
        client, stats, "POST", "/folders",
        json={"name": f"folder-{rand_suffix()}"},
        headers={"Authorization": f"Bearer {token}"},
    )
    if resp is None or not resp.is_success:
        return None
    body = resp.json()
    # The folder handler returns the model directly; ID lives under "id".
    return body.get("id") or (body.get("folder") or {}).get("id")


async def create_note(client: httpx.AsyncClient, stats: Stats, token: str, folder_id: int) -> int | None:
    resp = await timed_request(
        client, stats, "POST", f"/folders/{folder_id}/notes",
        json={"title": f"note-{rand_suffix()}", "content": f"content-{uuid.uuid4()}"},
        headers={"Authorization": f"Bearer {token}"},
    )
    if resp is None or not resp.is_success:
        return None
    body = resp.json()
    return body.get("id") or (body.get("note") or {}).get("id")


async def read_phase(
    client: httpx.AsyncClient,
    stats: Stats,
    sessions: list[tuple[str, int, int]],
    n: int,
    concurrency: int,
) -> None:
    """Fire `n` random read requests across the cached endpoints in parallel.

    Each session is (token, folder_id, note_id). Same folder is requested
    multiple times by design so the asset/ACL caches get exercised. The
    semaphore caps how many sockets are open at once so the run stays under
    the host's file-descriptor limit at high request counts.
    """
    sem = asyncio.Semaphore(concurrency)

    async def one() -> None:
        async with sem:
            token, folder_id, note_id = random.choice(sessions)
            headers = {"Authorization": f"Bearer {token}"}
            kind = random.choice(["folder", "list_folders", "list_notes", "note"])
            if kind == "folder":
                await timed_request(client, stats, "GET", f"/folders/{folder_id}", headers=headers)
            elif kind == "list_folders":
                await timed_request(client, stats, "GET", "/folders", headers=headers)
            elif kind == "list_notes":
                await timed_request(client, stats, "GET", f"/folders/{folder_id}/notes", headers=headers)
            else:
                await timed_request(client, stats, "GET", f"/notes/{note_id}", headers=headers)

    await asyncio.gather(*(one() for _ in range(n)))


async def main() -> None:
    random.seed()  # fresh users every run
    stats = Stats()
    overall_start = time.perf_counter()

    write_sem = asyncio.Semaphore(WRITE_CONCURRENCY)

    async def gated(coro):
        async with write_sem:
            return await coro

    limits = httpx.Limits(max_connections=READ_CONCURRENCY, max_keepalive_connections=READ_CONCURRENCY)
    async with httpx.AsyncClient(base_url=BASE_URL, timeout=TIMEOUT, limits=limits) as client:
        # ── Phase 1: register users (gated so the per-endpoint limiter doesn't trip) ──
        users = await asyncio.gather(*(gated(register_user(client, stats, i)) for i in range(N_USERS)))

        # ── Phase 2: login each user, collect tokens ────────────────────────
        tokens = await asyncio.gather(*(gated(login(client, stats, u)) for u in users))
        sessions_partial = [t for t in tokens if t]
        if not sessions_partial:
            print("no successful logins; aborting")
            stats.report()
            return

        # ── Phase 3: create one folder per session ──────────────────────────
        folder_ids = await asyncio.gather(
            *(gated(create_folder(client, stats, t)) for t in sessions_partial)
        )

        # ── Phase 4: create one note per (token, folder) pair ───────────────
        note_tasks = [
            gated(create_note(client, stats, t, fid))
            for t, fid in zip(sessions_partial, folder_ids)
            if fid is not None
        ]
        note_ids = await asyncio.gather(*note_tasks)

        # Assemble usable sessions for the read phase.
        sessions: list[tuple[str, int, int]] = [
            (t, fid, nid)
            for t, fid, nid in zip(sessions_partial, folder_ids, note_ids)
            if fid is not None and nid is not None
        ]
        if not sessions:
            print("no usable sessions after setup; aborting")
            stats.report()
            return

        # ── Phase 5: hot read load fills the remainder up to TOTAL_REQUESTS ─
        read_fanout = max(0, TOTAL_REQUESTS - len(stats.samples))
        await read_phase(client, stats, sessions, read_fanout, READ_CONCURRENCY)

    duration = time.perf_counter() - overall_start
    rps = len(stats.samples) / duration if duration else 0
    print(f"Wall time      : {duration:.2f}s")
    print(f"Throughput     : {rps:.1f} req/s")
    stats.report()


if __name__ == "__main__":
    asyncio.run(main())
