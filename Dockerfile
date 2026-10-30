#
# Backend (Cross-compilation using host platform without QEMU)
#
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS server

ARG TARGETOS
ARG TARGETARCH

WORKDIR /server

# Cache Go modules layer
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy server source code
COPY server/ ./

# Cross-compile for target architecture (runs natively on build machine, super fast)
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -ldflags="-s -w" -o /server/watcharr .

#
# Frontend (Static assets build runs natively on host platform)
#
FROM --platform=$BUILDPLATFORM node:22-alpine AS ui

WORKDIR /app

# Cache package dependencies layer
COPY package*.json ./
RUN npm install

# Copy source and build static bundle
COPY ./src ./src
COPY ./static ./static
COPY vite.config.ts svelte.config.js tsconfig.json ./

RUN npm run build

#
# Production (Target architecture runner)
#
FROM alpine:3.21 AS runner

WORKDIR /

# Install ca-certificates and tzdata for HTTPS & timezones
RUN apk --no-cache add ca-certificates tzdata

ENV GOMEMLIMIT=35MiB \
    GODEBUG=madvdontneed=1

COPY --from=server /server/watcharr /watcharr
COPY --from=ui /app/build /ui

EXPOSE 3080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:${PORT:-${WATCHARR_PORT:-3080}}/api/health" || exit 1

CMD ["/watcharr"]
