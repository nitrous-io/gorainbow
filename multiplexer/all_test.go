package multiplexer

type closer struct {
	closed bool
	done   chan struct{}
}

func (c *closer) Close() error {
	c.closed = true
	return nil
}

func (c *closer) Wait() error {
	<-c.done
	close(c.done)
	c.Close()
	return nil
}
