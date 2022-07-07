package lderrors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributeErrorMessages(t *testing.T) {
	params := []struct {
		err     error
		message string
	}{
		{ErrAttributeEmpty{}, msgAttributeEmpty},
		{ErrAttributeExtraSlash{}, msgAttributeExtraSlash},
		{ErrAttributeInvalidEscape{}, msgAttributeInvalidEscape},
	}
	for _, p := range params {
		t.Run(fmt.Sprintf("%T", p.err), func(t *testing.T) {
			assert.Equal(t, p.message, p.err.Error())
		})
	}
}
