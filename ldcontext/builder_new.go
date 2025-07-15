package ldcontext

import (
	"sort"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

type ContextBuilder struct {
	singleBuilders map[Kind]*Builder
}

type KindBuilder struct {
	*ContextBuilder
	singleBuilder *Builder
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{singleBuilders: make(map[Kind]*Builder)}
}

func (cb *ContextBuilder) Kind(kind Kind, key string) *KindBuilder {
	if kind == "" {
		kind = DefaultKind
	}
	singleBuilder, ok := cb.singleBuilders[kind]
	if !ok {
		singleBuilder = NewBuilder(key).Kind(kind)
		cb.singleBuilders[kind] = singleBuilder
	}
	singleBuilder.Key(key)
	return &KindBuilder{ContextBuilder: cb, singleBuilder: singleBuilder}
}

func (cb *ContextBuilder) Build() Context {
	if len(cb.singleBuilders) == 0 {
		return Context{defined: true, err: lderrors.ErrContextKindMultiWithNoKinds{}}
	}

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
		kinds = append(kinds, string(kind))
	}
	sort.Strings(kinds)

	ret := Context{
		defined:       true,
		kind:          MultiKind,
		multiContexts: make([]Context, 0, len(cb.singleBuilders)),
	}

	var individualErrors map[string]error
	for _, kind := range kinds {
		singleBuilder := cb.singleBuilders[Kind(kind)]
		ctx, err := singleBuilder.TryBuild()
		if err != nil {
			if individualErrors == nil {
				individualErrors = make(map[string]error)
			}
			individualErrors[kind] = err
			continue
		}
		ret.multiContexts = append(ret.multiContexts, ctx)
	}
	if len(individualErrors) != 0 {
		ret.err = lderrors.ErrContextPerKindErrors{Errors: individualErrors}
		return ret
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

func (cb *ContextBuilder) TryBuild() (Context, error) {
	c := cb.Build()
	return c, c.Err()
}

func (kb *KindBuilder) Key(key string) *KindBuilder {
	kb.singleBuilder.Key(key)
	return kb
}

func (kb *KindBuilder) Name(name string) *KindBuilder {
	kb.singleBuilder.Name(name)
	return kb
}

func (kb *KindBuilder) OptName(name ldvalue.OptionalString) *KindBuilder {
	kb.singleBuilder.OptName(name)
	return kb
}

func (kb *KindBuilder) SetBool(attributeName string, value bool) *KindBuilder {
	kb.singleBuilder.SetBool(attributeName, value)
	return kb
}

func (kb *KindBuilder) SetFloat64(attributeName string, value float64) *KindBuilder {
	kb.singleBuilder.SetFloat64(attributeName, value)
	return kb
}

func (kb *KindBuilder) SetInt(attributeName string, value int) *KindBuilder {
	kb.singleBuilder.SetInt(attributeName, value)
	return kb
}

func (kb *KindBuilder) SetString(attributeName string, value string) *KindBuilder {
	kb.singleBuilder.SetString(attributeName, value)
	return kb
}

func (kb *KindBuilder) SetValue(attributeName string, value ldvalue.Value) *KindBuilder {
	kb.singleBuilder.SetValue(attributeName, value)
	return kb
}

func (kb *KindBuilder) TrySetValue(attributeName string, value ldvalue.Value) bool {
	return kb.singleBuilder.TrySetValue(attributeName, value)
}

func (kb *KindBuilder) Anonymous(value bool) *KindBuilder {
	kb.singleBuilder.Anonymous(value)
	return kb
}

func (kb *KindBuilder) Private(attrRefStrings ...string) *KindBuilder {
	kb.singleBuilder.Private(attrRefStrings...)
	return kb
}

func (kb *KindBuilder) PrivateRef(attrRefs ...ldattr.Ref) *KindBuilder {
	kb.singleBuilder.PrivateRef(attrRefs...)
	return kb
}

func (kb *KindBuilder) RemovePrivate(attrRefStrings ...string) *KindBuilder {
	kb.singleBuilder.RemovePrivate(attrRefStrings...)
	return kb
}

func (kb *KindBuilder) RemovePrivateRef(attrRefs ...ldattr.Ref) *KindBuilder {
	kb.singleBuilder.RemovePrivateRef(attrRefs...)
	return kb
}
