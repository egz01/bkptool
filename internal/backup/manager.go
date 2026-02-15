package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var ErrNoBackups = errors.New("no backups found for path")

type Entry struct {
	ID           string    `json:"id"`
	SourcePath   string    `json:"sourcePath"`
	SnapshotPath string    `json:"snapshotPath"`
	CreatedAt    time.Time `json:"createdAt"`
	SizeBytes    int64     `json:"sizeBytes"`
	Mode         uint32    `json:"mode"`
}

type Store struct {
	root string
}

func NewStore() (*Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve home directory: %w", err)
	}

	root := filepath.Join(home, ".local", "share", "bkptool")
	if err := os.MkdirAll(filepath.Join(root, "index"), 0o755); err != nil {
		return nil, fmt.Errorf("create index dir: %w", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "snapshots"), 0o755); err != nil {
		return nil, fmt.Errorf("create snapshots dir: %w", err)
	}

	return &Store{root: root}, nil
}

func (s *Store) Backup(path string) (*Entry, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	resolved = filepath.Clean(resolved)

	info, err := os.Stat(resolved)
	if err != nil {
		return nil, fmt.Errorf("stat target: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("directory backup is not implemented yet: %s", resolved)
	}

	history, key, err := s.loadHistory(resolved)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%d", time.Now().UnixNano())
	snapshotDir := filepath.Join(s.root, "snapshots", key)
	if err := os.MkdirAll(snapshotDir, 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot dir: %w", err)
	}

	snapshotPath := filepath.Join(snapshotDir, id+".bak")
	if err := copyFile(resolved, snapshotPath, info.Mode()); err != nil {
		return nil, fmt.Errorf("create snapshot: %w", err)
	}

	entry := Entry{
		ID:           id,
		SourcePath:   resolved,
		SnapshotPath: snapshotPath,
		CreatedAt:    time.Now().UTC(),
		SizeBytes:    info.Size(),
		Mode:         uint32(info.Mode()),
	}
	history = append(history, entry)
	if err := s.saveHistory(key, history); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *Store) Restore(path string, index int, pop bool) (*Entry, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	resolved = filepath.Clean(resolved)

	history, key, err := s.loadHistory(resolved)
	if err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return nil, ErrNoBackups
	}

	pos, err := historyPosFromLatestIndex(history, index)
	if err != nil {
		return nil, err
	}
	entry := history[pos]

	mode := os.FileMode(entry.Mode)
	if mode == 0 {
		mode = 0o644
	}
	if err := copyFile(entry.SnapshotPath, resolved, mode); err != nil {
		return nil, fmt.Errorf("restore file: %w", err)
	}

	if pop {
		history = append(history[:pos], history[pos+1:]...)
		if err := s.saveHistory(key, history); err != nil {
			return nil, err
		}
	}

	return &entry, nil
}

func (s *Store) List(path string) ([]Entry, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}
	resolved = filepath.Clean(resolved)

	history, _, err := s.loadHistory(resolved)
	if err != nil {
		return nil, err
	}
	if len(history) == 0 {
		return nil, ErrNoBackups
	}

	out := make([]Entry, 0, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		out = append(out, history[i])
	}
	return out, nil
}

func (s *Store) DiffWorkingVsLatest(path string) (string, error) {
	resolved, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	resolved = filepath.Clean(resolved)

	history, _, err := s.loadHistory(resolved)
	if err != nil {
		return "", err
	}
	if len(history) == 0 {
		return "", ErrNoBackups
	}

	latest := history[len(history)-1]

	if binary, err := isBinaryFile(resolved); err == nil && binary {
		return "binary file detected; textual diff is not supported for this target", nil
	}
	if binary, err := isBinaryFile(latest.SnapshotPath); err == nil && binary {
		return "binary snapshot detected; textual diff is not supported for this target", nil
	}

	cmd := exec.Command("diff", "-u", latest.SnapshotPath, resolved)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return "no changes", nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
		return string(output), nil
	}

	return "", fmt.Errorf("run diff: %w", err)
}

func (s *Store) loadHistory(sourcePath string) ([]Entry, string, error) {
	key := sourceKey(sourcePath)
	indexPath := filepath.Join(s.root, "index", key+".json")

	b, err := os.ReadFile(indexPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []Entry{}, key, nil
		}
		return nil, "", fmt.Errorf("read index: %w", err)
	}

	var entries []Entry
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, "", fmt.Errorf("decode index: %w", err)
	}

	return entries, key, nil
}

func (s *Store) saveHistory(key string, entries []Entry) error {
	indexPath := filepath.Join(s.root, "index", key+".json")
	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode index: %w", err)
	}
	if err := os.WriteFile(indexPath, b, 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}
	return nil
}

func sourceKey(path string) string {
	h := sha256.Sum256([]byte(strings.ToLower(path)))
	return hex.EncodeToString(h[:])
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}

func historyPosFromLatestIndex(history []Entry, index int) (int, error) {
	if index < 0 {
		return 0, fmt.Errorf("index must be >= 0")
	}
	latest := len(history) - 1
	pos := latest - index
	if pos < 0 || pos > latest {
		return 0, fmt.Errorf("backup index %d out of range (available: 0..%d)", index, latest)
	}
	return pos, nil
}

func isBinaryFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 8000)
	n, err := f.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}
	for _, b := range buf[:n] {
		if b == 0 {
			return true, nil
		}
	}
	return false, nil
}
