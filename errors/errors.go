package errors

import "errors"

var (
	ErrSaveNote           = errors.New("error saving note")
	ErrSearchPage         = errors.New("writing page not found")
	ErrNotEnoughArguments = errors.New("not enough arguments")
	ErrSetDefaultPage     = errors.New("error setting default page")
)
