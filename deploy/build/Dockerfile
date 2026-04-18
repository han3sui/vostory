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
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/server ./cmd/server
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/migration ./cmd/migration
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/syncmenu ./cmd/syncmenu
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o ./bin/createadmin ./cmd/createadmin

RUN cp -r config /build/server/bin/ && \
    (cp -r docs /build/server/bin/ || echo "docs directory not found, skipping...")

# ============================================================
# Stage 3: Download ffmpeg static build
# ============================================================
FROM alpine:3.20 AS ffmpeg-downloader

RUN apk add --no-cache curl xz
ARG FFMPEG_URL=https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz
RUN curl -fsSL -o /tmp/ffmpeg.tar.xz "$FFMPEG_URL" && \
    mkdir -p /tmp/ffmpeg && \
    tar -xJf /tmp/ffmpeg.tar.xz -C /tmp/ffmpeg --strip-components=2 && \
    cp /tmp/ffmpeg/ffmpeg /usr/local/bin/ffmpeg && \
    chmod +x /usr/local/bin/ffmpeg && \
    rm -rf /tmp/ffmpeg /tmp/ffmpeg.tar.xz

# ============================================================
# Stage 4: Final image (nginx + go server)
# ============================================================
FROM alpine:3.20

RUN apk add --no-cache nginx supervisor tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata && \
    mkdir -p /run/nginx /var/log/supervisor /app/storage/logs

COPY --from=ffmpeg-downloader /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg

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
