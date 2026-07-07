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

	out := captureStdout(func() { listListCmd.Run(listListCmd, nil) })
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list in output, got: %s", out)
	}
}

func TestListGetCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID := listID
	listID = "list1"
	t.Cleanup(func() { listID = origID })

	out := captureStdout(func() { listGetCmd.Run(listGetCmd, nil) })
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list in output, got: %s", out)
	}
}

func TestListCreateCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origTitle := listTitle
	listTitle = "Groceries"
	t.Cleanup(func() { listTitle = origTitle })

	out := captureStdout(func() { listCreateCmd.Run(listCreateCmd, nil) })
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

	out := captureStdout(func() { listCreateCmd.Run(listCreateCmd, nil) })
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected created list in output, got: %s", out)
	}
}

func TestListDeleteCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID := listID
	listID = "list1"
	t.Cleanup(func() { listID = origID })

	out := captureStdout(func() { listDeleteCmd.Run(listDeleteCmd, nil) })
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

	out := captureStdout(func() { listDeleteCmd.Run(listDeleteCmd, nil) })
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestListAddItemCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origTitle := listID, listItemTitle
	listID, listItemTitle = "list1", "Eggs"
	t.Cleanup(func() { listID, listItemTitle = origID, origTitle })

	out := captureStdout(func() { listAddItemCmd.Run(listAddItemCmd, nil) })
	if !strings.Contains(out, "Eggs") {
		t.Errorf("expected added item in output, got: %s", out)
	}
}

func TestListDeleteItemCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origID, origItemID := listID, listItemID
	listID, listItemID = "list1", "item1"
	t.Cleanup(func() { listID, listItemID = origID, origItemID })

	out := captureStdout(func() { listDeleteItemCmd.Run(listDeleteItemCmd, nil) })
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

	out := captureStdout(func() { listDeleteItemCmd.Run(listDeleteItemCmd, nil) })
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

	out := captureStdout(func() { listUpdateCmd.Run(listUpdateCmd, nil) })
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

	out := captureStdout(func() { listUpdateItemCmd.Run(listUpdateItemCmd, nil) })
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated item in output, got: %s", out)
	}
}

func TestTaskBoxItemCreateCmd(t *testing.T) {
	newCmdTestClient(t, listMockHandler())
	origTitle := listItemTitle
	listItemTitle = "Pack bag"
	t.Cleanup(func() { listItemTitle = origTitle })

	out := captureStdout(func() { taskBoxItemCreateCmd.Run(taskBoxItemCreateCmd, nil) })
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

	out := captureStdout(func() { listClearCompletedCmd.Run(listClearCompletedCmd, nil) })
	if !strings.Contains(out, "Deleted 1 completed item") {
		t.Errorf("expected deletion count in output, got: %s", out)
	}
}

func TestListCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "list")
}
