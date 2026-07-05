package main

import "embed"

// wikiFrontendDist 嵌入 Wiki Reader 前端构建产物（web-wiki/dist 目录）。
// 构建顺序：先 pnpm build（产出 web-wiki/dist），再 go build。
//
//go:embed all:web-wiki/dist
var wikiFrontendDist embed.FS
