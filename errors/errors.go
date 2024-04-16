package errors

import "errors"

var (
	ErrSaveNote           = errors.New("error saving note")
	ErrSearchPage         = errors.New("writing page not found")
	ErrNotEnoughArguments = errors.New("not enough arguments")
	ErrNotRegistered      = errors.New("it looks like you are not registered, try running `/register` command first")
	ErrSetDefaultPage     = errors.New("error setting default page")
	ErrPageNotFound       = errors.New("page not found")
	ErrRegistering        = errors.New("error registering")
	ErrDeleting           = errors.New("error deleting user")
	ErrStateToken         = errors.New("error generating state token, please retry later")
)
