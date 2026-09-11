package utils

import (
	"crypto/rand"
	"fmt"
)

// NewUUID generates a random RFC 4122 version 4 UUID string,
// e.g. "3fa85f64-5717-4562-b3fc-2c963f66afa6".
// Implemented with only the standard library so we don't need to
// add a new dependency (like google/uuid) just for this one field.
func NewUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)

	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
