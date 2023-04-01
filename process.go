package portforwarder

import (
	"context"
	"sync"
)

type PortForwardProcess struct {
	Port    uint
	err     error
	doneCh  chan struct{}
	stopCh  chan struct{}
	stopper sync.Once
	mx      sync.Mutex
	wg      sync.WaitGroup
}

func newPortForwardProcess(ctx context.Context, port uint) *PortForwardProcess {
	p := &PortForwardProcess{
		Port:   port,
		doneCh: make(chan struct{}),
		stopCh: make(chan struct{}),
	}

	go func() {
		select {
		case <-ctx.Done():
			p.setError(ctx.Err())
			p.Stop()
		case <-p.doneCh:
		}
	}()

	return p
}

func (p *PortForwardProcess) Stop() {
	p.stopper.Do(func() {
		close(p.stopCh)
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
