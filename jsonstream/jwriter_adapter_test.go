package jsonstream

import (
	"errors"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-jsonstream/jwriter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBufferAdapterFunc func(*jwriter.Writer)

func TestWriteToJSONBufferThroughWriter(t *testing.T) {
	marshalFunc := testBufferAdapterFunc(func(w *jwriter.Writer) {
		arr := w.Array()
		for i := 0; i < 100; i++ {
			arr.String(fmt.Sprintf("value%d", i))
		}
		arr.End()
	})

	var jb JSONBuffer
	WriteToJSONBufferThroughWriter(marshalFunc, &jb)
	require.NoError(t, jb.GetError())

	jw := jwriter.NewWriter()
	marshalFunc(&jw)

	bytes1, err := jb.Get()
	require.NoError(t, err)
	bytes2 := jw.Bytes()
	assert.Equal(t, string(bytes1), string(bytes2))
}

func TestWriteToJSONBufferThroughWriterCopiesErrorState(t *testing.T) {
	fakeError := errors.New("sorry")
	marshalFunc := testBufferAdapterFunc(func(w *jwriter.Writer) {
		w.AddError(fakeError)
	})

	var jb JSONBuffer
	WriteToJSONBufferThroughWriter(marshalFunc, &jb)
	require.Equal(t, fakeError, jb.GetError())
}

func (t testBufferAdapterFunc) WriteToJSONWriter(w *jwriter.Writer) {
	t(w)
}
