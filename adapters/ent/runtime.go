// Code generated by ent, DO NOT EDIT.

package ent

import (
	"github.com/notion-echo/adapters/ent/schema"
	"github.com/notion-echo/adapters/ent/user"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	userFields := schema.User{}.Fields()
	_ = userFields
	// userDescStateToken is the schema descriptor for state_token field.
	userDescStateToken := userFields[1].Descriptor()
	// user.DefaultStateToken holds the default value on creation for the state_token field.
	user.DefaultStateToken = userDescStateToken.Default.(string)
	// userDescNotionToken is the schema descriptor for notion_token field.
	userDescNotionToken := userFields[2].Descriptor()
	// user.DefaultNotionToken holds the default value on creation for the notion_token field.
	user.DefaultNotionToken = userDescNotionToken.Default.(string)
	// userDescDefaultPage is the schema descriptor for default_page field.
	userDescDefaultPage := userFields[3].Descriptor()
	// user.DefaultDefaultPage holds the default value on creation for the default_page field.
	user.DefaultDefaultPage = userDescDefaultPage.Default.(string)
}
