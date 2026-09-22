package main

import (
	"strings"
	"testing"
)

func mustParseVersion(t *testing.T, value string) version {
	t.Helper()
	parsed, err := parseVersion(value)
	if err != nil {
		t.Fatalf("parseVersion(%q) failed: %v", value, err)
	}
	return parsed
}

func assertAllowed(t *testing.T, previous, candidate string) {
	t.Helper()
	if err := validateTransition(mustParseVersion(t, previous), mustParseVersion(t, candidate)); err != nil {
		t.Fatalf("expected %s -> %s to pass: %v", previous, candidate, err)
	}
}

func assertRejected(t *testing.T, previous, candidate string) {
	t.Helper()
	if err := validateTransition(mustParseVersion(t, previous), mustParseVersion(t, candidate)); err == nil {
		t.Fatalf("expected %s -> %s to fail", previous, candidate)
	}
}

func TestConsecutiveStableVersions(t *testing.T) {
	assertAllowed(t, "v1.1.0", "v1.1.1")
	assertAllowed(t, "v1.1.9", "v1.2.0")
	assertAllowed(t, "v1.9.9", "v2.0.0")
}

func TestSkippedStableVersions(t *testing.T) {
	assertRejected(t, "v1.1.0", "v1.3.0")
	assertRejected(t, "v1.1.1", "v1.1.3")
	assertRejected(t, "v1.1.9", "v3.0.0")
	assertRejected(t, "v1.1.1", "v1.1.0")
}

func TestConsecutivePrereleaseVersions(t *testing.T) {
	assertAllowed(t, "v1.1.0-beta.20", "v1.1.0-beta.21")
	assertAllowed(t, "v1.1.0-beta.20", "v1.1.0")
	assertAllowed(t, "v1.1.0", "v1.2.0-beta.1")
}

func TestSkippedOrMismatchedPrereleaseVersions(t *testing.T) {
	assertRejected(t, "v1.1.0-beta.20", "v1.1.0-beta.25")
	assertRejected(t, "v1.1.0-beta.20", "v1.1.0-rc.21")
	assertRejected(t, "v1.1.0-beta.20", "v1.1.1-beta.1")
	assertRejected(t, "v1.1.0", "v1.2.0-beta.2")
}

func TestFindLatestIgnoresCandidateAndMalformedTags(t *testing.T) {
	candidate := mustParseVersion(t, "v1.0.0-beta.40")
	latest, err := findLatest(
		strings.NewReader("v1.0.0-beta.38\nv1.0.0-beta.39\ninvalid\nv1.0.0-beta.40\n"),
		candidate,
	)
	if err != nil {
		t.Fatalf("findLatest failed: %v", err)
	}
	if latest == nil || latest.raw != "v1.0.0-beta.39" {
		t.Fatalf("expected v1.0.0-beta.39, got %#v", latest)
	}
}

func TestParseVersionRejectsNonCanonicalVersion(t *testing.T) {
	for _, value := range []string{"1.1.0", "v1.1.0-beta", "v01.1.0"} {
		if _, err := parseVersion(value); err == nil {
			t.Fatalf("expected %q to fail", value)
		}
	}
}

func TestRunAllowsFirstRelease(t *testing.T) {
	var stdout strings.Builder
	var stderr strings.Builder
	if code := run([]string{"v1.0.0"}, strings.NewReader(""), &stdout, &stderr); code != 0 {
		t.Fatalf("expected first release to pass, code=%d stderr=%s", code, stderr.String())
	}
}
