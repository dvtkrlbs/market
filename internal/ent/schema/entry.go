package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// Entry records one movement of money. Positive amounts are income and
// negative amounts are expenses.
type Entry struct {
	ent.Schema
}

func (Entry) Fields() []ent.Field {
	return []ent.Field{
		field.String("description").NotEmpty(),
		field.Int64("amount_cents"),
		field.Time("occurred_at"),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

func (Entry) Edges() []ent.Edge {
	return nil
}
