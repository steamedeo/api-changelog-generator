package main

import (
	"context"
	"fmt"
	"log"
	"os"

	cli "github.com/urfave/cli/v3"
	"steamedeo.dev/api-changelog-generator/cmd"
)

// Version information (injected at build time via ldflags)
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	app := &cli.Command{
		Name:    "api-changelog",
		Usage:   "Generate changelogs from OpenAPI specification changes",
		Version: version,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "version-info",
				Aliases: []string{"V"},
				Usage:   "Show detailed version information",
				Action: func(ctx context.Context, cmd *cli.Command, b bool) error {
					if b {
						fmt.Printf("Version:    %s\n", version)
						fmt.Printf("Git Commit: %s\n", commit)
						fmt.Printf("Build Date: %s\n", date)
						os.Exit(0)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			cmd.NewCompareCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
