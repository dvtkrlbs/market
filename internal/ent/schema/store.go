package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Store is a global shop, visible to shoppers once approved.
type Store struct{ ent.Schema }

func (Store) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty(),
		field.Bool("approved").Default(false),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (Store) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("admins", StoreAdmin.Type),
		edge.To("products", Product.Type),
	}
}
