package ldcontext

// New creates a single-kind Context with a Kind of DefaultKind and the specified key.
//
// To specify additional properties, use NewBuilder(). To create a multi-kind Context, use
// NewMulti() or NewMultiBuilder(). To create a single-kind Context of a different kind than
// "user", use NewWithKind().
func New(key string) Context {
	return NewWithKind(DefaultKind, key)
}

// NewWithKind creates a single-kind Context with only the Kind and Key properties specified.
//
// To specify additional properties, use NewBuilder(). To create a multi-kind Context, use
// NewMulti() or NewMultiBuilder().
func NewWithKind(kind Kind, key string) Context {
	// Here we'll use Builder rather than directly constructing the Context struct. That
	// allows us to take advantage of logic in Builder like the setting of FullyQualifiedKey.
	// We avoid the heap allocation overhead of NewBuilder by declaring a Builder locally.
	var b Builder
	b.Kind(kind)
	b.Key(key)
	return b.Build()
}

// NewMulti creates a multi-kind Context out of the specified single-kind Contexts.
//
// To create a single-kind Context, use New(), NewWithKind, or NewBuilder().
//
// For the returned Context to be valid, the contexts list must not be empty, and all of its
// elements must be single-kind Contexts. Otherwise, the returned Context will be invalid as
// reported by Context.Err().
func NewMulti(contexts ...Context) Context {
	// Same rationale as for New/NewWithKey of using the builder instead of constructing directly.
	var m MultiBuilder
	for _, c := range contexts {
		m.Add(c)
	}
	return m.Build()
}
