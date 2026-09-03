package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }

type IDs struct{}

func (IDs) NewID() string {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		panic("runtime: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(value[:])
}
