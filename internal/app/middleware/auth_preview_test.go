package middleware

import "testing"

func TestPreviewShellAllowed(t *testing.T) {
	cases := []struct {
		name          string
		accessValid   bool
		refreshCookie string
		document      bool
		want          bool
	}{
		{name: "valid access", accessValid: true, document: true, want: true},
		{name: "document with refresh cookie only", refreshCookie: "rt", document: true, want: true},
		{name: "document with neither", document: true, want: false},
		{name: "subresource ignores refresh cookie", refreshCookie: "rt", document: false, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := previewShellAllowed(tc.accessValid, tc.refreshCookie, tc.document)
			if got != tc.want {
				t.Fatalf("previewShellAllowed() = %v, want %v", got, tc.want)
			}
		})
	}
}
