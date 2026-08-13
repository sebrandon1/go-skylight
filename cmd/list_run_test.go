package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func listMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/sections/") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/item1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"item1","type":"list_item","attributes":{"label":"Updated","status":"pending","position":0}}}`)
		case strings.HasSuffix(r.URL.Path, "/item1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/list_items") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"item1","type":"list_item","attributes":{"label":"Eggs","status":"pending","position":0}}}`)
		case strings.HasSuffix(r.URL.Path, "/list1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"list1","type":"list","attributes":{"label":"Updated","color":"#2178AF","kind":"to_do"}}}`)
		case strings.HasSuffix(r.URL.Path, "/list1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/list1"):
			fmt.Fprint(w, `{"data":{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"to_do"}}}`)
		case strings.HasSuffix(r.URL.Path, "/lists") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"to_do"}}}`)
		case strings.HasSuffix(r.URL.Path, "/task_box_items"):
			fmt.Fprint(w, `{"id":"t1","title":"Pack bag"}`)
		default:
			fmt.Fprint(w, `{"data":[{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"to_do"}}]}`)
		}
	}
}

func TestListListCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())

	out := captureStdout(func() {
		if err := listListCmd.RunE(listListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list in output, got: %s", out)
	}
}

func TestListGetCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID := listID
	listID = "list1"
	t.Cleanup(func() { listID = origID })

	out := captureStdout(func() {
		if err := listGetCmd.RunE(listGetCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list in output, got: %s", out)
	}
}

func TestListCreateCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origTitle := listTitle
	listTitle = "Groceries"
	t.Cleanup(func() { listTitle = origTitle })

	out := captureStdout(func() {
		if err := listCreateCmd.RunE(listCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected created list in output, got: %s", out)
	}
}

func TestListCreateCmd_HideFromFrame(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origTitle := listTitle
	listTitle = "Groceries"
	t.Cleanup(func() { listTitle = origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := listCreateCmd.Flags().Set("hide-from-frame", "true"); err != nil {
		t.Fatalf("setting hide-from-frame flag: %v", err)
	}

	out := captureStdout(func() {
		if err := listCreateCmd.RunE(listCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected created list in output, got: %s", out)
	}
}

func TestListDeleteCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origYes := listID, yes
	listID, yes = "list1", true
	t.Cleanup(func() { listID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := listDeleteCmd.RunE(listDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestListDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := listID, dryRun
	listID, dryRun = "list1", true
	t.Cleanup(func() { listID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := listDeleteCmd.RunE(listDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestListAddItemCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origTitle := listID, listItemTitle
	listID, listItemTitle = "list1", "Eggs"
	t.Cleanup(func() { listID, listItemTitle = origID, origTitle })

	out := captureStdout(func() {
		if err := listAddItemCmd.RunE(listAddItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Eggs") {
		t.Errorf("expected added item in output, got: %s", out)
	}
}

func TestListDeleteItemCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origItemID, origYes := listID, listItemID, yes
	listID, listItemID, yes = "list1", "item1", true
	t.Cleanup(func() { listID, listItemID, yes = origID, origItemID, origYes })

	out := captureStdout(func() {
		if err := listDeleteItemCmd.RunE(listDeleteItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestListDeleteItemCmd_DryRun(t *testing.T) {
	origID, origItemID, origDryRun := listID, listItemID, dryRun
	listID, listItemID, dryRun = "list1", "item1", true
	t.Cleanup(func() { listID, listItemID, dryRun = origID, origItemID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := listDeleteItemCmd.RunE(listDeleteItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestListUpdateCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origTitle := listID, listTitle
	listID, listTitle = "list1", "Updated"
	t.Cleanup(func() { listID, listTitle = origID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := listUpdateCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := listUpdateCmd.RunE(listUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated list in output, got: %s", out)
	}
}

func TestListUpdateItemCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origItemID, origTitle := listID, listItemID, listItemTitle
	listID, listItemID, listItemTitle = "list1", "item1", "Updated"
	t.Cleanup(func() { listID, listItemID, listItemTitle = origID, origItemID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := listUpdateItemCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := listUpdateItemCmd.RunE(listUpdateItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated item in output, got: %s", out)
	}
}

func TestTaskBoxItemCreateCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origTitle := listItemTitle
	listItemTitle = "Pack bag"
	t.Cleanup(func() { listItemTitle = origTitle })

	out := captureStdout(func() {
		if err := taskBoxItemCreateCmd.RunE(taskBoxItemCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Pack bag") {
		t.Errorf("expected created task box item in output, got: %s", out)
	}
}

func TestListClearCompletedCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/item1"):
			w.WriteHeader(http.StatusOK)
		default:
			fmt.Fprint(w, `{"data":{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"to_do"}},"included":[{"id":"item1","type":"list_item","attributes":{"label":"Milk","status":"completed","position":0}}]}`)
		}
	})
	origID := listID
	listID = "list1"
	t.Cleanup(func() { listID = origID })

	out := captureStdout(func() {
		if err := listClearCompletedCmd.RunE(listClearCompletedCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Deleted 1 completed item") {
		t.Errorf("expected deletion count in output, got: %s", out)
	}
}

func TestListCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "list")
}

func TestListReorderItemCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origItemID, origPos := listID, listItemID, listItemPosition
	listID, listItemID, listItemPosition = "list1", "item1", 2
	t.Cleanup(func() { listID, listItemID, listItemPosition = origID, origItemID, origPos })

	out := captureStdout(func() {
		if err := listReorderItemCmd.RunE(listReorderItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "item1") {
		t.Errorf("expected item in output, got: %s", out)
	}
}

func TestListDeleteSectionCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origSectionID, origYes := listID, listSectionID, yes
	listID, listSectionID, yes = "list1", "sec1", true
	t.Cleanup(func() { listID, listSectionID, yes = origID, origSectionID, origYes })

	out := captureStdout(func() {
		if err := listDeleteSectionCmd.RunE(listDeleteSectionCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted") {
		t.Errorf("expected success message in output, got: %s", out)
	}
}

func TestListDeleteSectionCmd_DryRun(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origSectionID, origDryRun := listID, listSectionID, dryRun
	listID, listSectionID, dryRun = "list1", "sec1", true
	t.Cleanup(func() { listID, listSectionID, dryRun = origID, origSectionID, origDryRun })

	out := captureStdout(func() {
		if err := listDeleteSectionCmd.RunE(listDeleteSectionCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry-run indicator in output, got: %s", out)
	}
}

func TestListCreateCmd_TableOutput(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origFmt, origTitle := outputFormat, listTitle
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat, listTitle = origFmt, origTitle })

	out := captureStdout(func() {
		if err := listCreateCmd.RunE(listCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "TITLE") {
		t.Errorf("expected table header in output, got: %s", out)
	}
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list title in table, got: %s", out)
	}
}
