package service

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	xEnv "github.com/bamboo-services/bamboo-base-go/defined/env"
	"go.yaml.in/yaml/v3"

	apiPlugin "github.com/xiaolfeng/Lumina/api/plugin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/resources"
)

const (
	pluginManifestPath = ".claude-plugin/plugin.json"
	pluginMCPPath      = ".mcp.json"
	pluginSkillsDir    = "skills"
	skillMarkdownName  = "SKILL.md"
	wellKnownSkillMD   = "skill-md"
)

// PluginManifest 对齐 Claude Code plugin.json 的运行时读取结构。
type PluginManifest struct {
	Name        string                  `json:"name"`
	Version     string                  `json:"version"`
	Description string                  `json:"description"`
	Author      *apiPlugin.PluginAuthor `json:"author,omitempty"`
	Homepage    string                  `json:"homepage,omitempty"`
	Repository  string                  `json:"repository,omitempty"`
	License     string                  `json:"license,omitempty"`
	Keywords    []string                `json:"keywords,omitempty"`
}

// SkillFile 单个技能的磁盘/内嵌内容。
type SkillFile struct {
	Name        string // 技能标识（目录名 / frontmatter name）
	Description string
	RelPath     string // zip 内路径，如 skills/lumina-mcp/SKILL.md
	Content     []byte
	SHA256      string // 技能文件原始字节的十六进制摘要
}

// PluginBundle 一次打包得到的 ZIP、哈希与清单快照。
type PluginBundle struct {
	Zip      []byte
	SHA256   string
	Manifest PluginManifest
	Skills   []SkillFile
}

// AIPluginService 从内嵌资源（或开发目录）打包 AI 插件。
//
// 生产模式使用 go:embed 结果并惰性缓存；调试模式（XLF_DEBUG=true）
// 每次请求从 resources/ai-plugin 重读，修改 SKILL.md 后无需重启即可生效。
type AIPluginService struct {
	mu     sync.Mutex
	cached *PluginBundle
}

// NewAIPluginService 创建 AI 插件打包服务。
func NewAIPluginService() *AIPluginService {
	return &AIPluginService{}
}

// Bundle 返回当前插件 ZIP 快照。调试模式下不使用缓存。
func (s *AIPluginService) Bundle() (*PluginBundle, error) {
	if s.hotReload() {
		return s.build()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cached != nil {
		return s.cached, nil
	}
	bundle, err := s.build()
	if err != nil {
		return nil, err
	}
	s.cached = bundle
	return bundle, nil
}

// BundleForInstance 返回带当前 Lumina 地址的插件 ZIP。
// Claude Code 安装插件后会自动读取根目录的 .mcp.json，无需再次手动添加 MCP。
func (s *AIPluginService) BundleForInstance(baseURL string) (*PluginBundle, error) {
	bundle, err := s.Bundle()
	if err != nil {
		return nil, err
	}
	zipBytes, err := packZipWithMCP(bundle.Zip, buildPluginMCPConfig(baseURL))
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(zipBytes)
	return &PluginBundle{
		Zip:      zipBytes,
		SHA256:   hex.EncodeToString(sum[:]),
		Manifest: bundle.Manifest,
		Skills:   bundle.Skills,
	}, nil
}

// MarketplaceJSON 按请求客户端渲染 Claude Code marketplace.json。
//
// ZCode 不支持 archive 源类型（会报 "Plugin source is invalid or
// unsupported"），故按 User-Agent 输出其 url+zip 变体；其余客户端
// 保持 Claude Code 的 archive 形态。
func (s *AIPluginService) MarketplaceJSON(baseURL, userAgent string) ([]byte, *PluginBundle, error) {
	bundle, err := s.BundleForInstance(baseURL)
	if err != nil {
		return nil, nil, err
	}

	manifest := bundle.Manifest
	pluginName := firstNonEmpty(manifest.Name, bConst.AIPluginName)
	doc := apiPlugin.Marketplace{
		Name:        bConst.AIPluginMarketplaceName,
		Description: firstNonEmpty(manifest.Description, "Lumina AI 插件与技能"),
		Version:     manifest.Version,
		Owner: apiPlugin.MarketplaceOwner{
			Name: bConst.AIPluginOwnerName,
			URL:  manifest.Homepage,
		},
		Plugins: []apiPlugin.MarketplacePlugin{
			{
				Name:        pluginName,
				Description: manifest.Description,
				Version:     manifest.Version,
				Author:      manifest.Author,
				Homepage:    manifest.Homepage,
				Repository:  manifest.Repository,
				License:     manifest.License,
				Keywords:    manifest.Keywords,
				Source:      marketplaceSource(baseURL, bundle.SHA256, isZcodeUserAgent(userAgent)),
			},
		},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("序列化 marketplace.json 失败: %w", err)
	}
	data = append(data, '\n')
	return data, bundle, nil
}

// marketplaceSource 按客户端产出 ZIP 源：ZCode 走 url+zip，
// Claude Code 及其他客户端走 archive。
func marketplaceSource(baseURL, sha256 string, zcode bool) apiPlugin.MarketplaceSource {
	zipURL := joinURL(baseURL, bConst.AIPluginZipPath)
	if zcode {
		return apiPlugin.MarketplaceSource{Source: "url", Type: "zip", URL: zipURL, SHA256: sha256}
	}
	return apiPlugin.MarketplaceSource{Source: "archive", URL: zipURL, SHA256: sha256}
}

// isZcodeUserAgent 识别 ZCode 客户端（其 HTTP 请求固定携带
// "ZCode/<版本>" User-Agent，见 ZCode 客户端 sourceHeaders 实现）。
func isZcodeUserAgent(userAgent string) bool {
	return strings.Contains(strings.ToLower(userAgent), "zcode/")
}

// WellKnownJSON 按 Agent Skills 发现协议渲染 /.well-known/skills/index.json。
func (s *AIPluginService) WellKnownJSON(baseURL string) ([]byte, *PluginBundle, error) {
	bundle, err := s.Bundle()
	if err != nil {
		return nil, nil, err
	}

	skills := make([]apiPlugin.WellKnownSkill, 0, len(bundle.Skills))
	for _, skill := range bundle.Skills {
		skills = append(skills, apiPlugin.WellKnownSkill{
			Name:        skill.Name,
			Type:        wellKnownSkillMD,
			Description: skill.Description,
			URL:         joinURL(baseURL, bConst.AIPluginWellKnownPath+"/"+skill.Name+"/"+skillMarkdownName),
			Digest:      "sha256:" + skill.SHA256,
		})
	}

	doc := apiPlugin.WellKnownIndex{
		Schema: bConst.AIPluginWellKnownSchema,
		Skills: skills,
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("序列化 well-known index 失败: %w", err)
	}
	data = append(data, '\n')
	return data, bundle, nil
}

// SkillByName 按技能标识读取 SKILL.md。未命中返回 (nil, false, nil)。
func (s *AIPluginService) SkillByName(name string) (*SkillFile, bool, error) {
	if !validSkillName(name) {
		return nil, false, nil
	}
	bundle, err := s.Bundle()
	if err != nil {
		return nil, false, err
	}
	for i := range bundle.Skills {
		if bundle.Skills[i].Name == name {
			skill := bundle.Skills[i]
			return &skill, true, nil
		}
	}
	return nil, false, nil
}

func (s *AIPluginService) hotReload() bool {
	if xEnv.GetEnvString(bConst.EnvAIPluginDir, "") != "" {
		return true
	}
	return xEnv.GetEnvBool(xEnv.Debug, false)
}

func (s *AIPluginService) build() (*PluginBundle, error) {
	fsys, err := openPluginFS()
	if err != nil {
		return nil, err
	}

	manifest, err := readPluginManifest(fsys)
	if err != nil {
		return nil, err
	}
	skills, err := collectSkills(fsys)
	if err != nil {
		return nil, err
	}
	zipBytes, err := packZip(fsys)
	if err != nil {
		return nil, err
	}

	sum := sha256.Sum256(zipBytes)
	return &PluginBundle{
		Zip:      zipBytes,
		SHA256:   hex.EncodeToString(sum[:]),
		Manifest: manifest,
		Skills:   skills,
	}, nil
}

func openPluginFS() (fs.FS, error) {
	if dir := strings.TrimSpace(xEnv.GetEnvString(bConst.EnvAIPluginDir, "")); dir != "" {
		info, err := os.Stat(dir)
		if err != nil {
			return nil, fmt.Errorf("读取 %s=%s 失败: %w", bConst.EnvAIPluginDir, dir, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s=%s 不是目录", bConst.EnvAIPluginDir, dir)
		}
		return os.DirFS(dir), nil
	}

	if xEnv.GetEnvBool(xEnv.Debug, false) {
		if info, err := os.Stat(bConst.AIPluginDebugDir); err == nil && info.IsDir() {
			return os.DirFS(bConst.AIPluginDebugDir), nil
		}
	}

	sub, err := fs.Sub(resources.AIPluginFS, bConst.AIPluginEmbedRoot)
	if err != nil {
		return nil, fmt.Errorf("打开内嵌插件目录失败: %w", err)
	}
	return sub, nil
}

func readPluginManifest(fsys fs.FS) (PluginManifest, error) {
	data, err := fs.ReadFile(fsys, pluginManifestPath)
	if err != nil {
		return PluginManifest{}, fmt.Errorf("读取 %s 失败: %w", pluginManifestPath, err)
	}
	var manifest PluginManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return PluginManifest{}, fmt.Errorf("解析 %s 失败: %w", pluginManifestPath, err)
	}
	if strings.TrimSpace(manifest.Name) == "" {
		manifest.Name = bConst.AIPluginName
	}
	return manifest, nil
}

func collectSkills(fsys fs.FS) ([]SkillFile, error) {
	entries, err := fs.ReadDir(fsys, pluginSkillsDir)
	if err != nil {
		return nil, fmt.Errorf("读取技能目录失败: %w", err)
	}

	skills := make([]SkillFile, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !validSkillName(name) {
			continue
		}
		relPath := path.Join(pluginSkillsDir, name, skillMarkdownName)
		content, err := fs.ReadFile(fsys, relPath)
		if err != nil {
			return nil, fmt.Errorf("读取 %s 失败: %w", relPath, err)
		}
		fmName, desc := parseSkillFrontmatter(content)
		if fmName == "" {
			fmName = name
		}
		sum := sha256.Sum256(content)
		skills = append(skills, SkillFile{
			Name:        fmName,
			Description: desc,
			RelPath:     relPath,
			Content:     content,
			SHA256:      hex.EncodeToString(sum[:]),
		})
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, nil
}

func packZip(fsys fs.FS) ([]byte, error) {
	var names []string
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		name := path.Clean(p)
		if skipZipEntry(name) {
			return nil
		}
		names = append(names, name)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("遍历插件文件失败: %w", err)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("插件目录为空，无法打包")
	}
	sort.Strings(names)

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, name := range names {
		data, readErr := fs.ReadFile(fsys, name)
		if readErr != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("读取 %s 失败: %w", name, readErr)
		}
		header := &zip.FileHeader{
			Name:   name,
			Method: zip.Deflate,
		}
		header.Modified = time.Unix(0, 0).UTC()
		header.SetMode(0o644)
		writer, createErr := zw.CreateHeader(header)
		if createErr != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("写入 zip 条目 %s 失败: %w", name, createErr)
		}
		if _, writeErr := writer.Write(data); writeErr != nil {
			_ = zw.Close()
			return nil, fmt.Errorf("写入 zip 内容 %s 失败: %w", name, writeErr)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("关闭 zip 失败: %w", err)
	}
	return buf.Bytes(), nil
}

func packZipWithMCP(sourceZip, mcpConfig []byte) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(sourceZip), int64(len(sourceZip)))
	if err != nil {
		return nil, fmt.Errorf("打开插件 zip 失败: %w", err)
	}

	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, file := range reader.File {
		if file.Name == pluginMCPPath {
			continue
		}
		data, readErr := readZipFile(file)
		if readErr != nil {
			_ = writer.Close()
			return nil, readErr
		}
		if writeErr := writeZipFile(writer, file.Name, data); writeErr != nil {
			_ = writer.Close()
			return nil, writeErr
		}
	}
	if err := writeZipFile(writer, pluginMCPPath, mcpConfig); err != nil {
		_ = writer.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("关闭实例插件 zip 失败: %w", err)
	}
	return buf.Bytes(), nil
}

func buildPluginMCPConfig(baseURL string) []byte {
	config := map[string]any{
		bConst.AIPluginName: map[string]any{
			"type": "http",
			"url":  joinURL(baseURL, bConst.AIPluginMCPPath),
			"headers": map[string]string{
				"Authorization": "Bearer ${LUMINA_API_KEY}",
			},
		},
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil
	}
	return append(data, '\n')
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开 zip 条目 %s 失败: %w", file.Name, err)
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("读取 zip 条目 %s 失败: %w", file.Name, err)
	}
	return data, nil
}

func writeZipFile(writer *zip.Writer, name string, data []byte) error {
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.Modified = time.Unix(0, 0).UTC()
	header.SetMode(0o644)
	entry, err := writer.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("写入 zip 条目 %s 失败: %w", name, err)
	}
	if _, err := entry.Write(data); err != nil {
		return fmt.Errorf("写入 zip 内容 %s 失败: %w", name, err)
	}
	return nil
}

func skipZipEntry(name string) bool {
	if name == "." || name == "" {
		return true
	}
	base := path.Base(name)
	if base == ".DS_Store" || strings.HasPrefix(base, "._") {
		return true
	}
	return false
}

func parseSkillFrontmatter(content []byte) (name, description string) {
	text := string(content)
	if !strings.HasPrefix(text, "---") {
		return "", ""
	}
	rest := strings.TrimPrefix(text, "---")
	rest = strings.TrimPrefix(rest, "\r")
	rest = strings.TrimPrefix(rest, "\n")
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return "", ""
	}

	var meta struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}
	if err := yaml.Unmarshal([]byte(rest[:end]), &meta); err != nil {
		return "", ""
	}
	return strings.TrimSpace(meta.Name), strings.TrimSpace(meta.Description)
}

func validSkillName(name string) bool {
	if len(name) < 1 || len(name) > 64 {
		return false
	}
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	prevHyphen := false
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9':
			prevHyphen = false
		case c == '-':
			if prevHyphen {
				return false
			}
			prevHyphen = true
		default:
			return false
		}
	}
	return true
}

func joinURL(base, p string) string {
	base = strings.TrimRight(strings.TrimSpace(base), "/")
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if base == "" {
		return p
	}
	return base + p
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
