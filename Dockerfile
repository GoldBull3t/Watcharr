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
RUN npm ci

# Copy source and build
COPY ./src ./src
COPY ./static ./static
COPY vite.config.ts svelte.config.js tsconfig.json ./

RUN npm run build

#
# Production
#
FROM node:24-alpine AS runner

COPY --from=server /server/watcharr /
COPY --from=ui /app/build /ui
COPY --from=ui /app/package.json /app/package-lock.json /ui/

# Install just the prod dependencies for final step.
# We --ignore-scripts, to stop the `prepare` script from auto-
# running, which is only meant for development (will error).
RUN cd /ui && npm ci --omit=dev --ignore-scripts=true

EXPOSE 3080

CMD ["/watcharr"]

