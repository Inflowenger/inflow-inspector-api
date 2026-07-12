# syntax=docker/dockerfile:1
#
# Builds the Inflowenger inflow-inspector-api as a self-contained image.
#
# NOTE: this assumes github.com/Inflowenger/inflow-fusion resolves as a normal
# Go module (public repo, or private with GOPRIVATE + credentials). While the
# local `replace` in go.mod points at a sibling checkout, comment it out first —
# a standalone build can't see a path outside its build context.

FROM golang:1.26-alpine AS build
RUN apk add --no-cache git
WORKDIR /src

# Cache modules against the manifests first.
COPY go.mod go.sum ./
RUN go mod download

# Build.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/inflow-inspector-api .

# ── runtime ──────────────────────────────────────────────────────────────────
# Root (default) distroless so the BadgerDB store under /data is writable.
FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /out/inflow-inspector-api /app/inflow-inspector-api
EXPOSE 8025
ENTRYPOINT ["/app/inflow-inspector-api"]
