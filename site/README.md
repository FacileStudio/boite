# Boite landing page

Static SvelteKit site for `boite.facile.studio` built with `@facile/muse` and `@facile/blob`. Serves the product landing page and the ~2 GB baked base image at `/base.qcow2`.

## Development

```sh
bun install
bun run dev
```

Build the static distribution:

```sh
bun run build
bun run check
```

## Docker build

The multi-stage Dockerfile compiles the static site with Bun and serves it with Nginx 1.27.

```sh
docker build -t boite-site .
docker run --rm -p 8080:80 -v /etc/dokploy/boite-assets:/assets:ro boite-site
```

The ~2 GB base image `boite.qcow2` is bind-mounted at runtime over `/assets` and served via Nginx at `/base.qcow2`.
