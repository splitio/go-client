package engine

import (
	"github.com/splitio/go-client/splitio/engine/grammar"
	"github.com/splitio/go-client/splitio/engine/hash"
	"math"
	"testing"
)

func TestProperHashFunctionIsUsed(t *testing.T) {
	eng := Engine{}

	murmurHash := hash.Murmur3_32([]byte("SOME_TEST"), 12345)
	murmurBucket := int(math.Abs(float64(murmurHash%100)) + 1)
	if murmurBucket != eng.calculateBucket(2, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}

	legacyHash := hash.Legacy([]byte("SOME_TEST"), 12345)
	legacyBucket := int(math.Abs(float64(legacyHash%100)) + 1)
	if legacyBucket != eng.calculateBucket(1, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}
}

func TestProperHashFunctionIsUsedWithConstants(t *testing.T) {
	eng := Engine{}

	murmurHash := hash.Murmur3_32([]byte("SOME_TEST"), 12345)
	murmurBucket := int(math.Abs(float64(murmurHash%100)) + 1)
	if murmurBucket != eng.calculateBucket(grammar.SplitAlgoMurmur, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}

	legacyHash := hash.Legacy([]byte("SOME_TEST"), 12345)
	legacyBucket := int(math.Abs(float64(legacyHash%100)) + 1)
	if legacyBucket != eng.calculateBucket(grammar.SplitAlgoLegacy, "SOME_TEST", 12345) {
		t.Error("Incorrect hash!")
	}
}
