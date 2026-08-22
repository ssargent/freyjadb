package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

const formatJSON = "json"

// outputEntity displays a single entity
func outputEntity(entity *Entity) error {
	if config.Format == formatJSON {
		return outputEntityJSON(entity)
	}
	return outputEntityTable(entity)
}

// outputEntities displays multiple entities
func outputEntities(entities []*Entity) error {
	if config.Format == formatJSON {
		return outputEntitiesJSON(entities)
	}
	return outputEntitiesTable(entities)
}

// entityTableField is a single label/value row rendered by outputEntityTable.
type entityTableField struct {
	label string
	value string
}

// entityTableFields builds the ordered list of label/value rows to display
// for an entity, omitting fields that are empty.
func entityTableFields(entity *Entity) []entityTableField {
	fields := []entityTableField{
		{"ID", entity.ID},
		{"Type", string(entity.Type)},
		{"Name", entity.Name},
	}

	if len(entity.Aka) > 0 {
		fields = append(fields, entityTableField{"AKA", formatStringSlice(entity.Aka)})
	}
	if entity.Summary != "" {
		fields = append(fields, entityTableField{"Summary", entity.Summary})
	}
	if entity.Details != "" {
		fields = append(fields, entityTableField{"Details", entity.Details})
	}
	if len(entity.Tags) > 0 {
		fields = append(fields, entityTableField{"Tags", formatStringSlice(entity.Tags)})
	}
	if len(entity.Links) > 0 {
		fields = append(fields, entityTableField{"Links", fmt.Sprintf("%d relationships", len(entity.Links))})
	}

	fields = append(fields,
		entityTableField{"Created", entity.CreatedAt.Format(time.RFC3339)},
		entityTableField{"Updated", entity.UpdatedAt.Format(time.RFC3339)},
	)

	return fields
}

// outputEntityTable displays a single entity in table format
func outputEntityTable(entity *Entity) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()

	for _, field := range entityTableFields(entity) {
		if _, err := fmt.Fprintf(w, "%s:\t%s\n", field.label, field.value); err != nil {
			return err
		}
	}

	return nil
}

// outputEntitiesTable displays multiple entities in table format
func outputEntitiesTable(entities []*Entity) error {
	if len(entities) == 0 {
		fmt.Println("No entities found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer func() { _ = w.Flush() }()

	if _, err := fmt.Fprintln(w, "ID\tNAME\tTYPE\tSUMMARY\tTAGS\tUPDATED"); err != nil {
		return err
	}

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
			if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				entity.ID,
				entity.Name,
				entity.Type,
				summary,
				tags,
				entity.UpdatedAt.Format("2006-01-02 15:04")); err != nil {
				return err
			}
		}
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
	if config.Format == formatJSON {
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
