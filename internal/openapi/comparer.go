package openapi

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/pb33f/libopenapi"
	"github.com/pb33f/libopenapi/what-changed/model"
)

type Comparer struct {
	outputPath       string
	latestDocument   libopenapi.Document
	previousDocument libopenapi.Document
	totalChanges     int
	changes          *model.DocumentChanges
	changelogPath    string
}

func (c *Comparer) compare() error {
	changes, err := libopenapi.CompareDocuments(c.previousDocument, c.latestDocument)
	if err != nil {
		return err
	}
	c.totalChanges = changes.TotalChanges()
	c.changes = changes
	return nil
}

func (c *Comparer) formatChange(change *model.Change) string {
	var desc string
	property := change.Property

	// Make property names more human-readable
	switch property {
	case "codes":
		if change.New != "" {
			desc = fmt.Sprintf("Response code `%s` added", change.New)
		} else if change.Original != "" {
			desc = fmt.Sprintf("Response code `%s` removed", change.Original)
		}
	case "parameters":
		if change.New != "" {
			desc = fmt.Sprintf("Parameter `%s` added", change.New)
		} else if change.Original != "" {
			desc = fmt.Sprintf("Parameter `%s` removed", change.Original)
		}
	case "properties":
		if change.New != "" {
			desc = fmt.Sprintf("Property `%s` added to schema", change.New)
		} else if change.Original != "" {
			desc = fmt.Sprintf("Property `%s` removed from schema", change.Original)
		}
	case "schemas":
		if change.New != "" {
			desc = fmt.Sprintf("Schema `%s` added", change.New)
		} else if change.Original != "" {
			desc = fmt.Sprintf("Schema `%s` removed", change.Original)
		}
	case "deprecated":
		if change.New == "true" {
			desc = "Endpoint marked as deprecated"
		} else {
			desc = "Endpoint no longer deprecated"
		}
	case "version":
		desc = fmt.Sprintf("API version updated from `%s` to `%s`", change.Original, change.New)
	default:
		// Handle paths (endpoints)
		if len(property) > 0 && property[0] == '/' {
			desc = fmt.Sprintf("Endpoint `%s` added", property)
		} else {
			// Generic fallback
			if change.Original != "" && change.New != "" {
				desc = fmt.Sprintf("%s: changed from `%s` to `%s`", property, change.Original, change.New)
			} else if change.New != "" {
				desc = fmt.Sprintf("%s: `%s`", property, change.New)
			} else if change.Original != "" {
				desc = fmt.Sprintf("%s: was `%s`", property, change.Original)
			} else {
				desc = property
			}
		}
	}

	return desc
}

func (c *Comparer) GenerateChangelog() error {
	if err := c.compare(); err != nil {
		return err
	}
	if c.totalChanges > 0 {
		date := time.Now().Format("2006-01-02")
		allChanges := c.changes.GetAllChanges()

		// Get API version from the document info
		apiVersion := "Unknown"
		v3Model, errs := c.latestDocument.BuildV3Model()
		if errs == nil && v3Model != nil && v3Model.Model.Info != nil {
			apiVersion = v3Model.Model.Info.Version
		}

		// Group changes by type
		added := []string{}
		modified := []string{}
		removed := []string{}
		breaking := []string{}

		for _, change := range allChanges {
			changeDesc := c.formatChange(change)

			if change.Breaking {
				breaking = append(breaking, changeDesc)
			}

			switch change.ChangeType {
			case model.ObjectAdded, model.PropertyAdded:
				added = append(added, changeDesc)
			case model.Modified:
				modified = append(modified, changeDesc)
			case model.ObjectRemoved, model.PropertyRemoved:
				removed = append(removed, changeDesc)
			}
		}

		file, err := os.Create(c.outputPath)
		if err != nil {
			return err
		}
		defer file.Close()
		w := bufio.NewWriter(file)

		// Write header
		fmt.Fprintln(w, "# API Changelog")
		fmt.Fprintf(w, "\n## [%s] - %s\n\n", apiVersion, date)

		// Write summary
		fmt.Fprintf(w, "**Total Changes:** %d\n", c.totalChanges)
		fmt.Fprintf(w, "**Breaking Changes:** %d\n\n", c.changes.TotalBreakingChanges())

		// Write breaking changes section if any
		if len(breaking) > 0 {
			fmt.Fprintln(w, "### ⚠️ Breaking Changes")
			for _, change := range breaking {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		// Write added section
		if len(added) > 0 {
			fmt.Fprintln(w, "### ✨ Added")
			for _, change := range added {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		// Write modified section
		if len(modified) > 0 {
			fmt.Fprintln(w, "### 🔄 Modified")
			for _, change := range modified {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		// Write removed section
		if len(removed) > 0 {
			fmt.Fprintln(w, "### ❌ Removed")
			for _, change := range removed {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		w.Flush()
		c.changelogPath = c.outputPath
	}

	return nil
}

func (c *Comparer) GetChangelogPath() string {
	return c.changelogPath
}

func NewComparer(latestPath, previousPath, outputPath string) (*Comparer, error) {
	// Read latest document
	latestDocumentOpenAPIFile, err := os.ReadFile(latestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", latestPath, err)
	}
	latestDocument, err := libopenapi.NewDocument(latestDocumentOpenAPIFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document %s: %w", latestPath, err)
	}

	// Read previous document
	previousDocumentOpenAPIFile, err := os.ReadFile(previousPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", previousPath, err)
	}
	previousDocument, err := libopenapi.NewDocument(previousDocumentOpenAPIFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI document %s: %w", previousPath, err)
	}

	return &Comparer{
		latestDocument:   latestDocument,
		previousDocument: previousDocument,
		outputPath:       outputPath,
	}, nil
}
