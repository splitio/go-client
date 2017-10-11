package hash

// Legacy calculates the bucket for the key and seed provided using the legacy algorithm
func Legacy(key string, seed int32) int32 {
	var h int32
	for _, char := range key {
		h = 31*h + int32(char)
	}
	return int32(h ^ seed)
}
