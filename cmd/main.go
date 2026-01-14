package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	var language string

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lang",
				Aliases:     []string{"l"},
				Value:       "english",
				Usage:       "language for the greeting",
				Destination: &language,
				Sources:     cli.EnvVars("APP_LANG"),
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			name := "Nefretiti"
			if c.NArg() > 0 {
				name = c.Args().Get(0)
			}
			if language == "spanish" {
				fmt.Println("Hola", name)
			} else {
				fmt.Println("Hello", name)
			}
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
