package adr_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"stellarbill-backend/internal/adr"
)

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	return path
}

func validADR(num, title, status string) string {
	return "# " + num + ". " + title + "\n\n" +
		"## Status\n\n" + status + "\n\n" +
		"## Context\n\nWhy.\n\n" +
		"## Decision\n\nWe will do X.\n\n" +
		"## Consequences\n\n### Positive\n\n- ok\n"
}

func TestParseFile_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := write(t, dir, "0001-example-decision.md", validADR("0001", "Example Decision", "Accepted"))
	rec, err := adr.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Number != 1 || rec.Slug != "example-decision" || rec.Title != "Example Decision" || rec.Status != "Accepted" {
		t.Fatalf("unexpected record: %+v", rec)
	}
}

func TestParseFile_RejectsBadFilename(t *testing.T) {
	dir := t.TempDir()
	path := write(t, dir, "1-bad.md", validADR("0001", "Bad", "Accepted"))
	if _, err := adr.ParseFile(path); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseFile_TitleNumberMismatch(t *testing.T) {
	dir := t.TempDir()
	path := write(t, dir, "0002-x.md", validADR("0001", "X", "Accepted"))
	if _, err := adr.ParseFile(path); err == nil {
		t.Fatal("expected title/filename mismatch error")
	}
}

func TestParseFile_MissingSection(t *testing.T) {
	dir := t.TempDir()
	body := "# 0003. No Decision\n\n## Status\n\nAccepted\n\n## Context\n\nctx\n\n## Consequences\n\n- x\n"
	path := write(t, dir, "0003-no-decision.md", body)
	if _, err := adr.ParseFile(path); err == nil || !strings.Contains(err.Error(), "Decision") {
		t.Fatalf("expected missing Decision error, got %v", err)
	}
}

func TestParseFile_InvalidStatus(t *testing.T) {
	dir := t.TempDir()
	path := write(t, dir, "0004-bad-status.md", validADR("0004", "Bad Status", "Maybe"))
	if _, err := adr.ParseFile(path); err == nil {
		t.Fatal("expected invalid status")
	}
}

func TestParseFile_SupersededStatus(t *testing.T) {
	dir := t.TempDir()
	body := validADR("0005", "Old", "Superseded by [ADR-0006](0006-new.md)")
	path := write(t, dir, "0005-old.md", body)
	rec, err := adr.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != "Superseded" {
		t.Fatalf("got %q", rec.Status)
	}
}

func TestValidateUniqueNumbers_Duplicate(t *testing.T) {
	recs := []adr.Record{
		{Number: 1, Path: "a.md"},
		{Number: 1, Path: "b.md"},
	}
	if err := adr.ValidateUniqueNumbers(recs); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestValidateUniqueNumbers_OK(t *testing.T) {
	recs := []adr.Record{{Number: 0, Path: "0000-template.md"}, {Number: 1, Path: "0001-a.md"}}
	if err := adr.ValidateUniqueNumbers(recs); err != nil {
		t.Fatal(err)
	}
}

func TestValidateTemplatePresent(t *testing.T) {
	if err := adr.ValidateTemplatePresent([]adr.Record{{Number: 1}}); err == nil {
		t.Fatal("expected missing template")
	}
	if err := adr.ValidateTemplatePresent([]adr.Record{{Number: 0, IsTemplate: true}}); err != nil {
		t.Fatal(err)
	}
}

func TestLoadDirAndGenerateIndex(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "0000-template.md", validADR("0000", "Title of the Decision", "Proposed"))
	write(t, dir, "0001-alpha.md", validADR("0001", "Alpha", "Accepted"))
	write(t, dir, "0002-beta.md", validADR("0002", "Beta", "Proposed"))
	write(t, dir, "README.md", "stale\n")
	write(t, dir, "ADR_TOOLS.md", "notes\n")

	recs, err := adr.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 3 {
		t.Fatalf("want 3 records, got %d", len(recs))
	}
	idx := adr.GenerateIndex(recs)
	if !strings.Contains(idx, "[0001](0001-alpha.md)") || !strings.Contains(idx, "Alpha") {
		t.Fatalf("index missing alpha: %s", idx)
	}
	if strings.Contains(idx, "0000-template.md") && strings.Contains(idx, "| [0000]") {
		t.Fatal("template should not appear in decision table rows")
	}
	if !strings.Contains(idx, "0000-template.md") {
		t.Fatal("template should be linked in Template section")
	}
}

func TestReadAdrDir(t *testing.T) {
	root := t.TempDir()
	got, err := adr.ReadAdrDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(root, adr.DefaultDir) {
		t.Fatalf("default dir: %s", got)
	}

	if err := os.WriteFile(filepath.Join(root, adr.AdrDirFileName), []byte("docs/adr\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = adr.ReadAdrDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(root, "docs/adr") {
		t.Fatalf("configured dir: %s", got)
	}
}

func TestLintAndWriteIndex(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "adr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".adr-dir"), []byte("docs/adr\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "0000-template.md", validADR("0000", "Title of the Decision", "Proposed"))
	write(t, dir, "0001-one.md", validADR("0001", "One", "Accepted"))

	if _, err := adr.Lint(root, true); err == nil {
		t.Fatal("expected stale/missing index error")
	}

	if err := adr.WriteIndex(root); err != nil {
		t.Fatal(err)
	}
	recs, err := adr.Lint(root, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("got %d", len(recs))
	}
}

func TestLint_DuplicateNumbers(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "adr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "0000-template.md", validADR("0000", "Title of the Decision", "Proposed"))
	write(t, dir, "0001-one.md", validADR("0001", "One", "Accepted"))
	// Second file with same number but different slug — ValidateUniqueNumbers catches it
	// after ParseFile; fabricate by writing another 0001-*
	write(t, dir, "0001-two.md", validADR("0001", "Two", "Accepted"))
	if err := adr.WriteIndex(root); err == nil {
		// WriteIndex also validates unique numbers
		t.Fatal("expected duplicate error from WriteIndex")
	}
	if _, err := adr.Lint(root, false); err == nil {
		t.Fatal("expected duplicate error from Lint")
	}
}

func TestReadAdrDir_EmptyAndAbsolute(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, adr.AdrDirFileName), []byte("   \n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := adr.ReadAdrDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Join(root, adr.DefaultDir) {
		t.Fatalf("empty config should fall back, got %s", got)
	}

	abs := filepath.Join(root, "custom-adr")
	if err := os.MkdirAll(abs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, adr.AdrDirFileName), []byte(abs+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = adr.ReadAdrDir(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != abs {
		t.Fatalf("want abs %s got %s", abs, got)
	}
}

func TestLoadDir_ParseErrorAndMissingDir(t *testing.T) {
	if _, err := adr.LoadDir(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("expected missing dir error")
	}

	dir := t.TempDir()
	write(t, dir, "0000-template.md", validADR("0000", "Title of the Decision", "Proposed"))
	write(t, dir, "0001-bad.md", "# 0001. Bad\n\n## Status\n\nNope\n\n## Context\n\nx\n\n## Decision\n\ny\n\n## Consequences\n\nz\n")
	if _, err := adr.LoadDir(dir); err == nil {
		t.Fatal("expected parse error from LoadDir")
	}
}

func TestLint_MissingTemplateAndReadErrors(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "adr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "0001-only.md", validADR("0001", "Only", "Accepted"))
	if _, err := adr.Lint(root, false); err == nil {
		t.Fatal("expected missing template")
	}

	// Point .adr-dir at a missing path
	if err := os.WriteFile(filepath.Join(root, ".adr-dir"), []byte("docs/missing-adr\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := adr.Lint(root, false); err == nil {
		t.Fatal("expected load error for missing adr dir")
	}
	if err := adr.WriteIndex(root); err == nil {
		t.Fatal("expected WriteIndex load error")
	}
}

func TestLint_StaleIndexMessage(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "docs", "adr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, dir, "0000-template.md", validADR("0000", "Title of the Decision", "Proposed"))
	write(t, dir, "0001-one.md", validADR("0001", "One", "Accepted"))
	write(t, dir, "README.md", "# wrong\n")
	if _, err := adr.Lint(root, true); err == nil || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("expected stale index error, got %v", err)
	}
}

func TestParseFile_ReadError(t *testing.T) {
	if _, err := adr.ParseFile(filepath.Join(t.TempDir(), "0008-missing.md")); err == nil {
		t.Fatal("expected read error")
	}
}

func TestLoadDir_SkipsSubdirectories(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "0000-template.md", validADR("0000", "Title of the Decision", "Proposed"))
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(dir, "nested"), "0001-hidden.md", validADR("0001", "Hidden", "Accepted"))
	recs, err := adr.LoadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("want only template, got %d", len(recs))
	}
}

func TestReadAdrDir_UnreadableConfig(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, adr.AdrDirFileName), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := adr.ReadAdrDir(root); err == nil {
		t.Fatal("expected read error for .adr-dir directory")
	}
	if _, err := adr.Lint(root, false); err == nil {
		t.Fatal("expected Lint to fail when .adr-dir unreadable")
	}
	if err := adr.WriteIndex(root); err == nil {
		t.Fatal("expected WriteIndex to fail when .adr-dir unreadable")
	}
}

func TestParseFile_MissingTitle(t *testing.T) {
	dir := t.TempDir()
	body := "## Status\n\nAccepted\n\n## Context\n\nx\n\n## Decision\n\ny\n\n## Consequences\n\nz\n"
	path := write(t, dir, "0006-no-title.md", body)
	if _, err := adr.ParseFile(path); err == nil {
		t.Fatal("expected missing title")
	}
}

func TestParseFile_EmptyStatus(t *testing.T) {
	dir := t.TempDir()
	body := "# 0007. Empty Status\n\n## Status\n\n\n## Context\n\nx\n\n## Decision\n\ny\n\n## Consequences\n\nz\n"
	path := write(t, dir, "0007-empty-status.md", body)
	if _, err := adr.ParseFile(path); err == nil {
		t.Fatal("expected empty status error")
	}
}

func TestRepoADRs_Lint(t *testing.T) {
	root := findRepoRoot(t)
	recs, err := adr.Lint(root, true)
	if err != nil {
		t.Fatalf("repo ADR lint failed: %v", err)
	}
	decisions := 0
	for _, r := range recs {
		if !r.IsTemplate {
			decisions++
		}
	}
	if decisions < 10 {
		t.Fatalf("expected at least 10 backfilled ADRs, got %d", decisions)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, ".adr-dir")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Skip("repo root with .adr-dir not found (package tested outside module tree)")
	return ""
}
