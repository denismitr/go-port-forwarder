package portforwarder

import (
	"context"
	"sync"
)

type PortForwardProcess struct {
	Port       uint
	err        error
	startedCh  chan struct{}
	finishedCh chan struct{}
	stopCh     chan struct{}
	stopper    sync.Once
	mx         sync.Mutex
	wg         sync.WaitGroup
}

func newPortForwardProcess(ctx context.Context, port uint) *PortForwardProcess {
	p := &PortForwardProcess{
		Port:       port,
		startedCh:  make(chan struct{}),
		finishedCh: make(chan struct{}),
		stopCh:     make(chan struct{}),
	}

	go func() {
		select {
		case <-ctx.Done():
			p.setError(ctx.Err())
			p.Stop()
			return
		case <-p.finishedCh:
			return
		}
	}()

	return p
}

func (p *PortForwardProcess) Stop() {
	p.stopper.Do(func() {
		close(p.stopCh)
		p.wg.Wait()
		close(p.finishedCh)
	})
}

// Started signals that port forward has started
func (p *PortForwardProcess) Started() <-chan struct{} {
	return p.startedCh
}

// Finished signals that port forward has finished
func (p *PortForwardProcess) Finished() <-chan struct{} {
	return p.finishedCh
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
	close(p.startedCh)
}
