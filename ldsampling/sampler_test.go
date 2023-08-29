package ldsampling

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRatioSampler(t *testing.T) {
	t.Run("simple sampling", func(t *testing.T) {
		sampler := NewSampler()

		assert.False(t, sampler.Sample(-1), "negative ratio should sample false")
		assert.False(t, sampler.Sample(0), "zero ratio should sample false")
		assert.True(t, sampler.Sample(1), "one ratio should sample true")
	})

	t.Run("random sampling", func(t *testing.T) {
		sampler := NewSamplerFromSource(rand.NewSource(1))

		picks := 0
		for i := 0; i < 1_000; i++ {
			if sampler.Sample(2) {
				picks += 1
			}
		}

		// This isn't a perfect 1 in 2 ratio, but the sampling is
		// probabilistic. Since we control the seed, this should be safe for
		// testing.
		assert.Equal(t, 508, picks)
	})
}
