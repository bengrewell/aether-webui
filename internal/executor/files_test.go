package executor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFile(t *testing.T) {
	e := New(Config{})

	// Create temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test read
	got, err := e.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("ReadFile() = %q, want %q", got, content)
	}
}

func TestReadFile_NotExists(t *testing.T) {
	e := New(Config{})

	_, err := e.ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("ReadFile() should return error for nonexistent file")
	}
}

func TestWriteFile(t *testing.T) {
	e := New(Config{})

	dir := t.TempDir()
	path := filepath.Join(dir, "output.txt")
	content := []byte("test content")

	err := e.WriteFile(path, content, 0644)
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	// Verify
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("WriteFile() content = %q, want %q", got, content)
	}
}

func TestFileExists(t *testing.T) {
	e := New(Config{})

	dir := t.TempDir()
	path := filepath.Join(dir, "exists.txt")

	// Should not exist yet
	if e.FileExists(path) {
		t.Error("FileExists() should return false for nonexistent file")
	}

	// Create file
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Should exist now
	if !e.FileExists(path) {
		t.Error("FileExists() should return true for existing file")
	}
}

func TestMkdirAll(t *testing.T) {
	e := New(Config{})

	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c")

	err := e.MkdirAll(path, 0755)
	if err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("MkdirAll() did not create a directory")
	}
}

func TestRenderTemplate(t *testing.T) {
	e := New(Config{})

	tests := []struct {
		name   string
		tmpl   string
		data   any
		want   string
	}{
		{
			name: "simple",
			tmpl: "Hello, {{.Name}}!",
			data: map[string]string{"Name": "World"},
			want: "Hello, World!",
		},
		{
			name: "no variables",
			tmpl: "static text",
			data: nil,
			want: "static text",
		},
		{
			name: "multiple variables",
			tmpl: "{{.First}} {{.Second}}",
			data: map[string]string{"First": "Hello", "Second": "World"},
			want: "Hello World",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := e.RenderTemplate(tc.tmpl, tc.data)
			if err != nil {
				t.Fatalf("RenderTemplate() error = %v", err)
			}
			if string(got) != tc.want {
				t.Errorf("RenderTemplate() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestRenderTemplate_Error(t *testing.T) {
	e := New(Config{})

	_, err := e.RenderTemplate("{{.Invalid", nil)
	if err == nil {
		t.Error("RenderTemplate() should return error for invalid template")
	}
}

func TestRenderTemplateFile(t *testing.T) {
	e := New(Config{})

	dir := t.TempDir()
	path := filepath.Join(dir, "template.txt")
	if err := os.WriteFile(path, []byte("Hello, {{.Name}}!"), 0644); err != nil {
		t.Fatalf("failed to create template file: %v", err)
	}

	got, err := e.RenderTemplateFile(path, map[string]string{"Name": "World"})
	if err != nil {
		t.Fatalf("RenderTemplateFile() error = %v", err)
	}
	if string(got) != "Hello, World!" {
		t.Errorf("RenderTemplateFile() = %q, want %q", got, "Hello, World!")
	}
}

func TestRenderTemplateFile_NotExists(t *testing.T) {
	e := New(Config{})

	_, err := e.RenderTemplateFile("/nonexistent/template.txt", nil)
	if err == nil {
		t.Error("RenderTemplateFile() should return error for nonexistent file")
	}
}
