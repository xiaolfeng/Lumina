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
	data, bundle, err := svc.MarketplaceJSON(base)
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
	wantURL := base + bConst.AIPluginZipPath
	if plugin.Source.URL != wantURL {
		t.Fatalf("source.url = %q, want %q", plugin.Source.URL, wantURL)
	}
	if plugin.Source.SHA256 != bundle.SHA256 {
		t.Fatalf("source.sha256 = %q, want %q", plugin.Source.SHA256, bundle.SHA256)
	}
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
