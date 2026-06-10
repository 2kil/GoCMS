package handlers

import (
	"path/filepath"
	"testing"

	"shuaitesteel.com/cms/config"
)

func TestIsValidSlug(t *testing.T) {
	valid := []string{"post-1", "column_2", "A123"}
	for _, slug := range valid {
		if !IsValidSlug(slug) {
			t.Fatalf("expected slug %q to be valid", slug)
		}
	}

	invalid := []string{"", "-post", "_post", "../post", `..\post`, "post/name", "post name"}
	for _, slug := range invalid {
		if IsValidSlug(slug) {
			t.Fatalf("expected slug %q to be invalid", slug)
		}
	}
}

func TestSafePublicFilePathRejectsTraversal(t *testing.T) {
	staticDir := t.TempDir()
	config.Cfg = &config.Config{Static: config.StaticConfig{Enable: true, Dir: staticDir}}

	cases := []string{
		"/../config.ini",
		`/..\config.ini`,
		`/post/..\config.ini`,
	}
	for _, requestPath := range cases {
		if _, ok := safePublicFilePath(requestPath); ok {
			t.Fatalf("expected path %q to be rejected", requestPath)
		}
	}
}

func TestSafePublicFilePathAllowsStaticFiles(t *testing.T) {
	staticDir := t.TempDir()
	config.Cfg = &config.Config{Static: config.StaticConfig{Enable: true, Dir: staticDir}}

	got, ok := safePublicFilePath("/post/demo.html")
	if !ok {
		t.Fatal("expected static path to be accepted")
	}
	want := filepath.Join(staticDir, "post", "demo.html")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
