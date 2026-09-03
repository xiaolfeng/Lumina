package logic

import (
	"testing"

	"gorm.io/datatypes"
)

func TestSupplementDeniedByQuestionStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		status string
		want   string
	}{
		{status: "pending", want: ""},
		{status: "answered", want: "问题已回答，无法推送补充内容"},
		{status: "skipped", want: "问题已跳过，无法推送补充内容"},
		{status: "", want: ""},
	}
	for _, tc := range cases {
		if got := supplementDeniedByQuestionStatus(tc.status); got != tc.want {
			t.Errorf("supplementDeniedByQuestionStatus(%q) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestQuestionHasOption(t *testing.T) {
	t.Parallel()

	options := datatypes.JSON([]byte(`[{"id":"opt-1","label":"A"},{"id":"opt-2","label":"B"}]`))

	if !questionHasOption(options, "opt-1") {
		t.Error("expected opt-1 to belong to the question")
	}
	if questionHasOption(options, "opt-missing") {
		t.Error("expected missing option to be rejected")
	}
	if questionHasOption(nil, "opt-1") {
		t.Error("expected empty options to reject")
	}
	if questionHasOption(options, "") {
		t.Error("expected empty option id to reject")
	}
}
