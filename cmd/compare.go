package cmd

import (
	"context"
	"log"

	"github.com/urfave/cli/v3"
	"steamedeo.dev/api-changelog-generator/internal/openapi"
)

func NewCompareCommand() *cli.Command {
	return &cli.Command{
		Name:  "compare",
		Usage: "compare two versions of OpenAPI files and generate a changelog",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "latest",
				Usage:    "path to the latest version of OpenAPI file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "previous",
				Usage:    "path to the previous version of OpenAPI file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "output",
				Usage:    "path where to save the output changelog file",
				Required: true,
			},
		},
		Action: compareOpenAPI,
	}
}

func compareOpenAPI(ctx context.Context, cmd *cli.Command) error {
	latestPath := cmd.String("latest")
	previousPath := cmd.String("previous")
	outputPath := cmd.String("output")

	comparer, err := openapi.NewComparer(latestPath, previousPath, outputPath)
	if err != nil {
		return err
	}
	changelogErr := comparer.GenerateChangelog()
	if changelogErr != nil {
		return changelogErr
	}
	if comparer.GetChangelogPath() == "" {
		log.Println("No changes detected between the two OpenAPI documents.")
	} else {
		log.Printf("Changelog generated at: %s\n", comparer.GetChangelogPath())
	}

	return nil
}
