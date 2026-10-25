#
# Backend
#
FROM golang:1.27-alpine AS server

WORKDIR /server

# Cache Go modules layer
COPY server/go.mod server/go.sum ./
RUN go mod download

# Copy server source code
COPY server/ ./

# Build optimized binary
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -o ./watcharr .

#
# Frontend
#
FROM node:24-alpine AS ui

WORKDIR /app

# Cache package dependencies layer
COPY package*.json ./
RUN npm install

# Copy source and build
COPY ./src ./src
COPY ./static ./static
COPY vite.config.ts svelte.config.js tsconfig.json ./

RUN npm run build

#
# Production (Ultra-lightweight Alpine image: Zero Node.js runtime)
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
  CMD wget -qO- http://127.0.0.1:3080/api/health || exit 1

CMD ["/watcharr"]
