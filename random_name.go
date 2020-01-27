package occam

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid"
)

func RandomName() (string, error) {
	now := time.Now()
	timestamp := ulid.Timestamp(now)
	entropy := ulid.Monotonic(rand.Reader, 0)

	guid, err := ulid.New(timestamp, entropy)
	if err != nil {
		return "", err
	}

	return guid.String(), nil
}
