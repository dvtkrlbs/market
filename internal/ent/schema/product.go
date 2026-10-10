package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Product struct {
	ent.Schema
}

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.String("barcode").NotEmpty(),
		field.String("name").NotEmpty(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("deleted_at").Optional().Nillable(),
		field.Int("store_id").Immutable(),
	}
}

func (Product) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("store", Store.Type).Ref("products").Field("store_id").Unique().Required().Immutable(),
	}
}

func (Product) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("store_id", "barcode").Unique(),
	}
}
