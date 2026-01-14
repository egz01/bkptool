//go:build tools
// +build tools

package tools

import (
	// --- DB / SQLite ---
	_ "github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	_ "modernc.org/sqlite"

	// --- CLI ---
	_ "github.com/spf13/cobra"
	_ "github.com/spf13/pflag"
	_ "github.com/urfave/cli/v2"

	// --- TUI ---
	_ "github.com/charmbracelet/bubbles"
	_ "github.com/charmbracelet/bubbletea"
	_ "github.com/charmbracelet/lipgloss"
	_ "github.com/gdamore/tcell/v2"
	_ "github.com/rivo/tview"

	// --- OS / FS ---
	_ "github.com/mitchellh/go-homedir"
	_ "github.com/spf13/afero"
	_ "golang.org/x/term"
	_ "golang.org/x/text"

	// --- Config ---
	_ "github.com/joho/godotenv"
	_ "github.com/spf13/viper"

	// --- Logging / Errors ---
	_ "github.com/pkg/errors"
	_ "github.com/rs/zerolog"
	_ "go.uber.org/zap"

	// --- Serialization ---
	_ "github.com/goccy/go-yaml"
	_ "github.com/json-iterator/go"
	_ "gopkg.in/yaml.v3"

	// --- Testing / Dev ---
	_ "github.com/davecgh/go-spew/spew"
	_ "github.com/stretchr/testify"
)
