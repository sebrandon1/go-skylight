package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func groceryMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/organize"):
			fmt.Fprint(w, `{}`)
		case strings.HasSuffix(r.URL.Path, "/order"):
			fmt.Fprint(w, `{"redirect_url":"https://instacart.example/order/1"}`)
		case strings.HasSuffix(r.URL.Path, "/recipe1/add_to_grocery_list"):
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/list_items") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"item1","type":"list_item","attributes":{"label":"Eggs","status":"pending","position":0}}}`)
		case strings.HasSuffix(r.URL.Path, "/item1") && r.Method == http.MethodPut:
			fmt.Fprint(w, `{"data":{"id":"item1","type":"list_item","attributes":{"label":"Updated Item","status":"pending","position":0}}}`)
		case strings.HasSuffix(r.URL.Path, "/item1"):
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/list1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/lists") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"grocery"}}}`)
		case strings.HasSuffix(r.URL.Path, "/lists"):
			fmt.Fprint(w, `{"data":[{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"grocery"}},{"id":"list2","type":"list","attributes":{"label":"To Do","kind":"to_do"}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"list1","type":"list","attributes":{"label":"Groceries","kind":"grocery"}},"included":[{"id":"item1","type":"list_item","attributes":{"label":"Milk","status":"completed","position":0}}]}`)
		}
	}
}

func TestGroceryListCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())

	out := captureStdout(func() {
		if err := groceryListCmd.RunE(groceryListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected grocery list in output, got: %s", out)
	}
	if strings.Contains(out, "To Do") {
		t.Errorf("expected non-grocery list filtered out, got: %s", out)
	}
}

func TestGroceryCreateCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origTitle := groceryTitle
	groceryTitle = "Groceries"
	t.Cleanup(func() { groceryTitle = origTitle })

	out := captureStdout(func() {
		if err := groceryCreateCmd.RunE(groceryCreateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected created list in output, got: %s", out)
	}
}

func TestGroceryOrganizeCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() {
		if err := groceryOrganizeCmd.RunE(groceryOrganizeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "organized successfully") {
		t.Errorf("expected organize confirmation, got: %s", out)
	}
}

func TestGroceryOrderCmd_WithURL(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() {
		if err := groceryOrderCmd.RunE(groceryOrderCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Order URL:") {
		t.Errorf("expected order URL in output, got: %s", out)
	}
}

func TestGroceryShowCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() {
		if err := groceryShowCmd.RunE(groceryShowCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list in output, got: %s", out)
	}
}

func TestGroceryAddCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID, origItems := groceryListID, groceryItems
	groceryListID, groceryItems = "list1", []string{"Eggs", "Milk"}
	t.Cleanup(func() { groceryListID, groceryItems = origID, origItems })

	out := captureStdout(func() {
		if err := groceryAddCmd.RunE(groceryAddCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Added 2 item") {
		t.Errorf("expected success count in output, got: %s", out)
	}
}

func TestGroceryAddRecipeCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryRecipeID
	groceryRecipeID = "recipe1"
	t.Cleanup(func() { groceryRecipeID = origID })

	out := captureStdout(func() {
		if err := groceryAddRecipeCmd.RunE(groceryAddRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "added to grocery list successfully") {
		t.Errorf("expected confirmation, got: %s", out)
	}
}

func TestGroceryClearCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() {
		if err := groceryClearCmd.RunE(groceryClearCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Cleared") {
		t.Errorf("expected clear confirmation, got: %s", out)
	}
}

func TestGroceryDeleteCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID, origYes := groceryListID, yes
	groceryListID, yes = "list1", true
	t.Cleanup(func() { groceryListID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := groceryDeleteCmd.RunE(groceryDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestGroceryDeleteCmd_DryRun(t *testing.T) {
	origID, origDryRun := groceryListID, dryRun
	groceryListID, dryRun = "list1", true
	t.Cleanup(func() { groceryListID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := groceryDeleteCmd.RunE(groceryDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestGroceryDeleteItemCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origListID, origItemID, origYes := groceryListID, groceryItemID, yes
	groceryListID, groceryItemID, yes = "list1", "item1", true
	t.Cleanup(func() { groceryListID, groceryItemID, yes = origListID, origItemID, origYes })

	out := captureStdout(func() {
		if err := groceryDeleteItemCmd.RunE(groceryDeleteItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestGroceryUpdateItemCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origListID, origItemID := groceryListID, groceryItemID
	groceryListID, groceryItemID = "list1", "item1"
	t.Cleanup(func() { groceryListID, groceryItemID = origListID, origItemID })

	if err := groceryUpdateItemCmd.Flags().Set("title", "Updated Item"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := groceryUpdateItemCmd.RunE(groceryUpdateItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated Item") {
		t.Errorf("expected updated item in output, got: %s", out)
	}
}

func TestGroceryDeleteCmd_ConfirmationDeclined(t *testing.T) {
	origID, origYes := groceryListID, yes
	groceryListID, yes = "list1", false
	t.Cleanup(func() { groceryListID, yes = origID, origYes })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	mockStdin(t, "n\n")

	out := captureStdout(func() {
		if err := groceryDeleteCmd.RunE(groceryDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "deleted successfully") {
		t.Errorf("expected no deletion when confirmation declined, got: %s", out)
	}
}

func TestGroceryDeleteItemCmd_DryRun(t *testing.T) {
	origListID, origItemID, origDryRun := groceryListID, groceryItemID, dryRun
	groceryListID, groceryItemID, dryRun = "list1", "item1", true
	t.Cleanup(func() { groceryListID, groceryItemID, dryRun = origListID, origItemID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := groceryDeleteItemCmd.RunE(groceryDeleteItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestGroceryDeleteItemCmd_ConfirmationDeclined(t *testing.T) {
	origListID, origItemID, origYes := groceryListID, groceryItemID, yes
	groceryListID, groceryItemID, yes = "list1", "item1", false
	t.Cleanup(func() { groceryListID, groceryItemID, yes = origListID, origItemID, origYes })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	mockStdin(t, "n\n")

	out := captureStdout(func() {
		if err := groceryDeleteItemCmd.RunE(groceryDeleteItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "deleted successfully") {
		t.Errorf("expected no deletion when confirmation declined, got: %s", out)
	}
}

func TestGroceryUpdateItemCmd_Completed(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origListID, origItemID := groceryListID, groceryItemID
	groceryListID, groceryItemID = "list1", "item1"
	t.Cleanup(func() { groceryListID, groceryItemID = origListID, origItemID })

	// pflag.Set() marks flags as permanently Changed on the shared singleton.
	if err := groceryUpdateItemCmd.Flags().Set("completed", "true"); err != nil {
		t.Fatalf("setting completed flag: %v", err)
	}

	out := captureStdout(func() {
		if err := groceryUpdateItemCmd.RunE(groceryUpdateItemCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Error("expected non-empty output after update")
	}
}

func TestGroceryCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "grocery")
}
