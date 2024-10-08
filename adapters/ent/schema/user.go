package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	mixin "github.com/notion-echo/adapters/ent/mixin"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("state_token").
			Default(""),
		field.String("notion_token").
			Default(""),
		field.String("default_page").
			Default(""),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.TimeMixin{},
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
