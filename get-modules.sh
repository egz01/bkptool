#!/usr/bin/env bash

go get modernc.org/sqlite
go get github.com/mattn/go-sqlite3
go get github.com/jmoiron/sqlx

go get github.com/urfave/cli/v2
go get github.com/spf13/cobra
go get github.com/spf13/pflag

go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles
go get github.com/charmbracelet/lipgloss
go get github.com/gdamore/tcell/v2
go get github.com/rivo/tview

go get golang.org/x/sys
go get golang.org/x/term
go get golang.org/x/text
go get github.com/spf13/afero
go get github.com/mitchellh/go-homedir

go get github.com/spf13/viper
go get github.com/joho/godotenv

go get github.com/rs/zerolog
go get go.uber.org/zap
go get github.com/pkg/errors

go get github.com/goccy/go-yaml
go get gopkg.in/yaml.v3
go get github.com/json-iterator/go

go get github.com/stretchr/testify
go get github.com/davecgh/go-spew/spew
go get golang.org/x/tools

go list -deps ./... >/dev/null

CGO_ENABLED=1 go get github.com/mattn/go-sqlite3
