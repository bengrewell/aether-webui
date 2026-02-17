package executor

import (
	"bytes"
	"os"
	"text/template"
)

// ReadFile reads the contents of a file.
func (e *DefaultExecutor) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file with the given permissions.
func (e *DefaultExecutor) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// RenderTemplate renders a template string with the given data.
func (e *DefaultExecutor) RenderTemplate(tmpl string, data any) ([]byte, error) {
	t, err := template.New("template").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// RenderTemplateFile reads a template file and renders it with the given data.
func (e *DefaultExecutor) RenderTemplateFile(tmplPath string, data any) ([]byte, error) {
	content, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, err
	}

	return e.RenderTemplate(string(content), data)
}

// FileExists returns true if the file exists.
func (e *DefaultExecutor) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MkdirAll creates a directory and all parent directories.
func (e *DefaultExecutor) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
