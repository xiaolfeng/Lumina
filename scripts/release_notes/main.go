package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://ai-intl.x-lf.com/v1"
	defaultModel   = "glm-5.3-flash"
	defaultRepo    = "xiaolfeng/Lumina"
	maxDiffLength  = 20000
)

type gitContext struct {
	Repo        string
	CurrentTag  string
	PreviousTag string
	Range       string
	Commits     []string
	DiffStat    string
	DiffSnippet string
}

type openAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []openAIChatMessage `json:"messages"`
}

type openAIChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func main() {
	var (
		tagFlag     = flag.String("tag", "", "当前发版的 Git Tag（如 v1.0.1-beta.1）")
		prevTagFlag = flag.String("previous-tag", "", "上一版本 Git Tag（可选，默认自动检测）")
		repoFlag    = flag.String("repo", "", "GitHub 仓库（格式 owner/repo，默认检测或 xiaolfeng/Lumina）")
		outFlag     = flag.String("out", "", "输出文件路径（可选，默认打印到 stdout）")
		dryRunFlag  = flag.Bool("dry-run", false, "仅提取 Git 上下文并不调用 AI")
	)
	flag.Parse()

	tag := strings.TrimSpace(*tagFlag)
	if tag == "" {
		tag = os.Getenv("TAG")
	}
	if tag == "" {
		tag = os.Getenv("GITHUB_REF_NAME")
	}
	if tag == "" {
		fmt.Fprintf(os.Stderr, "错误: 必须指定 -tag 参数或设置 TAG 环境变量\n")
		os.Exit(1)
	}

	repo := strings.TrimSpace(*repoFlag)
	if repo == "" {
		repo = os.Getenv("GITHUB_REPOSITORY")
	}
	if repo == "" {
		repo = defaultRepo
	}

	gitCtx, err := collectGitContext(tag, strings.TrimSpace(*prevTagFlag), repo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "提取 Git 上下文失败: %v\n", err)
		os.Exit(1)
	}

	if *dryRunFlag {
		fmt.Printf("Range: %s\nCommits: %d\n", gitCtx.Range, len(gitCtx.Commits))
		fmt.Println("DiffStat:\n" + gitCtx.DiffStat)
		return
	}

	baseURL := strings.TrimRight(getEnvOrDefault("AI_BASE_URL", defaultBaseURL), "/")
	model := getEnvOrDefault("AI_MODEL", defaultModel)
	apiKey := os.Getenv("AI_API_KEY")

	var releaseNotes string
	if apiKey != "" {
		notes, aiErr := generateWithAI(gitCtx, baseURL, model, apiKey)
		if aiErr != nil {
			fmt.Fprintf(os.Stderr, "AI 生成发版说明失败 (%v)，回退到内置模板渲染\n", aiErr)
			releaseNotes = renderFallbackTemplate(gitCtx)
		} else {
			releaseNotes = strings.TrimSpace(notes)
		}
	} else {
		fmt.Fprintf(os.Stderr, "未配置 AI_API_KEY，使用结构化模板本地渲染\n")
		releaseNotes = renderFallbackTemplate(gitCtx)
	}

	// 确保末尾有且仅有一个换行符
	releaseNotes = formatReleaseNotes(releaseNotes, gitCtx) + "\n"

	if *outFlag != "" {
		if err := os.MkdirAll(filepath.Dir(*outFlag), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "创建输出目录失败: %v\n", err)
			os.Exit(1)
		}
		if err := os.WriteFile(*outFlag, []byte(releaseNotes), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "写入文件失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "发版说明已成功写入: %s\n", *outFlag)
	} else {
		fmt.Print(releaseNotes)
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := strings.TrimSpace(os.Getenv(key)); val != "" {
		return val
	}
	return defaultVal
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func collectGitContext(currentTag, explicitPrevTag, repo string) (*gitContext, error) {
	prevTag := explicitPrevTag
	if prevTag == "" {
		// 优先尝试从 git 历史中找祖先 tag
		if out, err := runGit("describe", "--tags", "--abbrev=0", currentTag+"^"); err == nil && out != "" {
			prevTag = strings.TrimSpace(out)
		} else {
			// 回退到语义版本列表检索
			tagsOut, err := runGit("tag", "--list", "v*", "--sort=-v:refname")
			if err == nil && tagsOut != "" {
				lines := strings.Split(tagsOut, "\n")
				for i, t := range lines {
					if strings.TrimSpace(t) == currentTag && i+1 < len(lines) {
						prevTag = strings.TrimSpace(lines[i+1])
						break
					}
				}
			}
		}
	}

	rangeStr := currentTag
	if prevTag != "" && prevTag != currentTag {
		rangeStr = prevTag + ".." + currentTag
	}

	// 1. 获取 Commit 列表
	logOut, _ := runGit("log", "--no-merges", "--pretty=format:%h %s (%an)", rangeStr)
	var commits []string
	if logOut != "" {
		for _, line := range strings.Split(logOut, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" {
				commits = append(commits, trimmed)
			}
		}
	}

	// 2. 获取 Diff Stat
	diffStat, _ := runGit("diff", "--stat", rangeStr)

	// 3. 获取核心代码 Diff 概览（排除锁文件、静态产物）
	diffArgs := []string{
		"diff", rangeStr, "--",
		".",
		":(exclude)pnpm-lock.yaml",
		":(exclude)*.lock",
		":(exclude)go.sum",
		":(exclude)resources/web/dist",
		":(exclude)resources/web-wiki/dist",
		":(exclude)docs/swagger.json",
		":(exclude)docs/swagger.yaml",
	}
	diffSnippet, _ := runGit(diffArgs...)
	if len(diffSnippet) > maxDiffLength {
		diffSnippet = diffSnippet[:maxDiffLength] + "\n... [diff truncated for length] ..."
	}

	return &gitContext{
		Repo:        repo,
		CurrentTag:  currentTag,
		PreviousTag: prevTag,
		Range:       rangeStr,
		Commits:     commits,
		DiffStat:    diffStat,
		DiffSnippet: diffSnippet,
	}, nil
}

func buildSystemPrompt() string {
	return `你是一名开源项目发布说明编辑。根据 Git 提交记录、Diff Stat 与 Diff Snippet，撰写准确、简洁的 GitHub Release 正文。

## 排版要求
1. 全文使用简体中文；代码符号、产品名、路径保持原样。参考以下版式，仅模仿信息层次，不借用示例项目的内容：
   > [!IMPORTANT]
   > 本次最值得关注的实际变化及其影响。可以分成两段，每段聚焦一个主题。

   ### 新增特性
   - 具体变化及其用户价值。

   ### 修复与改进
   - 具体修复及其影响。
2. 重点摘要只写能够从输入核实的变化，突出影响；不要重复罗列提交。分类条目同样基于事实，按主题归纳，简明直接。
3. 仅输出本次确实涉及的分类；基础设施等变更可放入“### 其他变更”。分类标题不加 emoji。
4. 不要输出版本标题、逐条提交记录、对比链接、贡献者鸣谢或总结段落；程序会在正文之后附加提交记录与完整变更对比。
5. 直接输出 Markdown 正文，不要使用代码围栏或客套话。`
}

func buildUserPrompt(ctx *gitContext) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("项目名称: Lumina · 微明 (仓库: %s)\n", ctx.Repo))
	sb.WriteString(fmt.Sprintf("当前发布版本: %s\n", ctx.CurrentTag))
	if ctx.PreviousTag != "" {
		sb.WriteString(fmt.Sprintf("上一对比版本: %s (区间: %s)\n\n", ctx.PreviousTag, ctx.Range))
	} else {
		sb.WriteString("上一对比版本: 首发版本\n\n")
	}

	sb.WriteString("### Git Commits 列表:\n")
	if len(ctx.Commits) == 0 {
		sb.WriteString("(无独立提交)\n\n")
	} else {
		for _, c := range ctx.Commits {
			sb.WriteString("- " + c + "\n")
		}
		sb.WriteString("\n")
	}

	if ctx.DiffStat != "" {
		sb.WriteString("### 变更统计 (Diff Stat):\n```text\n")
		sb.WriteString(ctx.DiffStat)
		sb.WriteString("\n```\n\n")
	}

	if ctx.DiffSnippet != "" {
		sb.WriteString("### 关键代码变动摘要 (Diff Snippet):\n```diff\n")
		sb.WriteString(ctx.DiffSnippet)
		sb.WriteString("\n```\n\n")
	}

	sb.WriteString("请严格按照前述规范生成正式的 Release Notes Markdown。")
	return sb.String()
}

func generateWithAI(ctx *gitContext, baseURL, model, apiKey string) (string, error) {
	reqBody := openAIChatRequest{
		Model: model,
		Messages: []openAIChatMessage{
			{Role: "system", Content: buildSystemPrompt()},
			{Role: "user", Content: buildUserPrompt(ctx)},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	endpoint := baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(context.Background(), http.MethodPost, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建 HTTP 请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP 请求执行失败: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResp openAIChatResponse
	if err := json.Unmarshal(bodyBytes, &chatResp); err != nil {
		return "", fmt.Errorf("解析响应 JSON 失败: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("AI 错误响应: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 || strings.TrimSpace(chatResp.Choices[0].Message.Content) == "" {
		return "", fmt.Errorf("AI 返回内容为空")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

// renderFallbackTemplate 在没有 AI 或 AI 失败时按提交类型整理正文。
func renderFallbackTemplate(ctx *gitContext) string {
	var featGroup, fixGroup, otherGroup []string
	reType := regexp.MustCompile(`^[[:xdigit:]]+\s+([a-zA-Z]+)(?:\([^)]+\))?!?:\s*(.*?)(?:\s+\([^()]+\))?$`)

	for _, commit := range ctx.Commits {
		matches := reType.FindStringSubmatch(commit)
		if matches == nil {
			otherGroup = append(otherGroup, "- "+commit)
			continue
		}

		item := "- " + matches[2]
		switch strings.ToLower(matches[1]) {
		case "feat":
			featGroup = append(featGroup, item)
		case "fix", "perf", "refactor":
			fixGroup = append(fixGroup, item)
		default:
			otherGroup = append(otherGroup, item)
		}
	}

	var sb strings.Builder
	sb.WriteString("> [!IMPORTANT]\n")
	if len(ctx.Commits) == 0 {
		sb.WriteString("> 本次发布暂无可列出的独立提交。\n")
	} else {
		highlights := make([]string, 0, 2)
		for _, group := range [][]string{featGroup, fixGroup, otherGroup} {
			if len(group) > 0 {
				highlights = append(highlights, strings.TrimPrefix(group[0], "- "))
			}
			if len(highlights) == 2 {
				break
			}
		}
		sb.WriteString("> 本次更新重点：" + strings.Join(highlights, "；") + "。\n")
	}

	for _, group := range []struct {
		title string
		items []string
	}{
		{"新增特性", featGroup},
		{"修复与改进", fixGroup},
		{"其他变更", otherGroup},
	} {
		if len(group.items) > 0 {
			sb.WriteString("\n### " + group.title + "\n")
			for _, item := range group.items {
				sb.WriteString(item + "\n")
			}
		}
	}

	return strings.TrimSpace(sb.String())
}

func formatReleaseNotes(body string, ctx *gitContext) string {
	var sb strings.Builder
	sb.WriteString(strings.TrimSpace(body))

	if len(ctx.Commits) > 0 {
		sb.WriteString("\n\n### 变更记录\n")
		for _, commit := range ctx.Commits {
			sb.WriteString("- " + commit + "\n")
		}
	}

	if ctx.PreviousTag != "" && ctx.PreviousTag != ctx.CurrentTag {
		compareURL := fmt.Sprintf("https://github.com/%s/compare/%s...%s", ctx.Repo, ctx.PreviousTag, ctx.CurrentTag)
		sb.WriteString(fmt.Sprintf("\n完整变更记录：[%s...%s](%s)\n", ctx.PreviousTag, ctx.CurrentTag, compareURL))
	}

	return strings.TrimSpace(sb.String())
}
