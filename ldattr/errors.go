package ldattr

import (
	"errors"
)

var (
	errAttributeEmpty         = errors.New("attribute reference cannot be empty")
	errAttributeExtraSlash    = errors.New("attribute reference contained a double slash or a trailing slash")
	errAttributeInvalidEscape = errors.New(
		"attribute reference contained an escape character (~) that was not followed by 0 or 1")
)
