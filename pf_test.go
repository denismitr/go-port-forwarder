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
	"net/url"
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
		assert.NotNil(t, pf.podProvider)
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
			podProvider:      pl,
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

func Test_getPodName(t *testing.T) {
	t.Run("find by namespace and provider", func(t *testing.T) {
		ctx := context.TODO()
		namespace := "kafka-ns"
		ls := map[string]string{"app": "foo"}
		pod := v1.Pod{}
		pod.Name = "kafka-pod-0"

		pl := newMockPodProvider(t)
		pl.EXPECT().
			listPods(ctx, &listPodsCommand{
				namespace:      namespace,
				labelSelectors: ls,
				fieldSelectors: map[string]string(nil),
			}).
			Times(1).
			Return(&v1.PodList{Items: []v1.Pod{pod}}, nil)

		podName, err := getPodName(ctx, pl, &TargetPod{
			Namespace:     namespace,
			LabelSelector: ls,
		})
		require.NoError(t, err)
		assert.Equal(t, "kafka-pod-0", podName)
	})
}

func Test_resolveServerURL(t *testing.T) {
	type args struct {
		host      string
		namespace string
		podName   string
	}
	tests := []struct {
		name string
		args args
		want url.URL
	}{
		{
			name: "https local host",
			args: args{
				host:      "https://127.0.0.1:5545",
				namespace: "nginx-ns",
				podName:   "nginx",
			},
			want: resolveTestURL(t, "https://127.0.0.1:5545/api/v1/namespaces/nginx-ns/pods/nginx/portforward"),
		},
		{
			name: "https local host",
			args: args{
				host:      "http://some-cluster:99882",
				namespace: "my-namespace",
				podName:   "kafka-broker-0",
			},
			want: resolveTestURL(t, "https://some-cluster:99882/api/v1/namespaces/my-namespace/pods/kafka-broker-0/portforward"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, resolveServerURL(tt.args.host, tt.args.namespace, tt.args.podName), "resolveServerURL(%v, %v, %v)", tt.args.host, tt.args.namespace, tt.args.podName)
		})
	}
}

func resolveTestURL(t *testing.T, path string) url.URL {
	t.Helper()

	u, err := url.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	return *u
}
