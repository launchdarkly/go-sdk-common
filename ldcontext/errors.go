package ldcontext

import (
	"errors"
)

var (
	errContextUninitialized              = errors.New("tried to use uninitialized Context")
	errContextKeyEmpty                   = errors.New("context key must not be empty")
	errContextKindCannotBeKind           = errors.New(`"kind" is not a valid context kind`)
	errContextKindMultiWithSimpleBuilder = errors.New(`context of kind "multi" must be built with NewMultiBuilder`)
	errContextKindMultiWithNoKinds       = errors.New("multi-kind context must contain at least one kind")
	errContextKindMultiWithinMulti       = errors.New("multi-kind context cannot contain other multi-kind contexts")
	errContextKindMultiDuplicates        = errors.New("multi-kind context cannot have same kind more than once")
	errContextKindInvalidChars           = errors.New("context kind contains disallowed characters")
)
