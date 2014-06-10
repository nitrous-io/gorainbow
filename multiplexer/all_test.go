package multiplexer

type closer struct {
	closed bool
}

func (c *closer) Close() error {
	c.closed = true
	return nil
}
