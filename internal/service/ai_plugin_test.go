package service

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"testing"

	apiPlugin "github.com/xiaolfeng/Lumina/api/plugin"
	bConst "github.com/xiaolfeng/Lumina/internal/constant"
)

var expectedSkillNames = []string{
	"lumina-pin",
	"lumina-preview",
	"lumina-qa",
	"lumina-repowiki",
}

func TestPluginBundleContainsManifestAndSkills(t *testing.T) {
	svc := NewAIPluginService()
	bundle, err := svc.Bundle()
	if err != nil {
		t.Fatalf("打包失败: %v", err)
	}
	if len(bundle.Zip) == 0 {
		t.Fatal("ZIP 为空")
	}
	if len(bundle.SHA256) != 64 {
		t.Fatalf("SHA-256 长度应为 64，实际 %d (%q)", len(bundle.SHA256), bundle.SHA256)
	}

	sum := sha256.Sum256(bundle.Zip)
	if got := hex.EncodeToString(sum[:]); got != bundle.SHA256 {
		t.Fatalf("ZIP SHA-256 与声明不一致：got=%s want=%s", got, bundle.SHA256)
	}
	if bundle.Manifest.Name != bConst.AIPluginName {
		t.Fatalf("plugin.json name = %q, want %q", bundle.Manifest.Name, bConst.AIPluginName)
	}

	reader, err := zip.NewReader(bytes.NewReader(bundle.Zip), int64(len(bundle.Zip)))
	if err != nil {
		t.Fatalf("打开 ZIP 失败: %v", err)
	}
	names := make(map[string]struct{}, len(reader.File))
	for _, file := range reader.File {
		names[file.Name] = struct{}{}
	}
	if _, ok := names[pluginManifestPath]; !ok {
		t.Fatalf("ZIP 缺少 %s", pluginManifestPath)
	}
	for _, skill := range expectedSkillNames {
		rel := "skills/" + skill + "/" + skillMarkdownName
		if _, ok := names[rel]; !ok {
			t.Fatalf("ZIP 缺少 %s", rel)
		}
	}
	for _, extra := range []string{
		"skills/_shared/lumina-core-rules.md",
		"skills/_shared/project-resolver.md",
		"skills/_shared/preview-qa-contract.md",
		"skills/lumina-qa/examples/select-with-supplement.md",
		"skills/lumina-preview/assets/index.html",
		"skills/lumina-pin/references/fields.md",
		"skills/lumina-repowiki/examples/list-and-query.md",
	} {
		if _, ok := names[extra]; !ok {
			t.Fatalf("ZIP 缺少 %s", extra)
		}
	}
	if len(bundle.Skills) != len(expectedSkillNames) {
		t.Fatalf("技能数量 = %d, want %d", len(bundle.Skills), len(expectedSkillNames))
	}
}

func TestPluginBundleForInstanceContainsAutomaticMCPConfig(t *testing.T) {
	const base = "https://lumina.example"
	bundle, err := NewAIPluginService().BundleForInstance(base)
	if err != nil {
		t.Fatalf("实例插件打包失败: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(bundle.Zip), int64(len(bundle.Zip)))
	if err != nil {
		t.Fatalf("打开实例插件 ZIP 失败: %v", err)
	}
	var mcpConfig []byte
	for _, file := range reader.File {
		if file.Name != pluginMCPPath {
			continue
		}
		mcpConfig, err = readZipFile(file)
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", pluginMCPPath, err)
		}
		break
	}
	if len(mcpConfig) == 0 {
		t.Fatalf("实例插件 ZIP 缺少 %s", pluginMCPPath)
	}

	var config map[string]struct {
		Type    string            `json:"type"`
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	}
	if err := json.Unmarshal(mcpConfig, &config); err != nil {
		t.Fatalf("解析 %s 失败: %v\n%s", pluginMCPPath, err, mcpConfig)
	}
	server, ok := config[bConst.AIPluginName]
	if !ok {
		t.Fatalf("%s 缺少 %s server", pluginMCPPath, bConst.AIPluginName)
	}
	if server.Type != "http" {
		t.Fatalf("MCP type = %q, want http", server.Type)
	}
	if want := base + bConst.AIPluginMCPPath; server.URL != want {
		t.Fatalf("MCP url = %q, want %q", server.URL, want)
	}
	if len(server.Headers) != 0 {
		// OAuth 直连模式：不带 Authorization 头，由客户端发起 OAuth 登录
		t.Fatalf("MCP headers = %v, want 空（OAuth 模式）", server.Headers)
	}
	if bytes.Contains(mcpConfig, []byte("LUMINA_API_KEY")) {
		t.Fatal("OAuth 模式下 .mcp.json 不应引用 LUMINA_API_KEY")
	}
}

func TestPluginBundleForInstanceDeterministic(t *testing.T) {
	const base = "https://lumina.example"
	first, err := NewAIPluginService().BundleForInstance(base)
	if err != nil {
		t.Fatalf("第一次实例打包失败: %v", err)
	}
	second, err := NewAIPluginService().BundleForInstance(base)
	if err != nil {
		t.Fatalf("第二次实例打包失败: %v", err)
	}
	if first.SHA256 != second.SHA256 || !bytes.Equal(first.Zip, second.Zip) {
		t.Fatal("相同实例地址应生成稳定一致的插件 ZIP")
	}
}

func TestPluginBundleDeterministic(t *testing.T) {
	first, err := NewAIPluginService().Bundle()
	if err != nil {
		t.Fatalf("第一次打包失败: %v", err)
	}
	second, err := NewAIPluginService().Bundle()
	if err != nil {
		t.Fatalf("第二次打包失败: %v", err)
	}
	if first.SHA256 != second.SHA256 {
		t.Fatalf("两次打包 SHA-256 不一致: %s vs %s", first.SHA256, second.SHA256)
	}
	if !bytes.Equal(first.Zip, second.Zip) {
		t.Fatal("两次打包 ZIP 字节不一致")
	}
}

func TestMarketplaceJSONUsesRequestHostAndZipHash(t *testing.T) {
	svc := NewAIPluginService()
	const base = "http://127.0.0.1:8800"
	data, bundle, err := svc.MarketplaceJSON(base, "claude-cli/2.1.240 (external, cli)")
	if err != nil {
		t.Fatalf("渲染 marketplace.json 失败: %v", err)
	}

	var doc apiPlugin.Marketplace
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("解析 marketplace.json 失败: %v\n%s", err, data)
	}
	if doc.Name != bConst.AIPluginMarketplaceName {
		t.Fatalf("marketplace name = %q, want %q", doc.Name, bConst.AIPluginMarketplaceName)
	}
	if len(doc.Plugins) != 1 {
		t.Fatalf("plugins 数量 = %d, want 1", len(doc.Plugins))
	}
	plugin := doc.Plugins[0]
	if plugin.Name != bConst.AIPluginName {
		t.Fatalf("plugin name = %q, want %q", plugin.Name, bConst.AIPluginName)
	}
	if plugin.Source.Source != "archive" {
		t.Fatalf("source.source = %q, want archive", plugin.Source.Source)
	}
	if plugin.Source.Type != "" {
		t.Fatalf("source.type = %q, want 空", plugin.Source.Type)
	}
	wantURL := base + bConst.AIPluginZipPath
	if plugin.Source.URL != wantURL {
		t.Fatalf("source.url = %q, want %q", plugin.Source.URL, wantURL)
	}
	if plugin.Source.SHA256 != bundle.SHA256 {
		t.Fatalf("source.sha256 = %q, want %q", plugin.Source.SHA256, bundle.SHA256)
	}
	mcpConfig := readZipEntryForTest(t, bundle.Zip, pluginMCPPath)
	if !bytes.Contains(mcpConfig, []byte(base+bConst.AIPluginMCPPath)) {
		t.Fatalf("市场清单指向的 ZIP 未绑定当前实例 MCP 地址: %s", mcpConfig)
	}
}

func TestMarketplaceJSONServesZcodeURLZipVariant(t *testing.T) {
	svc := NewAIPluginService()
	const base = "https://lumina.example"

	t.Run("zcode user agent", func(t *testing.T) {
		data, bundle, err := svc.MarketplaceJSON(base, "ZCode/0.5.13")
		if err != nil {
			t.Fatalf("渲染 ZCode marketplace.json 失败: %v", err)
		}
		var doc apiPlugin.Marketplace
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("解析 ZCode marketplace.json 失败: %v\n%s", err, data)
		}
		source := doc.Plugins[0].Source
		if source.Source != "url" || source.Type != "zip" {
			t.Fatalf("source = (%q, %q), want (url, zip)", source.Source, source.Type)
		}
		if want := base + bConst.AIPluginZipPath; source.URL != want {
			t.Fatalf("source.url = %q, want %q", source.URL, want)
		}
		if source.SHA256 != bundle.SHA256 {
			t.Fatalf("source.sha256 = %q, want %q", source.SHA256, bundle.SHA256)
		}
	})

	t.Run("case insensitive match", func(t *testing.T) {
		data, _, err := svc.MarketplaceJSON(base, "zcode/0.5.13")
		if err != nil {
			t.Fatalf("渲染 marketplace.json 失败: %v", err)
		}
		var doc apiPlugin.Marketplace
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatalf("解析 marketplace.json 失败: %v", err)
		}
		if source := doc.Plugins[0].Source; source.Source != "url" {
			t.Fatalf("source.source = %q, want url", source.Source)
		}
	})

	t.Run("empty and foreign agents keep archive", func(t *testing.T) {
		for _, ua := range []string{"", "curl/8.7.1"} {
			data, _, err := svc.MarketplaceJSON(base, ua)
			if err != nil {
				t.Fatalf("渲染 marketplace.json 失败 (ua=%q): %v", ua, err)
			}
			var doc apiPlugin.Marketplace
			if err := json.Unmarshal(data, &doc); err != nil {
				t.Fatalf("解析 marketplace.json 失败 (ua=%q): %v", ua, err)
			}
			if source := doc.Plugins[0].Source; source.Source != "archive" {
				t.Fatalf("ua=%q source.source = %q, want archive", ua, source.Source)
			}
		}
	})
}

func TestMarketplaceJSONZcodeForced(t *testing.T) {
	svc := NewAIPluginService()
	const base = "https://lumina.example"

	// Zcode 专用端点无论 UA 如何都输出 url+zip 形态
	data, _, err := svc.MarketplaceJSONZcode(base)
	if err != nil {
		t.Fatalf("渲染 ZCode 市场清单失败: %v", err)
	}
	var doc apiPlugin.Marketplace
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("解析 ZCode 市场清单失败: %v\n%s", err, data)
	}
	source := doc.Plugins[0].Source
	if source.Source != "url" || source.Type != "zip" {
		t.Fatalf("source = (%q, %q), want (url, zip)", source.Source, source.Type)
	}

	// 非浏览器 UA 访问常规端点时仍保持 archive 形态（Claude Code 语义）
	data, _, err = svc.MarketplaceJSON(base, "")
	if err != nil {
		t.Fatalf("渲染市场清单失败: %v", err)
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("解析市场清单失败: %v", err)
	}
	if source := doc.Plugins[0].Source; source.Source != "archive" {
		t.Fatalf("source.source = %q, want archive", source.Source)
	}
}

func readZipEntryForTest(t *testing.T, zipBytes []byte, name string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("打开 ZIP 失败: %v", err)
	}
	for _, file := range reader.File {
		if file.Name != name {
			continue
		}
		data, err := readZipFile(file)
		if err != nil {
			t.Fatalf("读取 ZIP 条目 %s 失败: %v", name, err)
		}
		return data
	}
	t.Fatalf("ZIP 缺少 %s", name)
	return nil
}

func TestWellKnownJSONListsEmbeddedSkills(t *testing.T) {
	svc := NewAIPluginService()
	const base = "https://lumina.example"
	data, _, err := svc.WellKnownJSON(base)
	if err != nil {
		t.Fatalf("渲染 well-known index 失败: %v", err)
	}

	var doc apiPlugin.WellKnownIndex
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("解析 well-known index 失败: %v\n%s", err, data)
	}
	if doc.Schema != bConst.AIPluginWellKnownSchema {
		t.Fatalf("$schema = %q, want %q", doc.Schema, bConst.AIPluginWellKnownSchema)
	}
	if len(doc.Skills) != len(expectedSkillNames) {
		t.Fatalf("skills 数量 = %d, want %d", len(doc.Skills), len(expectedSkillNames))
	}

	got := make(map[string]apiPlugin.WellKnownSkill, len(doc.Skills))
	for _, skill := range doc.Skills {
		got[skill.Name] = skill
		if skill.Type != wellKnownSkillMD {
			t.Fatalf("skill %s type = %q, want %s", skill.Name, skill.Type, wellKnownSkillMD)
		}
		if skill.Description == "" {
			t.Fatalf("skill %s 缺少 description", skill.Name)
		}
		if !strings.HasPrefix(skill.Digest, "sha256:") || len(skill.Digest) != len("sha256:")+64 {
			t.Fatalf("skill %s digest 格式错误: %q", skill.Name, skill.Digest)
		}
		wantURL := base + bConst.AIPluginWellKnownPath + "/" + skill.Name + "/" + skillMarkdownName
		if skill.URL != wantURL {
			t.Fatalf("skill %s url = %q, want %q", skill.Name, skill.URL, wantURL)
		}
	}
	for _, name := range expectedSkillNames {
		if _, ok := got[name]; !ok {
			t.Fatalf("well-known 缺少技能 %s", name)
		}
		skill, exists, err := svc.SkillByName(name)
		if err != nil || !exists {
			t.Fatalf("SkillByName(%s) 失败 exists=%v err=%v", name, exists, err)
		}
		if "sha256:"+skill.SHA256 != got[name].Digest {
			t.Fatalf("skill %s digest 与文件哈希不一致", name)
		}
		if !bytes.Contains(skill.Content, []byte("name: "+name)) {
			t.Fatalf("skill %s 正文未包含 frontmatter name", name)
		}
	}
}

func TestSkillByNameRejectsUnknown(t *testing.T) {
	svc := NewAIPluginService()
	if _, ok, err := svc.SkillByName("not-a-skill"); err != nil || ok {
		t.Fatalf("未知技能应返回 ok=false, 实际 ok=%v err=%v", ok, err)
	}
	if _, ok, err := svc.SkillByName("../lumina-qa"); err != nil || ok {
		t.Fatalf("路径穿越应被拒绝, 实际 ok=%v err=%v", ok, err)
	}
}
