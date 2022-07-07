package ldcontext

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
)

// Kind is a string type set by the application to describe what kind of entity a Context
// represents. The meaning of this is completely up to the application. When no Kind is
// specified, the default is "user".
//
// For a multi-kind Context (see NewMultiBuilder), the Kind of the top-level Context is
// always "multi"; there is a specific Kind for each of the Contexts contained within it.
type Kind string

const (
	// DefaultKind is a constant for the default Kind of "user".
	DefaultKind Kind = "user"

	// MultiKind is a constant for the Kind that all multi-kind Contexts have.
	MultiKind Kind = "multi"
)

// Used internally to enforce validation and defaulting logic. Per the users-to-contexts spec,
// valid characters in "kind" are ASCII alphanumerics, period, hyphen, and underscore, it
// cannot be the string "kind", and in a single-kind context it cannot be the string "multi".
func validateSingleKind(kind Kind) (Kind, error) {
	switch kind {
	case "":
		return DefaultKind, nil

	case MultiKind:
		return "", lderrors.ErrContextKindMultiForSingleKind{}

	case Kind(ldattr.KindAttr):
		return "", lderrors.ErrContextKindCannotBeKind{}

	default:
		for _, ch := range kind {
			if (ch < 'a' || ch > 'z') && (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') &&
				ch != '.' && ch != '_' && ch != '-' {
				return "", lderrors.ErrContextKindInvalidChars{}
			}
		}
		return kind, nil
	}
}
