package openapi

import (
	"bufio"
	"fmt"
	"os"
	"strings"
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
	property := change.Property

	// Helper to format text values - no truncation, show everything
	formatValue := func(val string) string {
		// Skip if empty
		if val == "" {
			return ""
		}

		// Replace newlines with spaces for single-line display
		return strings.ReplaceAll(val, "\n", " ")
	}

	// Make property names more human-readable
	switch property {
	case "codes":
		if change.New != "" {
			return fmt.Sprintf("Response code `%s` added", change.New)
		} else if change.Original != "" {
			return fmt.Sprintf("Response code `%s` removed", change.Original)
		}
	case "parameters":
		if change.New != "" {
			return fmt.Sprintf("Parameter `%s` added", change.New)
		} else if change.Original != "" {
			return fmt.Sprintf("Parameter `%s` removed", change.Original)
		}
	case "properties":
		// Note: context must be added by caller - this is just the base message
		if change.New != "" {
			return fmt.Sprintf("Property `%s` added", change.New)
		} else if change.Original != "" {
			return fmt.Sprintf("Property `%s` removed", change.Original)
		}
	case "schemas":
		if change.New != "" {
			return fmt.Sprintf("Schema `%s` added", change.New)
		} else if change.Original != "" {
			return fmt.Sprintf("Schema `%s` removed", change.Original)
		}
	case "deprecated":
		if change.New == "true" {
			return "Marked as deprecated"
		} else {
			return "No longer deprecated"
		}
	case "version":
		return fmt.Sprintf("API version updated from `%s` to `%s`", change.Original, change.New)
	case "description":
		// Special handling for description changes
		oldDesc := formatValue(change.Original)
		newDesc := formatValue(change.New)
		return fmt.Sprintf("Description updated:\n  - Old: %s\n  - New: %s", oldDesc, newDesc)
	case "summary":
		oldSummary := formatValue(change.Original)
		newSummary := formatValue(change.New)
		return fmt.Sprintf("Summary updated from '%s' to '%s'", oldSummary, newSummary)
	case "$ref", "reference":
		if change.New != "" && change.Original != "" {
			// Skip if values are the same
			if change.New == change.Original {
				return ""
			}
			return fmt.Sprintf("Reference changed from `%s` to `%s`", change.Original, change.New)
		} else if change.New != "" {
			return fmt.Sprintf("Reference set to `%s`", change.New)
		} else {
			return fmt.Sprintf("Reference `%s` removed", change.Original)
		}
	case "url":
		if change.New != "" && change.Original != "" {
			return fmt.Sprintf("URL changed from `%s` to `%s`", change.Original, change.New)
		}
	case "type":
		if change.New != "" && change.Original != "" {
			return fmt.Sprintf("Type changed from `%s` to `%s`", change.Original, change.New)
		}
	case "required":
		if change.New == "true" && change.Original == "false" {
			return "Now required"
		} else if change.New == "false" && change.Original == "true" {
			return "No longer required"
		}
	case "format":
		if change.New != "" && change.Original != "" {
			return fmt.Sprintf("Format changed from `%s` to `%s`", change.Original, change.New)
		}
	default:
		// Handle paths (endpoints)
		if len(property) > 0 && property[0] == '/' {
			return fmt.Sprintf("Endpoint `%s` added", property)
		} else if strings.HasPrefix(property, "x-") {
			// Extension property
			if change.Original != "" && change.New != "" {
				return fmt.Sprintf("Extension `%s` modified", property)
			} else if change.New != "" {
				return fmt.Sprintf("Extension `%s` added", property)
			} else {
				return fmt.Sprintf("Extension `%s` removed", property)
			}
		} else {
			// Generic fallback with better formatting
			if change.Original != "" && change.New != "" {
				// Skip if values are the same
				if change.Original == change.New {
					return ""
				}
				oldVal := formatValue(change.Original)
				newVal := formatValue(change.New)
				return fmt.Sprintf("`%s` changed from '%s' to '%s'", property, oldVal, newVal)
			} else if change.New != "" {
				newVal := formatValue(change.New)
				return fmt.Sprintf("`%s` set to '%s'", property, newVal)
			} else if change.Original != "" {
				oldVal := formatValue(change.Original)
				return fmt.Sprintf("`%s` removed (was '%s')", property, oldVal)
			} else {
				return fmt.Sprintf("`%s` modified", property)
			}
		}
	}

	return ""
}

// Helper function to format a change with context
func (c *Comparer) formatChangeWithContext(change *model.Change, context string) string {
	desc := c.formatChange(change)
	if desc == "" {
		return ""
	}

	// Always prepend context if provided and not already in the description
	if context != "" && !strings.Contains(desc, context) {
		return fmt.Sprintf("**%s**: %s", context, desc)
	}
	return desc
}

// Process operation changes with endpoint context
func (c *Comparer) processOperationChangesWithContext(endpoint string, opChanges *model.OperationChanges,
	added, modified, removed, breaking *[]string, seen map[string]bool) {

	if opChanges == nil {
		return
	}

	// Get all changes from this operation
	changes := opChanges.GetAllChanges()
	for _, change := range changes {
		// Create unique key for deduplication based on change properties
		changeKey := fmt.Sprintf("%d|%s|%s|%s", change.ChangeType, change.Property, change.Original, change.New)
		if seen[changeKey] {
			continue
		}
		seen[changeKey] = true

		changeDesc := c.formatChangeWithContext(change, endpoint)

		// Skip empty descriptions
		if changeDesc == "" {
			continue
		}

		// Categorize
		isActuallyBreaking := change.Breaking &&
			(change.ChangeType == model.ObjectRemoved ||
				change.ChangeType == model.PropertyRemoved ||
				(change.Property == "required" && change.New == "true" && change.Original == "false"))

		if isActuallyBreaking {
			*breaking = append(*breaking, changeDesc)
		}

		switch change.ChangeType {
		case model.ObjectAdded, model.PropertyAdded:
			*added = append(*added, changeDesc)
		case model.Modified:
			*modified = append(*modified, changeDesc)
		case model.ObjectRemoved, model.PropertyRemoved:
			*removed = append(*removed, changeDesc)
		}
	}
}

// Process schema changes with schema context
func (c *Comparer) processSchemaChangesWithContext(schemaName string, schemaChanges *model.SchemaChanges,
	added, modified, removed, breaking *[]string, seen map[string]bool) {

	if schemaChanges == nil {
		return
	}

	// Get all changes from this schema
	changes := schemaChanges.GetAllChanges()
	for _, change := range changes {
		// Create unique key for deduplication based on change properties
		changeKey := fmt.Sprintf("%d|%s|%s|%s", change.ChangeType, change.Property, change.Original, change.New)
		if seen[changeKey] {
			continue
		}
		seen[changeKey] = true

		changeDesc := c.formatChangeWithContext(change, schemaName)

		// Skip empty descriptions
		if changeDesc == "" {
			continue
		}

		// Categorize
		isActuallyBreaking := change.Breaking &&
			(change.ChangeType == model.ObjectRemoved ||
				change.ChangeType == model.PropertyRemoved ||
				(change.Property == "required" && change.New == "true" && change.Original == "false"))

		if isActuallyBreaking {
			*breaking = append(*breaking, changeDesc)
		}

		switch change.ChangeType {
		case model.ObjectAdded, model.PropertyAdded:
			*added = append(*added, changeDesc)
		case model.Modified:
			*modified = append(*modified, changeDesc)
		case model.ObjectRemoved, model.PropertyRemoved:
			*removed = append(*removed, changeDesc)
		}
	}
}

func (c *Comparer) GenerateChangelog() error {
	if err := c.compare(); err != nil {
		return err
	}
	if c.totalChanges > 0 {
		date := time.Now().Format("2006-01-02")

		// Get API version from the document info
		apiVersion := "Unknown"
		v3Model, errs := c.latestDocument.BuildV3Model()
		if errs == nil && v3Model != nil && v3Model.Model.Info != nil {
			apiVersion = v3Model.Model.Info.Version
		}

		// Separate endpoint additions/removals from other changes
		endpointAdditions := []string{}
		endpointRemovals := []string{}
		added := []string{}
		modified := []string{}
		removed := []string{}
		breaking := []string{}

		// Track seen changes to avoid duplicates
		seen := make(map[string]bool)

		// Process path changes with context
		if c.changes.PathsChanges != nil && c.changes.PathsChanges.PathItemsChanges != nil {
			for path, pathItemChanges := range c.changes.PathsChanges.PathItemsChanges {
				// Check each HTTP method
				methods := []struct {
					name    string
					changes *model.OperationChanges
				}{
					{"GET", pathItemChanges.GetChanges},
					{"POST", pathItemChanges.PostChanges},
					{"PUT", pathItemChanges.PutChanges},
					{"DELETE", pathItemChanges.DeleteChanges},
					{"PATCH", pathItemChanges.PatchChanges},
					{"OPTIONS", pathItemChanges.OptionsChanges},
					{"HEAD", pathItemChanges.HeadChanges},
					{"TRACE", pathItemChanges.TraceChanges},
				}

				for _, method := range methods {
					if method.changes != nil {
						endpoint := method.name + " " + path
						c.processOperationChangesWithContext(endpoint, method.changes, &added, &modified, &removed, &breaking, seen)
					}
				}
			}
		}

		// Process component/schema changes with context
		if c.changes.ComponentsChanges != nil && c.changes.ComponentsChanges.SchemaChanges != nil {
			for schemaName, schemaChanges := range c.changes.ComponentsChanges.SchemaChanges {
				context := "Schema `" + schemaName + "`"
				c.processSchemaChangesWithContext(context, schemaChanges, &added, &modified, &removed, &breaking, seen)
			}
		}

		// Process component-level changes (parameters, responses, examples, etc.)
		if c.changes.ComponentsChanges != nil {
			componentChanges := c.changes.ComponentsChanges.GetAllChanges()
			for _, change := range componentChanges {
				// Create unique key for deduplication
				changeKey := fmt.Sprintf("%d|%s|%s|%s", change.ChangeType, change.Property, change.Original, change.New)
				if seen[changeKey] {
					continue
				}
				seen[changeKey] = true

				// Add context to indicate these are reusable component definitions
				var context string
				switch change.Property {
				case "parameters":
					context = "Components"
				case "responses":
					context = "Components"
				case "examples":
					context = "Components"
				case "requestBodies":
					context = "Components"
				case "headers":
					context = "Components"
				case "links":
					context = "Components"
				case "callbacks":
					context = "Components"
				default:
					context = "Components"
				}

				changeDesc := c.formatChangeWithContext(change, context)
				if changeDesc == "" {
					continue
				}

				// Categorize
				isActuallyBreaking := change.Breaking &&
					(change.ChangeType == model.ObjectRemoved ||
						change.ChangeType == model.PropertyRemoved ||
						(change.Property == "required" && change.New == "true" && change.Original == "false"))

				if isActuallyBreaking {
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
		}

		// Process any remaining changes without specific context
		allChanges := c.changes.GetAllChanges()
		for _, change := range allChanges {
			// Create unique key for deduplication based on change properties
			changeKey := fmt.Sprintf("%d|%s|%s|%s", change.ChangeType, change.Property, change.Original, change.New)

			// Skip duplicates - these were already processed with proper context
			if seen[changeKey] {
				continue
			}
			seen[changeKey] = true

			changeDesc := c.formatChange(change)

			// Skip empty descriptions
			if changeDesc == "" {
				continue
			}

			// Separate endpoint additions/removals
			if strings.HasPrefix(changeDesc, "Endpoint ") && strings.Contains(changeDesc, " added") {
				endpointAdditions = append(endpointAdditions, changeDesc)
				continue
			}
			if strings.HasPrefix(changeDesc, "Endpoint ") && strings.Contains(changeDesc, " removed") {
				endpointRemovals = append(endpointRemovals, changeDesc)
				breaking = append(breaking, changeDesc)
				continue
			}

			// Only mark truly breaking changes as breaking
			isActuallyBreaking := change.Breaking &&
				(change.ChangeType == model.ObjectRemoved ||
					change.ChangeType == model.PropertyRemoved ||
					(change.Property == "required" && change.New == "true" && change.Original == "false"))

			if isActuallyBreaking {
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

		// Show all endpoints - no truncation

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
		fmt.Fprintf(w, "**Breaking Changes:** %d\n", len(breaking))
		if len(endpointAdditions) > 0 {
			fmt.Fprintf(w, "**Endpoints Added:** %d\n", len(endpointAdditions))
		}
		if len(endpointRemovals) > 0 {
			fmt.Fprintf(w, "**Endpoints Removed:** %d\n", len(endpointRemovals))
		}
		fmt.Fprintln(w)

		// Write breaking changes section if any
		if len(breaking) > 0 {
			fmt.Fprintln(w, "### ⚠️ Breaking Changes")
			for _, change := range breaking {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		// Write endpoint additions section
		if len(endpointAdditions) > 0 {
			fmt.Fprintln(w, "### 🆕 New Endpoints")
			for _, change := range endpointAdditions {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		// Write endpoint removals section
		if len(endpointRemovals) > 0 {
			fmt.Fprintln(w, "### 🗑️ Removed Endpoints")
			for _, change := range endpointRemovals {
				fmt.Fprintf(w, "- %s\n", change)
			}
			fmt.Fprintln(w)
		}

		// Write added section (non-endpoint additions)
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

		// Write removed section (non-endpoint removals)
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
