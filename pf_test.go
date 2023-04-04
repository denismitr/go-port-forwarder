package portforwarder

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"testing"
	"time"
)

func TestNewPortForwarder(t *testing.T) {
	t.Run("valid connector", func(t *testing.T) {
		mc := newMockConnector(t)
		mc.EXPECT().Connect().Times(1).Return(&rest.Config{}, &kubernetes.Clientset{}, nil)
		pf, err := NewPortForwarder(mc)
		assert.NoError(t, err)
		assert.NotNil(t, pf)

		assert.NotNil(t, pf.forwarder)
		assert.NotNil(t, pf.restCfg)
		assert.NotNil(t, pf.freePortProvider)
		assert.NotNil(t, pf.podLister)
	})

	t.Run("errored connector", func(t *testing.T) {
		mc := newMockConnector(t)
		mc.EXPECT().Connect().Times(1).Return(nil, nil, errors.New("connector error"))
		pf, err := NewPortForwarder(mc)
		require.Error(t, err)
		assert.Nil(t, pf)
		assert.Equal(t, "connector error", err.Error())
	})
}

func TestPortForwarder_PortForwardAPod(t *testing.T) {
	t.Run("forward a pod can be stopped", func(t *testing.T) {
		restCfg := &rest.Config{}
		ctx := context.TODO()
		namespace := "kafka-ns"
		ls := map[string]string{"app": "foo"}
		pod := v1.Pod{}
		pod.Name = "kafka-pod-0"

		fpp := newMockFreePortProvider(t)
		fpp.EXPECT().getFreePort().Times(1).Return(uint(3999), nil)

		pl := newMockPodProvider(t)
		pl.EXPECT().
			listPods(ctx, &listPodsCommand{
				namespace:      namespace,
				labelSelectors: ls,
				fieldSelectors: map[string]string(nil),
			}).
			Times(1).
			Return(&v1.PodList{Items: []v1.Pod{pod}}, nil)

		f := newMockPortForwarder(t)
		f.EXPECT().
			forward(mock.Anything, []string{"3999:3000"}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Times(1).
			Return(nil)

		pf := &PortForwarder{
			freePortProvider: fpp,
			podLister:        pl,
			forwarder:        f,
			restCfg:          restCfg,
		}

		process, err := pf.PortForwardAPod(
			context.TODO(),
			&TargetPod{
				Port:          3000,
				Namespace:     namespace,
				LabelSelector: ls,
			},
		)

		go func() {
			<-time.After(500 * time.Millisecond)
			process.Stop()
		}()

		require.NoError(t, err)
		<-process.Finished()
		assert.NoError(t, process.Err())
	})
}
