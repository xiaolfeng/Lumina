package service

import (
	"os"
	"path/filepath"
	"testing"
)

// ── Sample 文件内容常量（依赖提取测试夹具）──

const sampleGoMain = `package main

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

const sampleGoMod = `module example.com/sample

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
)
`

const sampleGoHandler = `package handler

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

const sampleGoLogic = `package logic

import (
	"context"
	"fmt"
)

type UserLogic struct{}

func (l *UserLogic) GetUser(ctx context.Context, id int64) (string, error) {
	return fmt.Sprintf("User %d", id), nil
}
`

const sampleTSIndex = `import { UserService } from './utils';
import express from 'express';

const app = express();
const userService = new UserService();

app.get('/users/:id', (req, res) => {
    res.json(userService.getUser(req.params.id));
});

app.listen(3000);
`

const sampleTSUtils = `import axios from 'axios';

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

const samplePackageJSON = `{
  "name": "sample-project",
  "version": "1.0.0",
  "main": "src/index.ts",
  "dependencies": {
    "express": "^4.18.0",
    "axios": "^1.6.0"
  }
}
`

const sampleGitignore = `node_modules/
dist/
*.log
.env
`

const sampleREADME = `# Sample Project

This is a sample project for testing RepoWiki file scanner.
`

// ── Sample Import 语句（用于依赖提取单测）──

var sampleGoImports = []string{
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

var samplePythonImports = []string{
	`import os`,
	`import sys`,
	`from flask import Flask, request`,
	`from pathlib import Path`,
}

var sampleJSImports = []string{
	`import express from 'express'`,
	`import { useState } from 'react'`,
	`const axios = require('axios')`,
	`import type { User } from './types'`,
}

var sampleJavaImports = []string{
	`import java.util.List;`,
	`import com.example.service.UserService;`,
}

var sampleRustImports = []string{
	`use std::collections::HashMap;`,
	`use serde::{Serialize, Deserialize};`,
}

// createSampleRepo 在 testing.T 内置临时目录中创建模拟仓库结构
//
// 仓库布局：
//
//	├── main.go (Go 入口)
//	├── go.mod
//	├── internal/
//	│   ├── handler/
//	│   │   └── user.go
//	│   └── logic/
//	│       └── user.go
//	├── src/
//	│   ├── index.ts (TS 入口)
//	│   └── utils.ts
//	├── package.json
//	├── .git/ (空目录，用于测试过滤)
//	└── node_modules/ (空目录，用于测试过滤)
func createSampleRepo(t *testing.T) string {
	t.Helper()
	repoPath := filepath.Join(t.TempDir(), "sample-repo")

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
		if err := os.MkdirAll(filepath.Join(repoPath, d), 0755); err != nil {
			t.Fatalf("创建目录失败 %s: %v", d, err)
		}
	}

	files := map[string]string{
		"main.go":                        sampleGoMain,
		"go.mod":                         sampleGoMod,
		"internal/handler/user.go":       sampleGoHandler,
		"internal/logic/user.go":         sampleGoLogic,
		"src/index.ts":                   sampleTSIndex,
		"src/utils.ts":                   sampleTSUtils,
		"package.json":                   samplePackageJSON,
		".gitignore":                     sampleGitignore,
		"README.md":                      sampleREADME,
		"node_modules/some-pkg/index.js": "// should be excluded",
		"vendor/lib/lib.go":              "// should be excluded",
		"dist/bundle.js":                 "// should be excluded",
	}
	for path, content := range files {
		fullPath := filepath.Join(repoPath, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("创建父目录失败 %s: %v", fullPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("写入文件失败 %s: %v", fullPath, err)
		}
	}
	return repoPath
}
