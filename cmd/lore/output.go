package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// outputEntity displays a single entity
func outputEntity(entity *Entity) error {
	if config.Format == "json" {
		return outputEntityJSON(entity)
	}
	return outputEntityTable(entity)
}

// outputEntities displays multiple entities
func outputEntities(entities []*Entity) error {
	if config.Format == "json" {
		return outputEntitiesJSON(entities)
	}
	return outputEntitiesTable(entities)
}

// outputEntityTable displays a single entity in table format
func outputEntityTable(entity *Entity) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintf(w, "ID:\t%s\n", entity.ID)
	fmt.Fprintf(w, "Type:\t%s\n", entity.Type)
	fmt.Fprintf(w, "Name:\t%s\n", entity.Name)

	if len(entity.Aka) > 0 {
		fmt.Fprintf(w, "AKA:\t%s\n", formatStringSlice(entity.Aka))
	}

	if entity.Summary != "" {
		fmt.Fprintf(w, "Summary:\t%s\n", entity.Summary)
	}

	if entity.Details != "" {
		fmt.Fprintf(w, "Details:\t%s\n", entity.Details)
	}

	if len(entity.Tags) > 0 {
		fmt.Fprintf(w, "Tags:\t%s\n", formatStringSlice(entity.Tags))
	}

	if len(entity.Links) > 0 {
		fmt.Fprintf(w, "Links:\t%d relationships\n", len(entity.Links))
	}

	fmt.Fprintf(w, "Created:\t%s\n", entity.CreatedAt.Format(time.RFC3339))
	fmt.Fprintf(w, "Updated:\t%s\n", entity.UpdatedAt.Format(time.RFC3339))

	return nil
}

// outputEntitiesTable displays multiple entities in table format
func outputEntitiesTable(entities []*Entity) error {
	if len(entities) == 0 {
		fmt.Println("No entities found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Fprintln(w, "ID\tNAME\tTYPE\tSUMMARY\tTAGS\tUPDATED")

	for _, entity := range entities {
		summary := entity.Summary
		if len(summary) > 50 {
			summary = summary[:47] + "..."
		}

		tags := ""
		if len(entity.Tags) > 0 {
			tags = formatStringSlice(entity.Tags)
			if len(tags) > 30 {
				tags = tags[:27] + "..."
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			entity.ID,
			entity.Name,
			entity.Type,
			summary,
			tags,
			entity.UpdatedAt.Format("2006-01-02 15:04"))
	}

	return nil
}

// outputEntityJSON displays a single entity in JSON format
func outputEntityJSON(entity *Entity) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entity)
}

// outputEntitiesJSON displays multiple entities in JSON format
func outputEntitiesJSON(entities []*Entity) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entities)
}

// outputEntityWithRelationships displays an entity with its relationships
func outputEntityWithRelationships(entityWithRels *EntityWithRelationships) error {
	if config.Format == "json" {
		return outputEntityWithRelationshipsJSON(entityWithRels)
	}
	return outputEntityWithRelationshipsTable(entityWithRels)
}

// outputEntityWithRelationshipsTable displays an entity with relationships in table format
func outputEntityWithRelationshipsTable(entityWithRels *EntityWithRelationships) error {
	entity := entityWithRels.Entity

	// Display the entity
	fmt.Printf("Entity: %s:%s (%s)\n", entity.Type, entity.ID, entity.Name)
	fmt.Printf("Summary: %s\n", entity.Summary)
	if len(entity.Tags) > 0 {
		fmt.Printf("Tags: %s\n", formatStringSlice(entity.Tags))
	}
	fmt.Println()

	// Display outgoing relationships
	if len(entityWithRels.Outgoing) > 0 {
		fmt.Println("Outgoing Relationships:")
		for _, rel := range entityWithRels.Outgoing {
			fmt.Printf("  --[%s]--> %s\n", rel.Relationship.Relation, rel.OtherKey)
		}
		fmt.Println()
	} else {
		fmt.Println("No outgoing relationships")
		fmt.Println()
	}

	// Display incoming relationships
	if len(entityWithRels.Incoming) > 0 {
		fmt.Println("Incoming Relationships:")
		for _, rel := range entityWithRels.Incoming {
			fmt.Printf("  <--[%s]-- %s\n", rel.Relationship.Relation, rel.OtherKey)
		}
		fmt.Println()
	} else {
		fmt.Println("No incoming relationships")
		fmt.Println()
	}

	return nil
}

// outputEntityWithRelationshipsJSON displays an entity with relationships in JSON format
func outputEntityWithRelationshipsJSON(entityWithRels *EntityWithRelationships) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entityWithRels)
}

// formatStringSlice formats a slice of strings for display
func formatStringSlice(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += ", " + slice[i]
	}
	return result
}
