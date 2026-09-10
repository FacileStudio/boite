FROM nginx:alpine
COPY index.html /usr/share/nginx/html/index.html
COPY nginx.conf /etc/nginx/conf.d/default.conf

# The ~2 GB base image is served from the host filesystem via a bind mount
# rather than committed to this image or the repo. Mount a ro directory that
# holds boite-base.qcow2 over /assets, e.g.
#   -v /srv/boite/assets:/assets:ro
EXPOSE 80