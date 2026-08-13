# syntax=docker/dockerfile:1
# --------------------------------------------------------------------------------
# Copyright (c) 2016-NOW(至今) 筱锋
# Author: 筱锋「xiao_lfeng」(https://www.x-lf.com)
# --------------------------------------------------------------------------------
# 许可证声明：版权所有 (c) 2016-2026 筱锋。保留所有权利。
# 有关MIT许可证的更多信息，请查看项目根目录下的LICENSE文件或访问：
# https://opensource.org/licenses/MIT
# --------------------------------------------------------------------------------

# =============================================================================
# Lumina · 微明 多阶段构建
#   Stage 1  frontend-builder : Node 22 + pnpm → 双前端 dist（resources/web + web-wiki）
#   Stage 2  go-builder       : golang 1.25 alpine → 单二进制（CGO_ENABLED=0）
#   Stage 3  runtime          : alpine + CA 证书 + tini（优雅退出） + tzdata
# =============================================================================

# ---------- Stage 1: 前端构建（pnpm workspace 双前端） ----------
FROM node:22-alpine AS frontend-builder
WORKDIR /build

# 仅复制依赖清单，最大化层缓存；workspace 根无 package.json，必须逐个包 COPY manifest
# pnpm 版本与本地开发对齐（11.9.0），allowBuilds(unrs-resolver) 依赖 pnpm>=11
COPY pnpm-workspace.yaml pnpm-lock.yaml ./
COPY web/package.json web/package.json
COPY web-wiki/package.json web-wiki/package.json
COPY components/package.json components/package.json
RUN npm install -g pnpm@11.9.0 && pnpm install --frozen-lockfile

# 复制前端源码（@lumina/components exports 直接指向 src 源码，vite 打包即包含，无需单独 build）
COPY web/ web/
COPY web-wiki/ web-wiki/
COPY components/ components/

# 双前端构建；vite outDir 相对各包目录，分别产出到 /build/resources/web/dist 与 /build/resources/web-wiki/dist
RUN pnpm --filter web build && pnpm --filter web-wiki build

# ---------- Stage 2: Go 编译 ----------
FROM golang:1.25-alpine AS go-builder
WORKDIR /src

# 复制依赖清单并预下载（层缓存友好；本项目无本地路径 replace，直接可用）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码（.dockerignore 已排除本地 dist / node_modules / .lumina 等；docs/ 保留，route_swagger 编译依赖）
COPY . .

# 用前端阶段产物覆盖，确保 go:embed all:web/dist 与 all:web-wiki/dist 内嵌最新产物
COPY --from=frontend-builder /build/resources/web/dist /src/resources/web/dist
COPY --from=frontend-builder /build/resources/web-wiki/dist /src/resources/web-wiki/dist

# CGO_ENABLED=0 静态编译（pgx / sonyflake / go-git 等均纯 Go，不依赖 cgo）
# GOOS/GOARCH 不写死，随 golang 镜像目标架构（amd64/arm64）自动匹配，配合 buildx 多平台构建
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lumina .

# ---------- Stage 3: 运行镜像 ----------
FROM alpine:3.21 AS runtime

# CA 证书（HTTPS 克隆 / LLM API 调用）+ 优雅退出 init + 时区数据（系统时间 Asia/Shanghai）
RUN apk add --no-cache \
    ca-certificates \
    tini \
    tzdata

# RepoWiki 存储 / Q&A 缓存为相对进程工作目录的路径，固定 WORKDIR=/app
ENV TZ=Asia/Shanghai

WORKDIR /app

COPY --from=go-builder /out/lumina /app/lumina
RUN mkdir -p /app/.lumina/repowiki /app/.lumina/cache

EXPOSE 8080

# tini 作为 PID 1 转发 SIGTERM/SIGINT，保证框架 Runner 优雅退出
ENTRYPOINT ["/sbin/tini", "--"]
CMD ["/app/lumina"]
