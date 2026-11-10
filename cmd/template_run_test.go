package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
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
		Rewards:   []rewardTemplate{{Title: "Ice cream", Points: 10, CategoryIDs: []int{12}}},
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

func setTemplateFlags(t *testing.T, name, startDate string, resources []string) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	origName, origStart, origResources, origDry := templateName, templateStartDate, templateResources, templateDryRun
	templateName, templateStartDate, templateResources, templateDryRun = name, startDate, resources, false
	t.Cleanup(func() {
		templateName, templateStartDate, templateResources, templateDryRun = origName, origStart, origResources, origDry
	})
}

func readTestTemplate(t *testing.T, name string) frameTemplate {
	t.Helper()
	dir, err := templateDir()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		t.Fatal(err)
	}
	var tmpl frameTemplate
	if err := json.Unmarshal(raw, &tmpl); err != nil {
		t.Fatal(err)
	}
	return tmpl
}

func writeRawTemplate(t *testing.T, name, content string) {
	t.Helper()
	dir, err := templateDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".json"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateSaveCmd_CollapsesSeriesAndKeepsCategory(t *testing.T) {
	setTemplateFlags(t, "weekly", "", nil)
	var upForGrabsAsked bool
	var after string
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/chores"):
			after = r.URL.Query().Get("after")
			if r.URL.Query().Get("include_up_for_grabs") == "true" {
				upForGrabsAsked = true
				fmt.Fprint(w, `{"data":[]}`)
				return
			}
			due := time.Now().AddDate(0, 0, 1).Format(lib.DateFormat)
			row := `{"id":"7-%d","attributes":{"summary":"Dishes","status":"pending","start":%q,"recurrence_set":["RRULE:FREQ=DAILY"]}}`
			fmt.Fprintf(w, `{"data":[`+row+`,`+row+`,`+row+`]}`, 1, due, 2, due, 3, due)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"Ice cream","point_value":10},"relationships":{"category":{"data":{"id":"12","type":"category"}}}}]}`)
		}
	})

	captureStdout(func() {
		if err := templateSaveCmd.RunE(templateSaveCmd, nil); err != nil {
			t.Errorf("save: %v", err)
		}
	})
	if !upForGrabsAsked || after == "" {
		t.Errorf("expected an up-for-grabs query with a date window, asked=%v after=%q", upForGrabsAsked, after)
	}
	tmpl := readTestTemplate(t, "weekly")
	if len(tmpl.Chores) != 1 || len(tmpl.Chores[0].RecurrenceSet) != 1 {
		t.Fatalf("chores = %+v, want one series with its rule", tmpl.Chores)
	}
	if len(tmpl.Rewards) != 1 || len(tmpl.Rewards[0].CategoryIDs) != 1 || tmpl.Rewards[0].CategoryIDs[0] != 12 {
		t.Errorf("rewards = %+v, want category 12", tmpl.Rewards)
	}
}

// templateApplyRecorder counts POSTs by path and records their bodies.
func templateApplyRecorder(t *testing.T) (posts *[]string, bodies *[]map[string]any) {
	t.Helper()
	posts, bodies = &[]string{}, &[]map[string]any{}
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		*posts = append(*posts, r.URL.Path)
		*bodies = append(*bodies, body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		switch {
		case strings.Contains(r.URL.Path, "create_multiple"):
			fmt.Fprint(w, `{"data":[{"id":"c9","attributes":{"summary":"x"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/rewards"):
			fmt.Fprint(w, `{"data":[{"id":"r9","attributes":{"name":"x"}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"c9","attributes":{"summary":"x"}}}`)
		}
	})
	return posts, bodies
}

func TestTemplateApplyCmd_SendsRuleAndCategory(t *testing.T) {
	setTemplateFlags(t, "weekly", "2026-12-01", nil)
	writeRawTemplate(t, "weekly", `{"name":"weekly","chores":[{"title":"Dishes","due_date":"2026-11-01","recurrence_set":["RRULE:FREQ=WEEKLY"],"recurring_until":"2027-01-01"}],"rewards":[{"title":"Ice cream","points":5,"category_ids":[12]}]}`)
	posts, bodies := templateApplyRecorder(t)

	captureStdout(func() {
		if err := templateApplyCmd.RunE(templateApplyCmd, nil); err != nil {
			t.Errorf("apply: %v", err)
		}
	})
	if len(*posts) != 2 {
		t.Fatalf("POSTs = %v, want one chore and one reward", *posts)
	}
	chore, reward := (*bodies)[0], (*bodies)[1]
	if rules, ok := chore["recurrence_set"].([]any); !ok || len(rules) != 1 || chore["recurring_until"] != "2027-01-01" {
		t.Errorf("chore body = %v, want the rule and its end date", chore)
	}
	if chore["start"] != "2026-12-01" {
		t.Errorf("start = %v, want the --start-date value", chore["start"])
	}
	if ids, ok := reward["category_ids"].([]any); !ok || len(ids) != 1 || ids[0] != float64(12) {
		t.Errorf("reward body = %v, want category 12", reward)
	}
}

func TestTemplateApplyCmd_UpForGrabsEndpoint(t *testing.T) {
	setTemplateFlags(t, "grabs", "", []string{exportResourceChores})
	writeRawTemplate(t, "grabs", `{"name":"grabs","chores":[{"title":"Sweep","due_date":"2026-11-01","up_for_grabs":true}]}`)
	posts, _ := templateApplyRecorder(t)

	captureStdout(func() {
		if err := templateApplyCmd.RunE(templateApplyCmd, nil); err != nil {
			t.Errorf("apply: %v", err)
		}
	})
	if len(*posts) != 1 || !strings.Contains((*posts)[0], "create_multiple") {
		t.Errorf("POSTs = %v, want one create_multiple call", *posts)
	}
}

func TestTemplateApplyCmd_RewardWithoutCategoryStopsBeforeWrites(t *testing.T) {
	setTemplateFlags(t, "old", "", nil)
	writeRawTemplate(t, "old", `{"name":"old","chores":[{"title":"Dishes"}],"rewards":[{"title":"Ice cream","points":5}]}`)
	posts, _ := templateApplyRecorder(t)

	err := templateApplyCmd.RunE(templateApplyCmd, nil)
	if err == nil || !strings.Contains(err.Error(), `"Ice cream"`) {
		t.Fatalf("error = %v, want one naming the reward", err)
	}
	if len(*posts) != 0 {
		t.Errorf("POSTs = %v, want none", *posts)
	}
}

func TestTemplateApplyCmd_OldFormatChoresStillApply(t *testing.T) {
	setTemplateFlags(t, "old", "", []string{exportResourceChores})
	writeRawTemplate(t, "old", `{"name":"old","chores":[{"title":"Dishes","recurring":true}],"rewards":[{"title":"Ice cream"}]}`)
	posts, bodies := templateApplyRecorder(t)

	captureStdout(func() {
		if err := templateApplyCmd.RunE(templateApplyCmd, nil); err != nil {
			t.Errorf("apply: %v", err)
		}
	})
	if len(*posts) != 1 || (*bodies)[0]["recurring"] != true {
		t.Errorf("POSTs = %v bodies = %v", *posts, *bodies)
	}
}

func TestTemplateApplyCmd_FailureReportsProgress(t *testing.T) {
	setTemplateFlags(t, "weekly", "", nil)
	writeRawTemplate(t, "weekly", `{"name":"weekly","chores":[{"title":"Dishes"}],"rewards":[{"title":"Ice cream","category_ids":[12]}]}`)
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/rewards") {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"data":{"id":"c9","attributes":{"summary":"x"}}}`)
	})

	err := templateApplyCmd.RunE(templateApplyCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "after creating 1 chore(s) and 0 reward(s)") {
		t.Errorf("error = %v, want the created counts", err)
	}
}

func TestTemplateSaveCmd_SkipsRoutinesWithNotice(t *testing.T) {
	setTemplateFlags(t, "weekly", "", []string{exportResourceChores})
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"9","attributes":{"summary":"Brush teeth","routine":true,"recurrence_set":["RRULE:FREQ=DAILY;BYHOUR=6"]}}]}`)
	})

	var stderr string
	captureStdout(func() {
		stderr = captureStderr(func() {
			if err := templateSaveCmd.RunE(templateSaveCmd, nil); err != nil {
				t.Errorf("save: %v", err)
			}
		})
	})
	if !strings.Contains(stderr, `"Brush teeth"`) {
		t.Errorf("expected a notice naming the routine, got: %q", stderr)
	}
	if tmpl := readTestTemplate(t, "weekly"); len(tmpl.Chores) != 0 {
		t.Errorf("routine saved as a chore: %+v", tmpl.Chores)
	}
}

func TestListChoresWithUpForGrabs_Errors(t *testing.T) {
	tests := []struct {
		name     string
		failGrab bool
	}{
		{"plain query fails", false},
		{"up-for-grabs query fails", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
				isGrab := r.URL.Query().Get("include_up_for_grabs") == "true"
				if isGrab == tt.failGrab {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				fmt.Fprint(w, `{"data":[]}`)
			})
			opts := lib.ChoreListOptions{After: "2026-10-01", Before: "2026-10-02"}
			if _, err := listChoresWithUpForGrabs(context.Background(), client, opts); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestTemplateApplyCmd_ChoreFailureReportsProgress(t *testing.T) {
	setTemplateFlags(t, "weekly", "", []string{exportResourceChores})
	writeRawTemplate(t, "weekly", `{"name":"weekly","chores":[{"title":"Dishes"}]}`)
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := templateApplyCmd.RunE(templateApplyCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "after creating 0 chore(s)") {
		t.Errorf("error = %v, want the created count", err)
	}
}
