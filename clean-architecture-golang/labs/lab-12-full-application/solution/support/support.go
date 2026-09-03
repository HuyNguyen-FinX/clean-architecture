package support

import (
	"fmt"
	"sync"
	"time"
)

type Clock struct{ Value time.Time }

func (c Clock) Now() time.Time { return c.Value }

type IDs struct {
	mu   sync.Mutex
	next int
}

func (g *IDs) NewID() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.next++
	return fmt.Sprintf("ID-%06d", g.next)
}
