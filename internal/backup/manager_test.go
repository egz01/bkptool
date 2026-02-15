package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupPushAndRestorePopLIFO(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := NewStore()
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	target := filepath.Join(home, "file.txt")
	if err := os.WriteFile(target, []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if _, err := store.Backup(target); err != nil {
		t.Fatalf("backup v1: %v", err)
	}
	if err := os.WriteFile(target, []byte("v2\n"), 0o644); err != nil {
		t.Fatalf("update file: %v", err)
	}
	if _, err := store.Backup(target); err != nil {
		t.Fatalf("backup v2: %v", err)
	}

	if err := os.WriteFile(target, []byte("broken\n"), 0o644); err != nil {
		t.Fatalf("break file: %v", err)
	}

	if _, err := store.Restore(target, 0, true); err != nil {
		t.Fatalf("restore latest: %v", err)
	}
	b, _ := os.ReadFile(target)
	if string(b) != "v2\n" {
		t.Fatalf("expected v2 after restore, got: %q", string(b))
	}

	entries, err := store.List(target)
	if err != nil {
		t.Fatalf("list after pop: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 backup left, got %d", len(entries))
	}
}

func TestRestoreByIndex(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := NewStore()
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	target := filepath.Join(home, "file.txt")
	if err := os.WriteFile(target, []byte("v1\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	if _, err := store.Backup(target); err != nil {
		t.Fatalf("backup v1: %v", err)
	}

	if err := os.WriteFile(target, []byte("v2\n"), 0o644); err != nil {
		t.Fatalf("update file: %v", err)
	}
	if _, err := store.Backup(target); err != nil {
		t.Fatalf("backup v2: %v", err)
	}

	if err := os.WriteFile(target, []byte("v3\n"), 0o644); err != nil {
		t.Fatalf("update file: %v", err)
	}
	if _, err := store.Backup(target); err != nil {
		t.Fatalf("backup v3: %v", err)
	}

	if err := os.WriteFile(target, []byte("oops\n"), 0o644); err != nil {
		t.Fatalf("break file: %v", err)
	}

	// index 2 from latest => oldest => v1
	if _, err := store.Restore(target, 2, false); err != nil {
		t.Fatalf("restore index 2: %v", err)
	}
	b, _ := os.ReadFile(target)
	if string(b) != "v1\n" {
		t.Fatalf("expected v1 after restore-by-index, got: %q", string(b))
	}
}

func TestRestoreEmptyHistory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	store, err := NewStore()
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	target := filepath.Join(home, "missing.txt")
	_, err = store.Restore(target, 0, true)
	if err == nil {
		t.Fatalf("expected error for empty history")
	}
}
