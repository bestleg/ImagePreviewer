package utils

import (
	"encoding/hex"
	"strings"

	"github.com/OneOfOne/xxhash"
	"github.com/bestleg/ImagePreviewer/pkg/services/cache"
)

// GetHash get new hash from string.
func GetHash(text string) (cache.Key, error) {
	hasher := xxhash.New64()
	_, err := hasher.Write([]byte(text))
	if err != nil {
		return "", err
	}
	return cache.Key(hex.EncodeToString(hasher.Sum(nil))), nil
}

// Contains returns true if target string is present in the strings slice.
func Contains(slice []string, lookup string) bool {
	for _, val := range slice {
		if strings.EqualFold(val, lookup) {
			return true
		}
	}
	return false
}
