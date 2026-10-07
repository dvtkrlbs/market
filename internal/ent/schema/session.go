package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Session is a revocable, database-backed login session.
type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("session_id").Immutable(),
		field.String("session_secret").Sensitive(),
		field.Int("user_id").Immutable(),
		field.String("user_agent").Optional().Nillable(),
		field.String("ip").Optional().Nillable().SchemaType(map[string]string{
			dialect.Postgres: "inet",
		}),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("revoked_at").Optional().Nillable(),
		field.Time("expires_at"),
		field.String("superseded_by").Optional().Nillable(),
	}
}

func (Session) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Field("user_id").
			Unique().
			Required().
			Immutable(),
		edge.To("predecessor", Session.Type).
			Unique().
			From("successor").
			Unique().
			Field("superseded_by"),
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("expires_at"),
	}
}
