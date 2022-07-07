package lderrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextErrorMessages(t *testing.T) {
	params := []struct {
		err     error
		message string
	}{
		{ErrContextUninitialized{}, msgContextUninitialized},
		{ErrContextKeyEmpty{}, msgContextKeyEmpty},
		{ErrContextKeyNull{}, msgContextKeyNull},
		{ErrContextKeyMissing{}, msgContextKeyMissing},
		{ErrContextKindEmpty{}, msgContextKindEmpty},
		{ErrContextKindCannotBeKind{}, msgContextKindCannotBeKind},
		{ErrContextKindMultiForSingleKind{}, msgContextKindMultiForSingleKind},
		{ErrContextKindMultiWithNoKinds{}, msgContextKindMultiWithNoKinds},
		{ErrContextKindMultiWithinMulti{}, msgContextKindMultiWithinMulti},
		{ErrContextKindMultiDuplicates{}, msgContextKindMultiDuplicates},
		{ErrContextKindInvalidChars{}, msgContextKindInvalidChars},
	}
	for _, p := range params {
		t.Run(fmt.Sprintf("%T", p.err), func(t *testing.T) {
			assert.Equal(t, p.message, p.err.Error())
		})
	}

	t.Run("ErrContextPerKindErrors", func(t *testing.T) {
		e1 := ErrContextPerKindErrors{
			Errors: map[string]error{
				"kind1": ErrContextKeyEmpty{},
			},
		}
		assert.Equal(t, "(kind1) "+ErrContextKeyEmpty{}.Error(), e1.Error())

		e2 := ErrContextPerKindErrors{
			Errors: map[string]error{
				"kind1": ErrContextKeyEmpty{},
				"kind2": ErrContextKeyNull{},
			},
		}
		assert.Equal(t, "(kind1) "+ErrContextKeyEmpty{}.Error()+", (kind2) "+ErrContextKeyNull{}.Error(), e2.Error())
	})
}
