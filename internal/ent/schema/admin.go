package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Admin grants a user platform-wide administration permissions.
type Admin struct{ ent.Schema }

func (Admin) Fields() []ent.Field {
	return []ent.Field{field.Int("user_id").Unique().Immutable()}
}

func (Admin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("admin").Field("user_id").Unique().Required().Immutable(),
	}
}
