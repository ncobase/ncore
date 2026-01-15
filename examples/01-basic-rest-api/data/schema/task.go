// Package schema defines Ent schemas for the basic REST API example.
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Task holds the schema definition for the Task entity.
type Task struct {
	ent.Schema
}

// Fields of the Task.
func (Task) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Positive().
			Comment("Task ID"),
		field.String("title").
			NotEmpty().
			MaxLen(255).
			Comment("Task title"),
		field.Text("description").
			Optional().
			Comment("Task description"),
		field.Enum("status").
			Values("pending", "in_progress", "completed", "cancelled").
			Default("pending").
			Comment("Task status"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("Creation timestamp"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("Last update timestamp"),
	}
}

// Edges of the Task.
func (Task) Edges() []ent.Edge {
	return nil
}
