package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// StoreAdmin grants a user permission to manage a store.
type StoreAdmin struct{ ent.Schema }

func (StoreAdmin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("user_id").Immutable(),
		field.Int("store_id").Immutable(),
	}
}

func (StoreAdmin) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).Ref("store_admins").Field("user_id").Unique().Required().Immutable(),
		edge.From("store", Store.Type).Ref("admins").Field("store_id").Unique().Required().Immutable(),
	}
}

func (StoreAdmin) Indexes() []ent.Index {
	return []ent.Index{index.Fields("user_id", "store_id").Unique(), index.Fields("store_id")}
}
