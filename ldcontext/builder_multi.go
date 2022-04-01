package ldcontext

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

const defaultMultiBuilderCapacity = 3 // arbitrary value based on presumed likely use cases

// MultiBuilder is a mutable object that uses the builder pattern to specify properties for a Context.
//
// Use this type if you need to construct a Context that has multiple Kind values, each with its
// own nested Context. To define a single-kind context, use Builder instead.
//
// Obtain an instance of MultiBuilder by calling NewMultiBuilder; then, call Add to specify the
// nested Context for each Kind. MultiBuilder setters return a reference the same builder, so they
// can be chained together:
//
//     context := ldcontext.NewMultiBuilder().
//         Add(ldcontext.New("my-user-key")).
//         Add(ldcontext.NewBuilder("my-org-key").Kind("organization").Name("Org1").Build()).
//         Build()
//
// A MultiBuilder should not be accessed by multiple goroutines at once. Once you have called Build(),
// the resulting Context is immutable and is safe to use from multiple goroutines. Instances
// created with Build() are not affected by subsequent actions taken on the MultiBuilder.
type MultiBuilder struct {
	contexts            []Context
	contextsCopyOnWrite bool
}

// NewMultiBuilder creates a MultiBuilder for building a Context.
//
// This method is for building a Context athat has multiple Kind values, each with its own
// nested Context. To define a single-kind context, use NewBuilder() instead.
func NewMultiBuilder() *MultiBuilder {
	return &MultiBuilder{contexts: make([]Context, 0, defaultMultiBuilderCapacity)}
}

// Build creates a Context from the current MultiBuilder properties.
//
// The Context is immutable and will not be affected by any subsequent actions on the MultiBuilder.
//
// It is possible for a MultiBuilder to represent an invalid state. Instead of returning two
// values (Context, error), the Builder always returns a Context and you can call Context.Err()
// to see if it has an error. See Context.Err() for more information about invalid Context
// conditions. Using a single-return-value syntax is more convenient for application code, since
// in normal usage an application will never build an invalid Context.
//
// If only one context kind was added to the builder, Build returns a single-kind context rather
// than a multi-kind context.
func (m *MultiBuilder) Build() Context {
	if len(m.contexts) == 0 {
		return Context{err: errContextKindMultiWithNoKinds}
	}

	if len(m.contexts) == 1 {
		// Never return a multi-kind context with just one kind; instead return the individual one
		c := m.contexts[0]
		if c.Multiple() {
			return Context{err: errContextKindMultiWithinMulti}
		}
		return c
	}

	m.contextsCopyOnWrite = true // see note on ___CopyOnWrite in Builder.Build()

	// Sort the list by kind - this makes our output deterministic and will also be important when we
	// compute a fully qualified key.
	sort.Slice(m.contexts, func(i, j int) bool { return m.contexts[i].Kind() < m.contexts[j].Kind() })

	// Check for conditions that could make a multi-kind context invalid
	var errs []string
	nestedMulti := false
	duplicates := false
	for i, c := range m.contexts {
		err := c.Err()
		switch {
		case err != nil: // one of the individual contexts already had an error
			errs = append(errs, fmt.Sprintf("(%s) %s", c.Kind(), err.Error()))
		case c.Multiple(): // multi-kind inside multi-kind is not allowed
			nestedMulti = true
		default:
			for j := 0; j < i; j++ {
				if c.Kind() == m.contexts[j].Kind() { // same kind was seen twice
					duplicates = true
					break
				}
			}
		}
	}
	if nestedMulti {
		errs = append(errs, errContextKindMultiWithinMulti.Error())
	}
	if duplicates {
		errs = append(errs, errContextKindMultiDuplicates.Error())
	}
	if len(errs) != 0 {
		return Context{
			err: errors.New(strings.Join(errs, ", ")),
		}
	}

	ret := Context{
		defined:       true,
		kind:          MultiKind,
		multiContexts: m.contexts,
	}

	// Fully-qualified key for multi-kind is defined as "kind1:key1:kind2:key2" etc., where kinds are in
	// alphabetical order (we have already sorted them above) and keys are URL-encoded. In this case we
	// do _not_ omit a default kind of "user".
	for _, c := range m.contexts {
		if ret.fullyQualifiedKey != "" {
			ret.fullyQualifiedKey += ":"
		}
		ret.fullyQualifiedKey += makeFullyQualifiedKeySingleKind(c.kind, c.key, false)
	}

	return ret
}

// TryBuild is an alternative to Build that returns any validation errors as a second value.
//
// As described in Build(), there are several ways the state of a Context could be invalid.
// Since in normal usage it is possible to be confident that these will not occur, the Build()
// method is designed for convenient use within expressions by returning a single Context
// value, and any validation problems are contained within that value where they can be
// detected by calling the context's Err() method. But, if you prefer to use the two-value
// pattern that is common in Go, you can call TryBuild instead:
//
//     c, err := ldcontext.NewMultiBuilder().
//         Add(context1).Add(context2).
//         TryBuild()
//     if err != nil {
//         // do whatever is appropriate if building the context failed
//     }
//
// The two return values are the same as to 1. the Context that would be returned by Build(),
// and 2. the result of calling Err() on that Context. So, the above example is exactly
// equivalent to:
//
//     c := ldcontext.NewMultiBuilder().
//         Add(context1).Add(context2).
//         Build()
//     if c.Err() != nil {
//         // do whatever is appropriate if building the context failed
//     }
//
// Note that unlike some Go methods where the first return value is normally an
// uninitialized zero value if the error is non-nil, the Context returned by TryBuild in case
// of an error is not completely uninitialized: it does contain the error information as well,
// so that if it is mistakenly passed to an SDK method, the SDK can tell what the error was.
func (m *MultiBuilder) TryBuild() (Context, error) {
	c := m.Build()
	return c, c.Err()
}

// Add adds a nested context for a specific Kind to a MultiBuilder.
//
// It is invalid to add more than one context with the same Kind. This error is detected
// when you call Build() or TryBuild().
func (m *MultiBuilder) Add(context Context) *MultiBuilder {
	if m.contextsCopyOnWrite {
		m.contexts = append([]Context(nil), m.contexts...)
		m.contextsCopyOnWrite = true
	}
	m.contexts = append(m.contexts, context)
	return m
}
