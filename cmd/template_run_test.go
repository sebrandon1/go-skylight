package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func templateMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"c1","attributes":{"summary":"Dishes","status":"pending","reward_points":5}}]}`)
		case strings.HasSuffix(r.URL.Path, "/chores") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"c2","attributes":{"summary":"Dishes"}}}`)
		case strings.HasSuffix(r.URL.Path, "/rewards") && r.Method == http.MethodGet:
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":10}}]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":[{"id":"r2","attributes":{"name":"Ice cream","point_value":10}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen","timezone":"UTC"}}}`)
		}
	}
}

func writeTestTemplate(t *testing.T, name string) string {
	t.Helper()
	dir, err := templateDir()
	if err != nil {
		t.Fatalf("templateDir: %v", err)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	tmpl := frameTemplate{
		Name:      name,
		CreatedAt: time.Now().Format(time.RFC3339),
		Chores:    []choreTemplate{{Title: "Dishes", Points: 5}},
		Rewards:   []rewardTemplate{{Title: "Ice cream", Points: 10}},
	}
	data, _ := json.MarshalIndent(tmpl, "", "  ")
	path := filepath.Join(dir, name+".json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestTemplateSaveCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	newCmdTestClient(t, templateMockHandler())

	origName, origResources := templateName, templateResources
	templateName, templateResources = "test-weekly", nil
	t.Cleanup(func() { templateName, templateResources = origName, origResources })

	out := captureStdout(func() {
		if err := templateSaveCmd.RunE(templateSaveCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "test-weekly") {
		t.Errorf("expected template name in output, got: %s", out)
	}

	dir, _ := templateDir()
	if _, err := os.Stat(filepath.Join(dir, "test-weekly.json")); err != nil {
		t.Errorf("expected template file to exist: %v", err)
	}
}

func TestTemplateSaveCmd_ChoresOnly(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	newCmdTestClient(t, templateMockHandler())

	origName, origResources := templateName, templateResources
	templateName, templateResources = "chores-only", []string{"chores"}
	t.Cleanup(func() { templateName, templateResources = origName, origResources })

	out := captureStdout(func() {
		if err := templateSaveCmd.RunE(templateSaveCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "1 chore") {
		t.Errorf("expected chore count in output, got: %s", out)
	}
	if !strings.Contains(out, "0 reward") {
		t.Errorf("expected zero rewards in output, got: %s", out)
	}
}

func TestTemplateApplyCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	writeTestTemplate(t, "weekly")
	newCmdTestClient(t, templateMockHandler())

	origName, origResources, origDryRun := templateName, templateResources, templateDryRun
	templateName, templateResources, templateDryRun = "weekly", nil, false
	t.Cleanup(func() { templateName, templateResources, templateDryRun = origName, origResources, origDryRun })

	out := captureStdout(func() {
		if err := templateApplyCmd.RunE(templateApplyCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "weekly") {
		t.Errorf("expected template name in output, got: %s", out)
	}
}

func TestTemplateApplyCmd_DryRun(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	writeTestTemplate(t, "weekly")
	newCmdTestClient(t, templateMockHandler())

	origName, origDryRun := templateName, templateDryRun
	templateName, templateDryRun = "weekly", true
	t.Cleanup(func() { templateName, templateDryRun = origName, origDryRun })

	out := captureStdout(func() {
		if err := templateApplyCmd.RunE(templateApplyCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestTemplateApplyCmd_NotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	newCmdTestClient(t, templateMockHandler())

	origName := templateName
	templateName = "nonexistent"
	t.Cleanup(func() { templateName = origName })

	err := templateApplyCmd.RunE(templateApplyCmd, nil)
	if err == nil {
		t.Error("expected error for missing template, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestTemplateListCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	writeTestTemplate(t, "weekly")

	out := captureStdout(func() {
		if err := templateListCmd.RunE(templateListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "weekly") {
		t.Errorf("expected template name in output, got: %s", out)
	}
}

func TestTemplateListCmd_Empty(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	out := captureStdout(func() {
		if err := templateListCmd.RunE(templateListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No templates") {
		t.Errorf("expected empty-state message in output, got: %s", out)
	}
}

func TestTemplateDeleteCmd(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	path := writeTestTemplate(t, "old-template")

	origName, origYes := templateName, yes
	templateName, yes = "old-template", true
	t.Cleanup(func() { templateName, yes = origName, origYes })

	out := captureStdout(func() {
		if err := templateDeleteCmd.RunE(templateDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "old-template") {
		t.Errorf("expected template name in output, got: %s", out)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected template file to be deleted")
	}
}

func TestTemplateDeleteCmd_NotFound(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	origName, origYes := templateName, yes
	templateName, yes = "nonexistent", true
	t.Cleanup(func() { templateName, yes = origName, origYes })

	err := templateDeleteCmd.RunE(templateDeleteCmd, nil)
	if err == nil {
		t.Error("expected error for missing template, got nil")
	}
}

func TestTemplateCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "template")
}
