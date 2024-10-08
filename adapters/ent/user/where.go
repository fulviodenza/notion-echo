// Code generated by ent, DO NOT EDIT.

package user

import (
	"time"

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

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldUpdatedAt, v))
}

// StateToken applies equality check predicate on the "state_token" field. It's identical to StateTokenEQ.
func StateToken(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldStateToken, v))
}

// NotionToken applies equality check predicate on the "notion_token" field. It's identical to NotionTokenEQ.
func NotionToken(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldNotionToken, v))
}

// DefaultPage applies equality check predicate on the "default_page" field. It's identical to DefaultPageEQ.
func DefaultPage(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldDefaultPage, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldUpdatedAt, v))
}

// StateTokenEQ applies the EQ predicate on the "state_token" field.
func StateTokenEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldStateToken, v))
}

// StateTokenNEQ applies the NEQ predicate on the "state_token" field.
func StateTokenNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldStateToken, v))
}

// StateTokenIn applies the In predicate on the "state_token" field.
func StateTokenIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldStateToken, vs...))
}

// StateTokenNotIn applies the NotIn predicate on the "state_token" field.
func StateTokenNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldStateToken, vs...))
}

// StateTokenGT applies the GT predicate on the "state_token" field.
func StateTokenGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldStateToken, v))
}

// StateTokenGTE applies the GTE predicate on the "state_token" field.
func StateTokenGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldStateToken, v))
}

// StateTokenLT applies the LT predicate on the "state_token" field.
func StateTokenLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldStateToken, v))
}

// StateTokenLTE applies the LTE predicate on the "state_token" field.
func StateTokenLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldStateToken, v))
}

// StateTokenContains applies the Contains predicate on the "state_token" field.
func StateTokenContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldStateToken, v))
}

// StateTokenHasPrefix applies the HasPrefix predicate on the "state_token" field.
func StateTokenHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldStateToken, v))
}

// StateTokenHasSuffix applies the HasSuffix predicate on the "state_token" field.
func StateTokenHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldStateToken, v))
}

// StateTokenEqualFold applies the EqualFold predicate on the "state_token" field.
func StateTokenEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldStateToken, v))
}

// StateTokenContainsFold applies the ContainsFold predicate on the "state_token" field.
func StateTokenContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldStateToken, v))
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

// DefaultPageEQ applies the EQ predicate on the "default_page" field.
func DefaultPageEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldDefaultPage, v))
}

// DefaultPageNEQ applies the NEQ predicate on the "default_page" field.
func DefaultPageNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldDefaultPage, v))
}

// DefaultPageIn applies the In predicate on the "default_page" field.
func DefaultPageIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldDefaultPage, vs...))
}

// DefaultPageNotIn applies the NotIn predicate on the "default_page" field.
func DefaultPageNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldDefaultPage, vs...))
}

// DefaultPageGT applies the GT predicate on the "default_page" field.
func DefaultPageGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldDefaultPage, v))
}

// DefaultPageGTE applies the GTE predicate on the "default_page" field.
func DefaultPageGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldDefaultPage, v))
}

// DefaultPageLT applies the LT predicate on the "default_page" field.
func DefaultPageLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldDefaultPage, v))
}

// DefaultPageLTE applies the LTE predicate on the "default_page" field.
func DefaultPageLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldDefaultPage, v))
}

// DefaultPageContains applies the Contains predicate on the "default_page" field.
func DefaultPageContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldDefaultPage, v))
}

// DefaultPageHasPrefix applies the HasPrefix predicate on the "default_page" field.
func DefaultPageHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldDefaultPage, v))
}

// DefaultPageHasSuffix applies the HasSuffix predicate on the "default_page" field.
func DefaultPageHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldDefaultPage, v))
}

// DefaultPageEqualFold applies the EqualFold predicate on the "default_page" field.
func DefaultPageEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldDefaultPage, v))
}

// DefaultPageContainsFold applies the ContainsFold predicate on the "default_page" field.
func DefaultPageContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldDefaultPage, v))
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
