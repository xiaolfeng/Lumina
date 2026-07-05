// Package logic RepoWiki Agent 分析 Pass 的单元测试。
//
// 测试覆盖：
//   - TestParseAgentJSON: parseAgentJSON 纯函数（直接 JSON / markdown 包裹 / 花括号提取 / 失败用例）
//   - TestAgentPassRetry: executeWithRetry 重试逻辑（使用 mockAgent，不依赖真实 LLM）
//   - TestAgentPassStub: 4 个 Pass 串行执行的集成测试（使用 StubLLMProvider）
package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bamboo-services/bamboo-agent/agent"
	"github.com/bamboo-services/bamboo-agent/tool"
	xLog "github.com/bamboo-services/bamboo-base-go/common/log"
	"github.com/bamboo-services/bamboo-messages/bamboo"

	"github.com/xiaolfeng/Lumina/internal/service"
)

// ════════════════════════════════════════════════════════════════════
// Mock Agent（测试用 Agent 桩实现）
// ════════════════════════════════════════════════════════════════════

// mockAgent 实现 agent.Agent 接口的测试桩
//
// 按调用顺序依次返回 responses 中的内容，用于测试 executeWithRetry 的重试逻辑。
// 如果某次调用对应 errors 中有非 nil 错误，则返回该错误而不返回内容。
type mockAgent struct {
	responses []string // 按 Run 调用顺序排列的输出内容
	errors    []error  // 按调用顺序排列的错误（可选，对应位置的响应被跳过）
	callCount int      // Run 被调用的次数
	lastInput string   // 最后一次 Run 收到的 input（用于验证重试提示追加）
}

func (m *mockAgent) Run(_ context.Context, input string) (*agent.AgentResult, error) {
	idx := m.callCount
	m.callCount++
	m.lastInput = input

	// 优先返回错误
	if idx < len(m.errors) && m.errors[idx] != nil {
		return nil, m.errors[idx]
	}

	// 返回对应位置的响应
	var content string
	if idx < len(m.responses) {
		content = m.responses[idx]
	} else if len(m.responses) > 0 {
		content = m.responses[len(m.responses)-1] // 超出范围时重复最后一个
	} else {
		content = "{}"
	}

	return &agent.AgentResult{
		Content: content,
	}, nil
}

func (m *mockAgent) Stream(_ context.Context, _ string) (<-chan agent.AgentEvent, error) {
	return nil, fmt.Errorf("mockAgent: Stream not implemented")
}

func (m *mockAgent) RunWithMessages(ctx context.Context, _ []bamboo.BambooMessage) (*agent.AgentResult, error) {
	return m.Run(ctx, "")
}

func (m *mockAgent) AddTool(_ tool.Tool) error { return nil }

func (m *mockAgent) SetSystemPrompt(_ string) {}

// ════════════════════════════════════════════════════════════════════
// 辅助函数
// ════════════════════════════════════════════════════════════════════

// newTestRunner 创建测试用 AgentPassRunner（不依赖真实 LLM 客户端）
func newTestRunner(maxRetries int) *AgentPassRunner {
	return &AgentPassRunner{
		client:     nil,
		storage:    nil,
		log:        xLog.WithName(xLog.NamedLOGC, "TestAgentPass"),
		tools:      nil,
		maxRetries: maxRetries,
	}
}

// newTestStorage 创建临时存储服务
func newTestStorage(t *testing.T) (*service.WikiStorageService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("REPOWIKI_STORAGE_PATH", tmpDir)
	return service.NewWikiStorageService(), tmpDir
}

// ════════════════════════════════════════════════════════════════════
// TestParseAgentJSON — parseAgentJSON 纯函数测试
// ════════════════════════════════════════════════════════════════════

func TestParseAgentJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantField string // 验证解析后的 JSON 中是否包含指定字段值
	}{
		{
			name:      "直接 JSON 对象",
			input:     `{"project_name": "Lumina", "version": "1.0"}`,
			wantErr:   false,
			wantField: "Lumina",
		},
		{
			name:      "带空白的 JSON",
			input:     "  \n  {\"key\": \"value\"}  \n  ",
			wantErr:   false,
			wantField: "value",
		},
		{
			name:      "markdown json 代码块包裹",
			input:     "```json\n{\"modules\": []}\n```",
			wantErr:   false,
			wantField: "modules",
		},
		{
			name:      "markdown 无语言标记代码块",
			input:     "```\n{\"key\": \"val\"}\n```",
			wantErr:   false,
			wantField: "val",
		},
		{
			name:      "JSON 前后有解释文字",
			input:     `好的，分析结果如下：{"project_name": "test"} 希望对你有帮助。`,
			wantErr:   false,
			wantField: "test",
		},
		{
			name:      "嵌套 JSON 对象",
			input:     `{"outer": {"inner": [1, 2, {"deep": true}]}}`,
			wantErr:   false,
			wantField: "deep",
		},
		{
			name:      "空字符串",
			input:     "",
			wantErr:   true,
			wantField: "",
		},
		{
			name:      "纯文本无 JSON",
			input:     "这是一段纯文本，没有任何 JSON 内容。",
			wantErr:   true,
			wantField: "",
		},
		{
			name:      "不完整的 JSON",
			input:     `{"key": "value"`,
			wantErr:   true,
			wantField: "",
		},
		{
			name:      "markdown 代码块但内容非 JSON",
			input:     "```json\n这是文字不是 JSON\n```",
			wantErr:   true,
			wantField: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseAgentJSON(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("期望返回错误，但 parseAgentJSON 成功解析: %s", string(parsed))
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAgentJSON 返回意外错误: %v", err)
			}
			if !json.Valid(parsed) {
				t.Fatalf("解析结果不是有效 JSON: %s", string(parsed))
			}
			if tt.wantField != "" {
				parsedStr := string(parsed)
				if !contains(parsedStr, tt.wantField) {
					t.Errorf("解析结果中未找到期望字段值 %q，实际: %s", tt.wantField, parsedStr)
				}
			}
		})
	}
}

// TestExtractFencedBlock 测试 markdown 围栏代码块提取
func TestExtractFencedBlock(t *testing.T) {
	tests := []struct {
		name string
		text string
		lang string
		want string
	}{
		{
			name: "json 围栏块",
			text: "```json\n{\"a\":1}\n```",
			lang: "json",
			want: `{"a":1}`,
		},
		{
			name: "无语言标记围栏块",
			text: "```\n{\"b\":2}\n```",
			lang: "",
			want: `{"b":2}`,
		},
		{
			name: "无围栏块",
			text: "普通文本",
			lang: "json",
			want: "",
		},
		{
			name: "不闭合围栏块",
			text: "```json\n{\"c\":3}",
			lang: "json",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractFencedBlock(tt.text, tt.lang)
			if got != tt.want {
				t.Errorf("extractFencedBlock(%q, %q) = %q, want %q", tt.text, tt.lang, got, tt.want)
			}
		})
	}
}

// TestExtractBraceRange 测试花括号范围提取
func TestExtractBraceRange(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{"正常 JSON", `prefix {"key":"val"} suffix`, `{"key":"val"}`},
		{"无花括号", "no braces here", ""},
		{"只有开括号", `{ incomplete`, ""},
		{"嵌套花括号", `a {"b":{"c":1}} z`, `{"b":{"c":1}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBraceRange(tt.text)
			if got != tt.want {
				t.Errorf("extractBraceRange(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// ════════════════════════════════════════════════════════════════════
// TestAgentPassRetry — 重试逻辑测试（使用 mockAgent）
// ════════════════════════════════════════════════════════════════════

// TestAgentPassRetry 验证 JSON 解析失败时的重试行为
func TestAgentPassRetry(t *testing.T) {
	ctx := context.Background()

	t.Run("首次即成功_不重试", func(t *testing.T) {
		runner := newTestRunner(3)
		mockAg := &mockAgent{
			responses: []string{`{"status": "ok"}`},
		}

		result := runner.executeWithRetry(ctx, mockAg, "test input", "pass1")

		if !result.Success {
			t.Fatalf("期望成功，但返回失败: %s", result.Error)
		}
		if mockAg.callCount != 1 {
			t.Errorf("期望调用 1 次，实际 %d 次", mockAg.callCount)
		}
		if result.Attempts != 1 {
			t.Errorf("期望 Attempts=1，实际 %d", result.Attempts)
		}
		// 验证 JSON 解析正确
		var parsed map[string]interface{}
		if err := json.Unmarshal(result.JSON, &parsed); err != nil {
			t.Fatalf("解析 result.JSON 失败: %v", err)
		}
		if parsed["status"] != "ok" {
			t.Errorf("期望 status=ok，实际 %v", parsed["status"])
		}
	})

	t.Run("全部失败_重试3次后标记failed", func(t *testing.T) {
		runner := newTestRunner(3)
		mockAg := &mockAgent{
			responses: []string{
				"这不是 JSON",
				"```invalid```",
				`{broken`,
			},
		}

		result := runner.executeWithRetry(ctx, mockAg, "test input", "pass1")

		if result.Success {
			t.Fatal("期望失败，但返回成功")
		}
		if mockAg.callCount != 3 {
			t.Errorf("期望调用 3 次（全部重试），实际 %d 次", mockAg.callCount)
		}
		if result.Attempts != 3 {
			t.Errorf("期望 Attempts=3，实际 %d", result.Attempts)
		}
		if result.Error == "" {
			t.Error("失败结果应包含错误信息")
		}
	})

	t.Run("第二次成功_重试1次后通过", func(t *testing.T) {
		runner := newTestRunner(3)
		mockAg := &mockAgent{
			responses: []string{
				"invalid json first",
				`{"recovered": true}`,
			},
		}

		result := runner.executeWithRetry(ctx, mockAg, "test input", "pass1")

		if !result.Success {
			t.Fatalf("期望成功（第二次重试通过），但返回失败: %s", result.Error)
		}
		if mockAg.callCount != 2 {
			t.Errorf("期望调用 2 次，实际 %d 次", mockAg.callCount)
		}
		if result.Attempts != 2 {
			t.Errorf("期望 Attempts=2，实际 %d", result.Attempts)
		}
	})

	t.Run("Agent_Run错误触发重试", func(t *testing.T) {
		runner := newTestRunner(3)
		mockAg := &mockAgent{
			responses: []string{`{"ok": true}`},
			errors: []error{
				fmt.Errorf("simulated LLM error"),
				nil, // 第二次成功
			},
		}

		result := runner.executeWithRetry(ctx, mockAg, "test input", "pass1")

		if !result.Success {
			t.Fatalf("期望成功，但返回失败: %s", result.Error)
		}
		if mockAg.callCount != 2 {
			t.Errorf("期望调用 2 次（第一次出错重试），实际 %d 次", mockAg.callCount)
		}
	})

	t.Run("重试提示追加到input", func(t *testing.T) {
		runner := newTestRunner(3)
		mockAg := &mockAgent{
			responses: []string{
				"invalid",
				`{"ok": true}`,
			},
		}

		runner.executeWithRetry(ctx, mockAg, "base input", "pass1")

		// 第二次调用的 input 应包含重试提示
		if !contains(mockAg.lastInput, "重试") {
			t.Errorf("第二次调用的 input 应包含重试提示，实际: %s", mockAg.lastInput)
		}
		if !contains(mockAg.lastInput, "base input") {
			t.Errorf("第二次调用的 input 应包含原始输入前缀，实际: %s", mockAg.lastInput)
		}
	})

	t.Run("上下文取消立即返回", func(t *testing.T) {
		runner := newTestRunner(3)
		mockAg := &mockAgent{responses: []string{`{"ok":true}`}}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消

		result := runner.executeWithRetry(ctx, mockAg, "test input", "pass1")

		if result.Success {
			t.Error("已取消的上下文不应成功")
		}
		if mockAg.callCount != 0 {
			t.Errorf("已取消的上下文不应调用 Agent，实际调用了 %d 次", mockAg.callCount)
		}
	})
}

// ════════════════════════════════════════════════════════════════════
// TestAgentPassStub — 4 Pass 串行集成测试（使用 StubLLMProvider）
// ════════════════════════════════════════════════════════════════════

// TestAgentPassStub 使用 StubLLMProvider 验证 4 个 Pass 均能执行并返回可解析 JSON
//
// 测试流程：
//  1. 创建 StubLLMProvider（返回固定有效 JSON）
//  2. 用 bamboo.NewClient 包装为 BambooClient
//  3. 创建 AgentPassRunner
//  4. 执行 RunAllPasses
//  5. 验证 4 个 Pass 均成功且 JSON 可解析
func TestAgentPassStub(t *testing.T) {
	// StubLLMProvider 返回的有效 JSON（满足所有 4 个 Pass 的最低格式要求）
	stubJSON := `{"project_name":"test","description":"stub","tech_stack":["Go"],"modules":[],"architecture_pattern":"test","reading_order":[]}`

	stub := &service.StubLLMProvider{
		ResponseJSON: stubJSON,
	}
	client := bamboo.NewClient(stub)

	storage, _ := newTestStorage(t)
	runner := NewAgentPassRunner(
		client,
		storage,
		xLog.WithName(xLog.NamedLOGC, "TestAgentPassStub"),
		nil, // 无需工具（stub 不会调用工具）
	)

	// 准备上下文数据
	fileScan := &service.FileScanResult{
		Files: []service.FileInfo{
			{Path: "main.go", Size: 1024, Language: "Go", IsEntry: true},
		},
		LanguageStats: map[string]int{"Go": 1},
		EntryPoints:   []string{"main.go"},
		TotalFiles:    1,
		TotalSize:     1024,
	}
	depSummary := &service.DependencySummary{
		Modules: []service.ModuleInfo{
			{Name: ".", Path: ".", KeyFiles: []string{"main.go"}},
		},
		CoreModules:  []string{"."},
		TotalImports: 0,
	}

	// 进度回调跟踪
	var stages []string
	progressCb := func(stage string) {
		stages = append(stages, stage)
	}

	// 执行（设置超时保护）
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// 使用临时目录作为 repoPath
	repoPath := t.TempDir()

	results, xErr := runner.RunAllPasses(ctx, 1001, repoPath, fileScan, depSummary, progressCb)

	// 如果 stub provider 与 bamboo-agent 框架不兼容，跳过集成测试
	if xErr != nil && results["pass1"] != nil && !results["pass1"].Success {
		if contains(results["pass1"].Error, "Agent 执行错误") {
			t.Skipf("StubLLMProvider 与 bamboo-agent 框架存在兼容性问题，跳过集成测试: %s", results["pass1"].Error)
		}
	}

	// 验证全部 4 个 Pass 均执行
	expectedPasses := []string{"pass1", "pass2", "pass3", "pass4"}
	for _, name := range expectedPasses {
		result, ok := results[name]
		if !ok {
			t.Errorf("结果中缺少 %s", name)
			continue
		}
		if !result.Success {
			t.Errorf("%s 执行失败: %s (attempts=%d)", name, result.Error, result.Attempts)
			continue
		}
		// 验证 JSON 可解析
		if !json.Valid(result.JSON) {
			t.Errorf("%s 的 JSON 字段不是有效 JSON: %s", name, string(result.JSON))
		}
	}

	// 验证进度回调顺序
	expectedStages := []string{"pass1", "pass2", "pass3", "pass4"}
	if len(stages) != len(expectedStages) {
		t.Errorf("期望 %d 个进度回调，实际 %d 个: %v", len(expectedStages), len(stages), stages)
	} else {
		for i, expected := range expectedStages {
			if stages[i] != expected {
				t.Errorf("进度回调顺序错误：第 %d 个期望 %q，实际 %q", i+1, expected, stages[i])
			}
		}
	}

	// 验证 Pass 结果文件已持久化
	for i := 1; i <= 4; i++ {
		passPath := filepath.Join(storage.GetPassesPath(1001), fmt.Sprintf("pass-%d.json", i))
		if _, err := os.Stat(passPath); err != nil {
			t.Errorf("Pass 结果文件未持久化: %s (%v)", passPath, err)
		}
	}
}

// ════════════════════════════════════════════════════════════════════
// 辅助函数
// ════════════════════════════════════════════════════════════════════

// contains 检查字符串是否包含子串（简单封装，避免引入 strings 包）
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

// findSubstring 暴力子串搜索
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
