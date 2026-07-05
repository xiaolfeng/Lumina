package service

import (
	"os"
	"path/filepath"
	"testing"
)

// ── 临时目录管理 ──

// RepoWikiTestHelper RepoWiki 测试辅助工具
type RepoWikiTestHelper struct {
	TempDir string // 临时目录路径
}

// NewRepoWikiTestHelper 创建测试辅助工具，初始化临时目录
func NewRepoWikiTestHelper(t *testing.T) *RepoWikiTestHelper {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "repowiki-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return &RepoWikiTestHelper{TempDir: tmpDir}
}

// CreateSampleRepo 在临时目录中创建模拟仓库结构
//
// 创建一个包含多语言文件的模拟项目结构：
// ├── main.go (Go 入口)
// ├── go.mod
// ├── internal/
// │   ├── handler/
// │   │   └── user.go
// │   └── logic/
// │       └── user.go
// ├── src/
// │   ├── index.ts (TS 入口)
// │   └── utils.ts
// ├── package.json
// ├── .git/ (空目录，用于测试过滤)
// └── node_modules/ (空目录，用于测试过滤)
func (h *RepoWikiTestHelper) CreateSampleRepo() string {
	repoPath := filepath.Join(h.TempDir, "sample-repo")
	dirs := []string{
		".git/objects",
		"node_modules/some-pkg",
		"internal/handler",
		"internal/logic",
		"src/components",
		"vendor/lib",
		"dist",
	}
	for _, d := range dirs {
		os.MkdirAll(filepath.Join(repoPath, d), 0755)
	}

	files := map[string]string{
		"main.go":                        SampleGoMain,
		"go.mod":                         SampleGoMod,
		"internal/handler/user.go":       SampleGoHandler,
		"internal/logic/user.go":         SampleGoLogic,
		"src/index.ts":                   SampleTSIndex,
		"src/utils.ts":                   SampleTSUtils,
		"package.json":                   SamplePackageJSON,
		".gitignore":                     SampleGitignore,
		"README.md":                      SampleREADME,
		"node_modules/some-pkg/index.js": "// should be excluded",
		"vendor/lib/lib.go":              "// should be excluded",
		"dist/bundle.js":                 "// should be excluded",
	}
	for path, content := range files {
		fullPath := filepath.Join(repoPath, path)
		os.MkdirAll(filepath.Dir(fullPath), 0755)
		os.WriteFile(fullPath, []byte(content), 0644)
	}
	return repoPath
}

// ── Sample 文件内容常量 ──

const SampleGoMain = `package main

import (
	"fmt"
	"example.com/sample/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	handler.RegisterRoutes(r)
	fmt.Println("Server starting...")
	r.Run(":8080")
}
`

const SampleGoMod = `module example.com/sample

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)
`

const SampleGoHandler = `package handler

import (
	"github.com/gin-gonic/gin"
	"example.com/sample/internal/logic"
)

type UserHandler struct {
	logic *logic.UserLogic
}

func RegisterRoutes(r *gin.Engine) {
	h := &UserHandler{}
	r.GET("/users/:id", h.GetUser)
}
`

const SampleGoLogic = `package logic

import (
	"context"
	"fmt"
)

type UserLogic struct{}

func (l *UserLogic) GetUser(ctx context.Context, id int64) (string, error) {
	return fmt.Sprintf("User %d", id), nil
}
`

const SampleTSIndex = `import { UserService } from './utils';
import express from 'express';

const app = express();
const userService = new UserService();

app.get('/users/:id', (req, res) => {
    res.json(userService.getUser(req.params.id));
});

app.listen(3000);
`

const SampleTSUtils = `import axios from 'axios';

export class UserService {
    private api = axios.create({ baseURL: '/api' });

    async getUser(id: string) {
        const res = await this.api.get('/users/' + id);
        return res.data;
    }
}

export function formatDate(d: Date): string {
    return d.toISOString();
}
`

const SamplePackageJSON = `{
  "name": "sample-project",
  "version": "1.0.0",
  "main": "src/index.ts",
  "dependencies": {
    "express": "^4.18.0",
    "axios": "^1.6.0"
  }
}
`

const SampleGitignore = `node_modules/
dist/
*.log
.env
`

const SampleREADME = `# Sample Project

This is a sample project for testing RepoWiki file scanner.
`

// ── Sample Import 语句（用于依赖提取测试）──

// SampleGoImports 包含多种 Go import 格式的测试用例
var SampleGoImports = []string{
	`import "fmt"`,
	`import (
	"fmt"
	"strings"
)`,
	`import (
	"context"
	
	"github.com/gin-gonic/gin"
	"example.com/sample/internal/logic"
)`,
}

// SamplePythonImports 包含多种 Python import 格式的测试用例
var SamplePythonImports = []string{
	`import os`,
	`import sys`,
	`from flask import Flask, request`,
	`from pathlib import Path`,
}

// SampleJSImports 包含多种 JS/TS import 格式的测试用例
var SampleJSImports = []string{
	`import express from 'express'`,
	`import { useState } from 'react'`,
	`const axios = require('axios')`,
	`import type { User } from './types'`,
}

// SampleJavaImports 包含 Java import 格式的测试用例
var SampleJavaImports = []string{
	`import java.util.List;`,
	`import com.example.service.UserService;`,
}

// SampleRustImports 包含 Rust use 格式的测试用例
var SampleRustImports = []string{
	`use std::collections::HashMap;`,
	`use serde::{Serialize, Deserialize};`,
}
