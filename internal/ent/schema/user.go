package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// User is an account that can authenticate with the API.
type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().Unique(),
		field.String("password_hash").Sensitive(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sessions", Session.Type),
		edge.To("store_admins", StoreAdmin.Type),
	}
}
