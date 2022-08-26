package hash

import (
	"github.com/splitio/go-toolkit/v5/hasher"
)

// Murmur3_32 is a wrapper to the hash implementation in go-toolkit
func Murmur3_32(data []byte, seed uint32) uint32 {
	h := hasher.NewMurmur332Hasher(seed)
	return h.Hash(data)
}


