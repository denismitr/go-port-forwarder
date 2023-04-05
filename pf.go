package portforwarder

import (
	"context"
	"fmt"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name freePortProvider
type freePortProvider interface {
	getFreePort() (uint, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name podProvider
type podProvider interface {
	listPods(ctx context.Context, cmd *listPodsCommand) (*corev1.PodList, error)
	getPod(ctx context.Context, namespace, name string) (*corev1.Pod, error)
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

//go:generate go run github.com/vektra/mockery/v2@v2.20.2 --name connector
type connector interface {
	Connect() (*rest.Config, *kubernetes.Clientset, error)
}

type PortForwarder struct {
	restCfg          *rest.Config
	freePortProvider freePortProvider
	forwarder        portForwarder
	podProvider      podProvider
}

func NewPortForwarder(conn connector) (*PortForwarder, error) {
	restCfg, k8sClientSet, err := conn.Connect()
	if err != nil {
		return nil, err
	}

	fpp := newNetFreePortProvider("tcp", "localhost", 0)
	s := newSelectorFromKubeConfig(k8sClientSet)

	return &PortForwarder{
		restCfg:          restCfg,
		freePortProvider: fpp,
		podProvider:      s,
		forwarder:        &spdyForwarder{},
	}, nil
}

type TargetPod struct {
	// Port of the pod in k8s
	Port uint
	// Name - optional pod name, to specify the exact pod name if known
	Name string
	// Namespace to look for the suitable pod to forward
	Namespace string
	// LabelSelector to match the suitable pod to forward
	LabelSelector map[string]string
}

func (p *TargetPod) applyDefaults() {
	if p.Namespace == "" {
		p.Namespace = "default"
	}
}

func (p *TargetPod) validate() error {
	if p.Port == 0 {
		return fmt.Errorf("%w target port is required", ErrTargetPodValidation)
	}

	if p.Namespace == "" {
		return fmt.Errorf("%w namespace cannot be empty", ErrTargetPodValidation)
	}

	if p.Name == "" && len(p.LabelSelector) == 0 {
		return fmt.Errorf("%w pod name or label selector should be specified", ErrTargetPodValidation)
	}

	return nil
}

func (pf *PortForwarder) PortForwardAPod(
	ctx context.Context,
	target *TargetPod,
) (*PortForwardProcess, error) {
	target.applyDefaults()
	if err := target.validate(); err != nil {
		return nil, err
	}

	freePort, err := pf.freePortProvider.getFreePort()
	if err != nil {
		return nil, fmt.Errorf("get free port failed: %w", err)
	}

	podName, err := getPodName(ctx, pf.podProvider, target)
	if err != nil {
		return nil, fmt.Errorf("could not port forward a pod: %w", err)
	}

	process := newPortForwardProcess(ctx, freePort)
	process.wg.Add(1)
	go func(p *PortForwardProcess) {
		defer func() {
			process.wg.Done()
			p.Stop()
		}()

		if err := pf.portForwardAPod(p, target.Namespace, podName, freePort, target.Port); err != nil {
			p.setError(fmt.Errorf(
				"init port forwarder for pod %s in namespace %s failed: %w",
				podName, target.Namespace, err,
			))
		}
	}(process)

	return process, nil
}

func (pf *PortForwarder) portForwardAPod(
	process *PortForwardProcess,
	namespace,
	podName string,
	freePort uint,
	targetPort uint,
) error {
	errCh := make(chan error, 1)

	roundTripper, upgrader, err := spdy.RoundTripperFor(pf.restCfg)
	if err != nil {
		return err
	}

	serverURL := resolveServerURL(pf.restCfg.Host, namespace, podName)
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
			process.stopCh, process.startedCh,
			os.Stdout, os.Stderr, // todo: maybe log std err
		); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-process.stopCh:
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

func resolveServerURL(host, namespace, podName string) url.URL {
	path := fmt.Sprintf(
		"/api/v1/namespaces/%s/pods/%s/portforward",
		namespace, podName,
	)
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")
	serverURL := url.URL{Scheme: "https", Path: path, Host: host}
	return serverURL
}

func getPodName(
	ctx context.Context,
	provider podProvider,
	target *TargetPod,
) (string, error) {
	if target.Name != "" {
		pod, err := provider.getPod(ctx, target.Namespace, target.Name)
		if err != nil {
			return "", fmt.Errorf("%w: %s", ErrPodNotFound, err.Error())
		}
		return pod.Name, nil
	}

	pods, err := provider.listPods(ctx, &listPodsCommand{
		namespace:      target.Namespace,
		labelSelectors: target.LabelSelector,
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) < 1 {
		return "", fmt.Errorf(
			"%w: pods not found in [%s] namespace with provider %+v",
			ErrPodNotFound, target.Namespace, target.LabelSelector,
		)
	}

	podName := pods.Items[0].GetName()
	if podName == "" {
		return "", fmt.Errorf("%w: pod name should not be empty", ErrPodNotFound)
	}

	return podName, nil
}

type spdyForwarder struct{}

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
