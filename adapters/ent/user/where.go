// Code generated by ent, DO NOT EDIT.

package user

import (
	"entgo.io/ent/dialect/sql"
	"github.com/notion-echo/adapters/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id int) predicate.User {
	return predicate.User(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id int) predicate.User {
	return predicate.User(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id int) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...int) predicate.User {
	return predicate.User(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...int) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id int) predicate.User {
	return predicate.User(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id int) predicate.User {
	return predicate.User(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id int) predicate.User {
	return predicate.User(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id int) predicate.User {
	return predicate.User(sql.FieldLTE(FieldID, id))
}

// NotionToken applies equality check predicate on the "notion_token" field. It's identical to NotionTokenEQ.
func NotionToken(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldNotionToken, v))
}

// NotionTokenEQ applies the EQ predicate on the "notion_token" field.
func NotionTokenEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldNotionToken, v))
}

// NotionTokenNEQ applies the NEQ predicate on the "notion_token" field.
func NotionTokenNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldNotionToken, v))
}

// NotionTokenIn applies the In predicate on the "notion_token" field.
func NotionTokenIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldNotionToken, vs...))
}

// NotionTokenNotIn applies the NotIn predicate on the "notion_token" field.
func NotionTokenNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldNotionToken, vs...))
}

// NotionTokenGT applies the GT predicate on the "notion_token" field.
func NotionTokenGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldNotionToken, v))
}

// NotionTokenGTE applies the GTE predicate on the "notion_token" field.
func NotionTokenGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldNotionToken, v))
}

// NotionTokenLT applies the LT predicate on the "notion_token" field.
func NotionTokenLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldNotionToken, v))
}

// NotionTokenLTE applies the LTE predicate on the "notion_token" field.
func NotionTokenLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldNotionToken, v))
}

// NotionTokenContains applies the Contains predicate on the "notion_token" field.
func NotionTokenContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldNotionToken, v))
}

// NotionTokenHasPrefix applies the HasPrefix predicate on the "notion_token" field.
func NotionTokenHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldNotionToken, v))
}

// NotionTokenHasSuffix applies the HasSuffix predicate on the "notion_token" field.
func NotionTokenHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldNotionToken, v))
}

// NotionTokenEqualFold applies the EqualFold predicate on the "notion_token" field.
func NotionTokenEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldNotionToken, v))
}

// NotionTokenContainsFold applies the ContainsFold predicate on the "notion_token" field.
func NotionTokenContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldNotionToken, v))
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.User) predicate.User {
	return predicate.User(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.User) predicate.User {
	return predicate.User(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.User) predicate.User {
	return predicate.User(sql.NotPredicates(p))
}
