package portforwarder

import (
	"context"
	"sync"
)

type PortForwardProcess struct {
	Port   uint
	err    error
	doneCh chan struct{}
	stop   sync.Once
	mx     sync.Mutex
	wg     sync.WaitGroup
	cancel context.CancelFunc
}

func (p *PortForwardProcess) Stop() {
	p.stop.Do(func() {
		p.cancel()
		p.wg.Wait()
		close(p.doneCh)
	})
}

func (p *PortForwardProcess) Done() <-chan struct{} {
	return p.doneCh
}

func (p *PortForwardProcess) Err() error {
	p.mx.Lock()
	defer p.mx.Unlock()
	return p.err
}
