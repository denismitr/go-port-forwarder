package portforwarder

import (
	"context"
	"sync"
)

type PortForwardProcess struct {
	Port    uint
	err     error
	readyCh chan struct{}
	doneCh  chan struct{}
	stopCh  chan struct{}
	stopper sync.Once
	mx      sync.Mutex
	wg      sync.WaitGroup
}

func newPortForwardProcess(ctx context.Context, port uint) *PortForwardProcess {
	p := &PortForwardProcess{
		Port:    port,
		readyCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
		stopCh:  make(chan struct{}),
	}

	go func() {
		select {
		case <-ctx.Done():
			p.setError(ctx.Err())
			p.Stop()
			return
		case <-p.doneCh:
			return
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

func (p *PortForwardProcess) Ready() <-chan struct{} {
	return p.readyCh
}

func (p *PortForwardProcess) Done() <-chan struct{} {
	return p.doneCh
}

func (p *PortForwardProcess) Err() error {
	p.mx.Lock()
	defer p.mx.Unlock()
	return p.err
}

func (p *PortForwardProcess) setError(err error) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.err = err
}

func (p *PortForwardProcess) markAsReady() {
	close(p.readyCh)
}
