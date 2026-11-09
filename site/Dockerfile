FROM oven/bun:1 AS build
WORKDIR /app
COPY package.json bun.lock ./
RUN for i in 1 2 3 4 5 6; do bun install --frozen-lockfile && exit 0; echo "install failed, retry $i"; sleep $((i * 10)); done; exit 1
COPY . .
RUN bun run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget -qO- http://127.0.0.1/ >/dev/null || exit 1
