package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected JSON output, got empty string")
	}

	if !strings.Contains(output, "\"key\": \"value\"") {
		t.Errorf("Expected pretty-printed JSON with key/value, got: %s", output)
	}
}

func TestPrintJSONWithStruct(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	type testStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	data := testStruct{Name: "test", Age: 30}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"name\": \"test\"") {
		t.Errorf("Expected name field in output, got: %s", output)
	}
	if !strings.Contains(output, "\"age\": 30") {
		t.Errorf("Expected age field in output, got: %s", output)
	}
}

func TestPrintJSONWithSlice(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := []string{"a", "b", "c"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output == "" {
		t.Error("Expected JSON output for slice")
	}
	if !strings.Contains(output, "\"a\"") {
		t.Errorf("Expected element 'a' in output, got: %s", output)
	}
	if !strings.Contains(output, "\"b\"") {
		t.Errorf("Expected element 'b' in output, got: %s", output)
	}
	if !strings.Contains(output, "\"c\"") {
		t.Errorf("Expected element 'c' in output, got: %s", output)
	}
}

func TestPrintJSONWithNestedStruct(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	type inner struct {
		Value string `json:"value"`
	}
	type outer struct {
		Inner inner `json:"inner"`
	}

	data := outer{Inner: inner{Value: "nested"}}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"value\": \"nested\"") {
		t.Errorf("Expected nested value in output, got: %s", output)
	}
	if !strings.Contains(output, "\"inner\"") {
		t.Errorf("Expected inner key in output, got: %s", output)
	}
}

func TestPrintJSONWithEmptyMap(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "{}") {
		t.Errorf("Expected empty JSON object, got: %s", output)
	}
}

func TestPrintJSONWithEmptySlice(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := []string{}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "[]") {
		t.Errorf("Expected empty JSON array, got: %s", output)
	}
}

func TestPrintJSONWithNumbers(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]int{"count": 42}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"count\": 42") {
		t.Errorf("Expected count field with integer value, got: %s", output)
	}
}

func TestPrintJSONWithBooleans(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]bool{"active": true, "deleted": false}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "\"active\": true") {
		t.Errorf("Expected active=true in output, got: %s", output)
	}
	if !strings.Contains(output, "\"deleted\": false") {
		t.Errorf("Expected deleted=false in output, got: %s", output)
	}
}

func TestPrintJSONWithNilValue(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printJSON(nil)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "null") {
		t.Errorf("Expected null for nil input, got: %s", output)
	}
}

func TestPrintJSONOutputIsIndented(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	data := map[string]string{"key": "value"}
	printJSON(data)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// MarshalIndent with two-space indent should produce indented output
	if !strings.Contains(output, "  ") {
		t.Errorf("Expected indented output with two spaces, got: %s", output)
	}
}
