package mcp

import (
	"strings"
	"testing"
)

func TestPinToolDescriptionsMatchResolveRules(t *testing.T) {
	t.Parallel()

	for _, def := range pinToolDefs {
		if strings.Contains(def.description, "项目名称/别名") || strings.Contains(def.description, "支持名称/别名") {
			t.Errorf("%s description still allows Project.Name: %s", def.name, def.description)
		}
	}

	push := pinToolDefs[0]
	if push.name != "pin_push" {
		t.Fatalf("first tool = %s, want pin_push", push.name)
	}
	if strings.Contains(push.description, "为可选") || strings.Contains(push.description, "不传时") {
		t.Errorf("pin_push still describes from_project_id as optional")
	}
	if !strings.Contains(push.description, "from_project_id 必填") {
		t.Errorf("pin_push should say from_project_id is required")
	}
}
