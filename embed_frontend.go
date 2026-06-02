package main

import "embed"

// frontendDist 嵌入前端构建产物（web/dist 目录）。
// 构建顺序：先 pnpm build（产出 web/dist），再 go build。
//
//go:embed all:web/dist
var frontendDist embed.FS
