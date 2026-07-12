# syntax=docker/dockerfile:1
#
# Self-building RUNTIME image for inflow-inspector-api.
#
# This image ships only the Go toolchain + git + an entrypoint — it does NOT
# compile anything at `docker build` time. At container START the entrypoint
# clones this repo (at $INSPECTOR_API_REF), compiles a binary native to the
# arch it's actually running on, and execs it. That's why a single published
# image runs on both amd64 and arm64 with no cross-compiling or buildx tricks:
# the compile happens on the target machine, at run time.
#
# Because nothing is copied in, the build context is irrelevant — you can build
# and publish this from an empty directory:
#   docker buildx build --platform linux/amd64,linux/arm64 \
#     -t mehdishokohi/inflow-inspector-api:latest --push - < Dockerfile
#
# Persist /src (and optionally /go) on a volume so restarts skip the re-clone
# and module download — see the getting-started compose file.

FROM golang:1.26-alpine

RUN apk add --no-cache git

ENV INSPECTOR_API_REPO=https://github.com/Inflowenger/inflow-inspector-api.git \
    INSPECTOR_API_REF=master \
    SRC_DIR=/src \
    PORT=8025

COPY <<'EOF' /usr/local/bin/entrypoint.sh
#!/bin/sh
set -e
if [ -d "$SRC_DIR/.git" ]; then
  echo "[entrypoint] updating $SRC_DIR -> $INSPECTOR_API_REF"
  git -C "$SRC_DIR" fetch --depth 1 origin "$INSPECTOR_API_REF"
  git -C "$SRC_DIR" checkout -q FETCH_HEAD
else
  echo "[entrypoint] cloning $INSPECTOR_API_REPO @ $INSPECTOR_API_REF"
  git clone --depth 1 -b "$INSPECTOR_API_REF" "$INSPECTOR_API_REPO" "$SRC_DIR"
fi
cd "$SRC_DIR"
echo "[entrypoint] building for $(go env GOOS)/$(go env GOARCH)"
CGO_ENABLED=0 go build -trimpath -o /app/inflow-inspector-api .
echo "[entrypoint] starting inflow-inspector-api on :$PORT"
exec /app/inflow-inspector-api
EOF
RUN chmod +x /usr/local/bin/entrypoint.sh

EXPOSE 8025
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
