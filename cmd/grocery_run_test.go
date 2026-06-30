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
		case strings.HasSuffix(r.URL.Path, "/item1"):
			w.WriteHeader(http.StatusOK)
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

	out := captureStdout(func() { groceryListCmd.Run(groceryListCmd, nil) })
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

	out := captureStdout(func() { groceryCreateCmd.Run(groceryCreateCmd, nil) })
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected created list in output, got: %s", out)
	}
}

func TestGroceryOrganizeCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() { groceryOrganizeCmd.Run(groceryOrganizeCmd, nil) })
	if !strings.Contains(out, "organized successfully") {
		t.Errorf("expected organize confirmation, got: %s", out)
	}
}

func TestGroceryOrderCmd_WithURL(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() { groceryOrderCmd.Run(groceryOrderCmd, nil) })
	if !strings.Contains(out, "Order URL:") {
		t.Errorf("expected order URL in output, got: %s", out)
	}
}

func TestGroceryShowCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() { groceryShowCmd.Run(groceryShowCmd, nil) })
	if !strings.Contains(out, "Groceries") {
		t.Errorf("expected list in output, got: %s", out)
	}
}

func TestGroceryAddCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID, origItems := groceryListID, groceryItems
	groceryListID, groceryItems = "list1", []string{"Eggs", "Milk"}
	t.Cleanup(func() { groceryListID, groceryItems = origID, origItems })

	out := captureStdout(func() { groceryAddCmd.Run(groceryAddCmd, nil) })
	if !strings.Contains(out, "Added 2 item(s)") {
		t.Errorf("expected item count in output, got: %s", out)
	}
}

func TestGroceryAddRecipeCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryRecipeID
	groceryRecipeID = "recipe1"
	t.Cleanup(func() { groceryRecipeID = origID })

	out := captureStdout(func() { groceryAddRecipeCmd.Run(groceryAddRecipeCmd, nil) })
	if !strings.Contains(out, "added to grocery list successfully") {
		t.Errorf("expected confirmation, got: %s", out)
	}
}

func TestGroceryClearCmd(t *testing.T) {
	newCmdTestClient(t, groceryMockHandler())
	origID := groceryListID
	groceryListID = "list1"
	t.Cleanup(func() { groceryListID = origID })

	out := captureStdout(func() { groceryClearCmd.Run(groceryClearCmd, nil) })
	if !strings.Contains(out, "Cleared") {
		t.Errorf("expected clear confirmation, got: %s", out)
	}
}

func TestGroceryCmdExists(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Use == "grocery" {
			found = true
			break
		}
	}
	if !found {
		t.Error("grocery command not registered on root")
	}
}
