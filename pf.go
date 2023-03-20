package portforwarder

import (
	"context"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name freePortProvider
type freePortProvider interface {
	getFreePort() (uint, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name podProvider
type podProvider interface {
	listPods(context.Context, *listPodsCommand) (*v1.PodList, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name portForwarder
type portForwarder interface {
	forward(
		dialer httpstream.Dialer,
		ports []string,
		stopChan <-chan struct{},
		readyChan chan struct{},
		out,
		errOut io.Writer,
	) error
}

type (
	PortForwardProcess struct {
		Port   uint
		err    error
		doneCh chan struct{}
		stop   sync.Once
		mx     sync.Mutex
	}

	PortForwarder struct {
		k8sCfg           *rest.Config
		freePortProvider freePortProvider
		forwarder        portForwarder
		podLister        podProvider
	}
)

func NewPortForwarder(masterURL string, kubeConfig string) (*PortForwarder, error) {
	k8sCfg, err := parseKubeConfig(
		masterURL, kubeConfig,
	)
	if err != nil {
		return nil, err
	}

	k8sClientSet, err := createK8SClientSet(k8sCfg)
	if err != nil {
		return nil, err
	}

	fpp := newNetFreePortProvider("tcp", "localhost", 0)
	s := newSelectorFromKubeConfig(k8sClientSet)

	return &PortForwarder{
		k8sCfg:           k8sCfg,
		freePortProvider: fpp,
		podLister:        s,
		forwarder:        &spdyForwarder{},
	}, nil
}

func (p *PortForwardProcess) setError(err error) {
	p.mx.Lock()
	defer p.mx.Unlock()
	p.err = err
}

func (p *PortForwardProcess) Stop() {
	p.stop.Do(func() {
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

func (pf *PortForwarder) PortForwardAPod(
	ctx context.Context,
	targetPort uint,
	namespace string,
	podLabelSelector map[string]string,
) (*PortForwardProcess, error) {
	freePort, err := pf.freePortProvider.getFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port failed: %w", err)
	}

	podName, err := pf.getPodName(ctx, namespace, podLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("could not port forward a pod: %w", err)
	}

	process := &PortForwardProcess{Port: freePort, doneCh: make(chan struct{}, 1)}
	go func(p *PortForwardProcess) {
		defer p.Stop()
		if err := pf.portForwardAPod(ctx, p.doneCh, namespace, podName, freePort, targetPort); err != nil {
			p.setError(fmt.Errorf(
				"init port forwarder for pod %s in namespace %s failed: %w",
				podName, namespace, err,
			))
		}
	}(process)

	return process, nil
}

func (pf *PortForwarder) portForwardAPod(
	ctx context.Context,
	doneCh chan struct{},
	namespace,
	podName string,
	freePort uint,
	targetPort uint,
) error {
	readyCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	roundTripper, upgrader, err := spdy.RoundTripperFor(pf.k8sCfg)
	if err != nil {
		return err
	}

	path := fmt.Sprintf(
		"/api/v1/namespaces/%s/pods/%s/portforward",
		namespace, podName,
	)
	hostIP := strings.TrimLeft(pf.k8sCfg.Host, "https://")
	serverURL := url.URL{Scheme: "https", Path: path, Host: hostIP}

	dialer := spdy.NewDialer(
		upgrader,
		&http.Client{Transport: roundTripper},
		http.MethodPost,
		&serverURL,
	)

	go func() {
		defer close(errCh)
		if err = pf.forwarder.forward(
			dialer,
			[]string{fmt.Sprintf("%d:%d", freePort, targetPort)},
			doneCh, readyCh,
			os.Stdout, os.Stderr, // todo: maybe log std err
		); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-readyCh:
		return nil
	case pfErr, ok := <-errCh:
		if ok {
			return fmt.Errorf(
				"pod %s port %d forward error in namespace %s: %w",
				podName, freePort, namespace, pfErr,
			)
		}
		return nil
	}
}

func (pf *PortForwarder) getPodName(
	ctx context.Context,
	namespace string,
	podLabelSelector map[string]string,
) (string, error) {
	pods, err := pf.podLister.listPods(ctx, &listPodsCommand{
		namespace:      namespace,
		labelSelectors: podLabelSelector,
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) < 1 {
		return "", fmt.Errorf(
			"pods not found in [%s] namespace with selector %+v",
			namespace, podLabelSelector,
		)
	}

	podName := pods.Items[0].GetName()
	if podName == "" {
		return "", fmt.Errorf("pod name should not be empty")
	}

	return podName, nil
}

type spdyForwarder struct {
}

func (f *spdyForwarder) forward(
	dialer httpstream.Dialer,
	ports []string,
	stopChan <-chan struct{},
	readyChan chan struct{},
	out,
	errOut io.Writer,
) error {
	forwarder, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return err
	}

	return forwarder.ForwardPorts()
}
