package logic

import (
	"testing"

	bConst "github.com/xiaolfeng/Lumina/internal/constant"
	"github.com/xiaolfeng/Lumina/internal/entity"
)

func TestIncrementPatchVersion(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"v1.0.0": "v1.0.1",
		"v1.1.9": "v1.1.10",
		"1.2.3":  "v1.2.4",
		"v2.0":   "v1.0.0",
		"":       "v1.0.0",
	}
	for input, want := range cases {
		if got := incrementPatchVersion(input); got != want {
			t.Errorf("incrementPatchVersion(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestIsHTMLFilename(t *testing.T) {
	t.Parallel()
	if !isHTMLFilename("index.html") || !isHTMLFilename("about.HTM") {
		t.Fatal("html filenames should be recognized")
	}
	if isHTMLFilename("style.css") {
		t.Fatal("css should not be treated as html entry")
	}
}

func TestFindHTMLEntry(t *testing.T) {
	t.Parallel()
	files := []*entity.PreviewFile{
		{Filename: "app.js", MimeType: bConst.PreviewMimeJS},
		{Filename: "index.html", MimeType: bConst.PreviewMimeHTML},
	}
	if got := findHTMLEntry(files); got != "index.html" {
		t.Fatalf("findHTMLEntry() = %q, want index.html", got)
	}
	if got := findHTMLEntry([]*entity.PreviewFile{{Filename: "app.js", MimeType: bConst.PreviewMimeJS}}); got != "" {
		t.Fatalf("findHTMLEntry(no html) = %q, want empty", got)
	}
}

func TestValidateSlug(t *testing.T) {
	t.Parallel()
	logic := &PagesLogic{}
	if xErr := logic.validateSlug(t.Context(), "design-system"); xErr != nil {
		t.Fatalf("valid slug rejected: %v", xErr)
	}
	for _, slug := range []string{"", "Bad_Slug", "has space", "UPPER", "a/b"} {
		if xErr := logic.validateSlug(t.Context(), slug); xErr == nil {
			t.Fatalf("invalid slug %q accepted", slug)
		}
	}
}

func TestValidateFilename(t *testing.T) {
	t.Parallel()
	if err := validateFilename("index.html"); err != nil {
		t.Fatalf("valid filename rejected: %v", err)
	}
	for _, name := range []string{"", "assets/app.js", "..", "foo\\bar", "a..b"} {
		if err := validateFilename(name); err == nil {
			t.Fatalf("invalid filename %q accepted", name)
		}
	}
}
