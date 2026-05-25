# syntax=docker/dockerfile:1.7

# ── builder ─────────────────────────────────────────────────────────────────
FROM node:22-alpine AS builder
WORKDIR /src

# Cache the dependency install layer separately from the source copy so that
# editing source files doesn't re-download packages.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline

COPY . .
RUN npm run build

# ── runtime ─────────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine
RUN addgroup -S app && adduser -S -G app app

COPY --from=builder /src/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- --tries=1 --timeout=2 http://127.0.0.1:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
