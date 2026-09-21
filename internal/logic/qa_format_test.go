package logic

import (
	"testing"
)

func TestFormatReviewAnswer(t *testing.T) {
	t.Parallel()

	t.Run("approve without feedback", func(t *testing.T) {
		ans := map[string]interface{}{
			"decision": "approve",
		}
		got := formatReviewAnswer(ans)
		want := "用户批准了该修改"
		if got != want {
			t.Errorf("formatReviewAnswer() = %q, want %q", got, want)
		}
	})

	t.Run("revise with feedback and annotations", func(t *testing.T) {
		ans := map[string]interface{}{
			"decision": "revise",
			"annotations": []interface{}{
				map[string]interface{}{
					"sectionId": "target",
					"content":   "请将目录命名调整为 profiles_v2",
				},
				map[string]interface{}{
					"sectionId": "verification",
					"content":   "补充压力测试标准",
				},
			},
			"feedback": "整体方向很好，注意测试补充",
		}
		got := formatReviewAnswer(ans)
		want := "用户要求修改\n[REVISIONS]\n1. [target] 请将目录命名调整为 profiles_v2\n2. [verification] 补充压力测试标准\n[FEEDBACK] 整体方向很好，注意测试补充"
		if got != want {
			t.Errorf("formatReviewAnswer() = %q, want %q", got, want)
		}
	})
}

func TestToJSON(t *testing.T) {
	t.Parallel()

	t.Run("nil input", func(t *testing.T) {
		if got := toJSON(nil); got != nil {
			t.Errorf("toJSON(nil) = %v, want nil", got)
		}
	})

	t.Run("valid JSON string preservation", func(t *testing.T) {
		raw := `{"sections":[{"id":"target","title":"1. 目标"}]}`
		got := toJSON(raw)
		if string(got) != raw {
			t.Errorf("toJSON(raw) = %s, want %s", string(got), raw)
		}
	})

	t.Run("map marshaling", func(t *testing.T) {
		m := map[string]string{"foo": "bar"}
		got := toJSON(m)
		want := `{"foo":"bar"}`
		if string(got) != want {
			t.Errorf("toJSON(m) = %s, want %s", string(got), want)
		}
	})
}
