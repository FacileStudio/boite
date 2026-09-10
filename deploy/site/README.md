# Boite landing page

Static nginx site for `boite.facile.studio`. Serves the landing page and the
~2 GB baked base image at `/base.qcow2`.

## Layout

- `index.html` — the landing page. The download button points at `./base.qcow2`.
- `nginx.conf` — serves `/base.qcow2` from the bind-mounted `/assets` path and
  sets a `Content-Disposition` so the download keeps its filename.
- `Dockerfile` — nginx:alpine, copies the static site and nginx config.

## Serving the base image

The ~2 GB `boite-base.qcow2` is deliberately **not** committed to the repo or
built into the image (it would bloat every deploy). It is served from a host
directory bind-mounted read-only into the container:

```
-v <host-dir-with-image>:/assets:ro
```

On la ruche the image is available at:

```
/etc/dokploy/boite-assets/boite-base.qcow2
```

so the Dokploy app should mount `/etc/dokploy/boite-assets:/assets:ro`.

## Build

```
docker build -t boite-site .
docker run --rm -p 8080:80 -v /etc/dokploy/boite-assets:/assets:ro boite-site
# -> http://localhost:8080/         landing page
# -> http://localhost:8080/base.qcow2  the image download
```

## Sha verification

The CLI pins the image by SHA-256 in `cmd/qemu/instance.go`. After a rebuild,
update `BaseImageSHA256` there and the running site stays correct because the
bind mount points at the same named file.