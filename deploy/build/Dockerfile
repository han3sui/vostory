# ============================================================
# Stage 1: Build frontend
# ============================================================
FROM node:22-alpine AS frontend-builder

RUN corepack enable && corepack prepare pnpm@10 --activate

WORKDIR /build/web
COPY vostory-web/package.json vostory-web/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY vostory-web/ ./
RUN pnpm run build

# ============================================================
# Stage 2: Build backend
# ============================================================
FROM golang:1.24-alpine AS backend-builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build/server
COPY vostory-server/go.mod vostory-server/go.sum ./

RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY vostory-server/ ./

ARG VERSION=dev
RUN echo "Building server..." && \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/server ./cmd/server && \
    echo "Building migration..." && \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/migration ./cmd/migration && \
    echo "Building syncmenu..." && \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/syncmenu ./cmd/syncmenu && \
    echo "Building createadmin..." && \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/createadmin ./cmd/createadmin && \
    echo "Building task..." && \
    CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/task ./cmd/task && \
    echo "Build completed!"

RUN cp -r config /build/server/bin/ && \
    (cp -r docs /build/server/bin/ || echo "docs directory not found, skipping...")

# ============================================================
# Stage 3: Final image (nginx + go server)
# ============================================================
FROM alpine:3.20

RUN apk add --no-cache nginx supervisor tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata && \
    mkdir -p /run/nginx /var/log/supervisor /app/storage/logs

ARG VERSION=dev
ENV APP_VERSION=${VERSION}

COPY deploy/build/nginx.conf /etc/nginx/http.d/default.conf
COPY deploy/build/supervisord.conf /etc/supervisord.conf
COPY deploy/build/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

COPY --from=frontend-builder /build/web/dist /usr/share/nginx/html
COPY --from=backend-builder /build/server/bin /app

WORKDIR /app

EXPOSE 80

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["-conf", "config/docker.yml"]
