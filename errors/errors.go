package errors

import "errors"

var (
	ErrSaveNote           = errors.New("error saving note")
	ErrSearchPage         = errors.New("writing page not found")
	ErrNotEnoughArguments = errors.New("not enough arguments")
	ErrNotRegistered      = errors.New("it looks like you are not registered, try running `/register` command first")
	ErrTokenNotFound      = errors.New("internal error")
	ErrSetDefaultPage     = errors.New("error setting default page")
	ErrPageNotFound       = errors.New("page not found, please run /defaultpage command first")
	ErrRegistering        = errors.New("error registering")
	ErrDeleting           = errors.New("error deleting user")
	ErrStateToken         = errors.New("error generating state token, please retry later")
	ErrInternal           = errors.New("it seems we are having internal troubles, please come back later")
	ErrBotNotAuthorized   = errors.New("it looks like I don't have access to any page, ensure running the command /register and select the pages you want me to access to")
)
