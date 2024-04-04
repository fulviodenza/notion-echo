package errors

import "errors"

var (
	ErrSaveNote           = errors.New("error saving note")
	ErrSearchPage         = errors.New("writing page not found")
	ErrNotEnoughArguments = errors.New("not enough arguments")
	ErrNotRegistered      = errors.New("it looks like you are not registered, try running `/register` command first")
	ErrSetDefaultPage     = errors.New("error setting default page")
	ErrPageNotFound       = errors.New("page not found")
)
