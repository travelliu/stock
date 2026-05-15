FROM node:22-alpine AS deps

RUN corepack enable && corepack prepare pnpm@10.28.2 --activate

WORKDIR /app

# Copy workspace config and all package.json files for dependency resolution
COPY web/pnpm-lock.yaml web/package.json ./

RUN pnpm install --frozen-lockfile

# --- Build ---
FROM node:22-alpine AS builder_web

RUN corepack enable && corepack prepare pnpm@10.28.2 --activate

WORKDIR /app

# Copy installed dependencies (preserves pnpm symlink structure)
COPY --from=deps /app ./

# Copy source
COPY web/ ./

# Re-link after source overlay (fixes any symlinks overwritten by COPY)
RUN pnpm build

FROM golang:1.26-alpine AS builder_backend
RUN apk add --no-cache git make

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct

# Copy server source
COPY . .
# Cache dependencies
COPY --from=builder_web /app/dist ./web/

RUN make build


# --- Runtime stage ---
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder_backend /src/bin/stockd .
COPY --from=builder_backend /src/bin/stockctl .
COPY docker/entrypoint.sh .
copy config.example.yaml config.yaml
RUN sed -i 's/\r$//' entrypoint.sh && chmod +x entrypoint.sh

EXPOSE 8443

ENTRYPOINT ["./entrypoint.sh"]