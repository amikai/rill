package th

import "sync"

type InProgressCounter struct {
	mu      sync.Mutex
	current int
	max     int
}

func (c *InProgressCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.current++
	if c.max < c.current {
		c.max = c.current
	}
}

func (c *InProgressCounter) Dec() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.current--
}

func (c *InProgressCounter) Current() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.current
}

func (c *InProgressCounter) Max() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.max
}
