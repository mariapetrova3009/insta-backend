package storage

import (
	"fmt"           // форматирование ошибок
	"net/url"       // сборка корректного file:// URL
	"os"            // файловые операции
	"path/filepath" // нормализация путей и склейка
	"strings"       // проверки префиксов
	"time"          // TTL для Presign (пока не используем)
)

type LocalFS struct{ root string }

func NewLocalFS(root string) *LocalFS { return &LocalFS{root: root} }

func (s *LocalFS) Put(name string, data []byte, _ string) (PutResult, error) {
	// check path
	clean := filepath.Clean(name)
	if clean == "." || clean == "" {
		return PutResult{}, fmt.Errorf("empty name")
	}
	// make abs path
	full := filepath.Join(s.root, clean)

	// compare abs paths
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return PutResult{}, fmt.Errorf("abs(root): %w", err)
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return PutResult{}, fmt.Errorf("abs(full): %w", err)
	}

	sep := string(os.PathSeparator)
	if !strings.HasPrefix(absFull+sep, absRoot+sep) {
		return PutResult{}, fmt.Errorf("path escapes root: %q", clean)
	}

	// make dir and write file
	if err := os.MkdirAll(filepath.Dir(absFull), 0o755); err != nil {
		return PutResult{}, fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(absFull, data, 0o644); err != nil {
		return PutResult{}, fmt.Errorf("write: %w", err)
	}

	// calculate size
	var size int64
	if fi, err := os.Stat(absFull); err == nil {
		size = fi.Size()
	}

	return PutResult{Key: clean, Size: size}, nil
}

func (s *LocalFS) Delete(key string) error {
	clean := filepath.Clean(key)
	full := filepath.Join(s.root, clean)

	// check path
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return fmt.Errorf("abs(root): %w", err)
	}
	absFull, err := filepath.Abs(full)
	if err != nil {
		return fmt.Errorf("abs(full): %w", err)
	}
	sep := string(os.PathSeparator)
	if !strings.HasPrefix(absFull+sep, absRoot+sep) {
		return fmt.Errorf("path escapes root: %q", clean)
	}

	// delete file
	if err := os.Remove(absFull); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove: %w", err)
	}
	return nil
}

// Presign returns link to download
func (s *LocalFS) Presign(key string, _ time.Duration) (string, error) {
	clean := filepath.Clean(key)
	abs := filepath.Join(s.root, clean)

	path := filepath.ToSlash(abs)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := url.URL{Scheme: "file", Path: path}
	return u.String(), nil
}
