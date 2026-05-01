import { Link } from 'react-router-dom'
import { ROUTES } from '../../shared/constants/routes.js'

export function NotFoundPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-100 p-4">
      <section className="max-w-md rounded-lg border border-slate-200 bg-white p-8 text-center shadow-lg">
        <h1 className="text-3xl font-black text-slate-950">404</h1>
        <p className="mt-2 text-slate-600">Route không tồn tại.</p>
        <Link
          className="mt-6 inline-flex min-h-10 items-center justify-center rounded-md bg-sky-700 px-4 py-2 text-sm font-bold text-white hover:bg-sky-800"
          to={ROUTES.dashboard}
        >
          Về dashboard
        </Link>
      </section>
    </main>
  )
}
