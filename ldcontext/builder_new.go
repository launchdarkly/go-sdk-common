package ldcontext

import (
	"sort"

	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
)

type ContextBuilder struct {
	singleBuilders map[string]*Builder
}

type KindBuilder struct {
	*ContextBuilder
	singleBuilder *Builder
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{singleBuilders: make(map[string]*Builder)}
}

func (cb *ContextBuilder) Kind(kind string, key string) *KindBuilder {
	singleBuilder, ok := cb.singleBuilders[kind]
	if !ok {
		singleBuilder = NewBuilder(key)
		cb.singleBuilders[kind] = singleBuilder
	}
	singleBuilder.Key(key)
	return &KindBuilder{ContextBuilder: cb, singleBuilder: singleBuilder}
}

func (cb *ContextBuilder) Build() Context {
	if len(cb.singleBuilders) == 0 {
		return Context{defined: true, err: lderrors.ErrContextKindMultiWithNoKinds{}}
	}

	// TODO: calling `.Build()` on a Builder does mostly what we want, but it also
	// converts a "" kind to "user". This builder should instead probably enforce that you provide
	// non-empty kinds.

	if len(cb.singleBuilders) == 1 {
		// If only one context kind was added, the result is just the same as building that one context
		for _, singleBuilder := range cb.singleBuilders {
			return singleBuilder.Build()
		}
		panic("impossible")
	}

	// Sort the list by kind - this makes our output deterministic and will also be important when we
	// compute a fully qualified key.
	var kinds []string
	for kind := range cb.singleBuilders {
		kinds = append(kinds, kind)
	}
	sort.Strings(kinds)

	ret := Context{
		defined:       true,
		kind:          MultiKind,
		multiContexts: make([]Context, 0, len(cb.singleBuilders)),
	}
	for _, kind := range kinds {
		singleBuilder := cb.singleBuilders[kind]
		ctx, err := singleBuilder.TryBuild()
		if err != nil {
			ret.err = err
			return ret
		}
		ret.multiContexts = append(ret.multiContexts, ctx)
	}

	// Fully-qualified key for multi-context is defined as "kind1:key1:kind2:key2" etc., where kinds are in
	// alphabetical order (we have already sorted them above) and keys are URL-encoded. In this case we
	// do _not_ omit a default kind of "user".
	for _, c := range ret.multiContexts {
		if ret.fullyQualifiedKey != "" {
			ret.fullyQualifiedKey += ":"
		}
		ret.fullyQualifiedKey += makeFullyQualifiedKeySingleKind(c.kind, c.key, false)
	}

	return ret
}
