package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/egz01/bkptool/internal/backup"
	"github.com/urfave/cli/v3"
)

func main() {
	store, err := backup.NewStore()
	if err != nil {
		log.Fatal(err)
	}

	cmd := &cli.Command{
		Name:  "bkptool",
		Usage: "Ergonomic local backup stack for files and directories",
		Commands: []*cli.Command{
			{
				Name:      "backup",
				ArgsUsage: "<path>",
				Action: func(ctx context.Context, c *cli.Command) error {
					path, err := requirePathArg(c)
					if err != nil {
						return err
					}
					entry, err := store.Backup(path)
					if err != nil {
						return err
					}
					fmt.Printf("backup created: id=%s path=%s at=%s\n", entry.ID, entry.SourcePath, entry.CreatedAt.Format(time.RFC3339))
					return nil
				},
			},
			{
				Name:      "list",
				ArgsUsage: "<path>",
				Action: func(ctx context.Context, c *cli.Command) error {
					path, err := requirePathArg(c)
					if err != nil {
						return err
					}
					entries, err := store.List(path)
					if err != nil {
						if errors.Is(err, backup.ErrNoBackups) {
							fmt.Println("no backups found")
							return nil
						}
						return err
					}

					for idx, entry := range entries {
						fmt.Printf("[%d] id=%s at=%s size=%d snapshot=%s\n", idx, entry.ID, entry.CreatedAt.Format(time.RFC3339), entry.SizeBytes, entry.SnapshotPath)
					}
					return nil
				},
			},
			{
				Name:      "restore",
				ArgsUsage: "<path>",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "index", Aliases: []string{"i"}, Value: 0, Usage: "backup index from latest (0 = latest, 1 = previous)"},
					&cli.BoolFlag{Name: "keep", Value: false, Usage: "restore without removing from stack"},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					path, err := requirePathArg(c)
					if err != nil {
						return err
					}

					index := c.Int("index")
					pop := !c.Bool("keep")
					entry, err := store.Restore(path, index, pop)
					if err != nil {
						if errors.Is(err, backup.ErrNoBackups) {
							return fmt.Errorf("no backups found for target")
						}
						return err
					}
					fmt.Printf("restored: id=%s -> %s (index=%d, pop=%t)\n", entry.ID, path, index, pop)
					return nil
				},
			},
			{
				Name:      "diff",
				ArgsUsage: "<path>",
				Action: func(ctx context.Context, c *cli.Command) error {
					path, err := requirePathArg(c)
					if err != nil {
						return err
					}

					d, err := store.DiffWorkingVsLatest(path)
					if err != nil {
						if errors.Is(err, backup.ErrNoBackups) {
							return fmt.Errorf("no backups found for target")
						}
						return err
					}
					fmt.Println(d)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func requirePathArg(c *cli.Command) (string, error) {
	if c.NArg() == 0 {
		return "", fmt.Errorf("missing required <path> argument")
	}
	return c.Args().Get(0), nil
}
