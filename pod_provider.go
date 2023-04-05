package portforwarder

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type provider struct {
	clientSet *kubernetes.Clientset
}

func newSelectorFromKubeConfig(k8sClientSet *kubernetes.Clientset) *provider {
	return &provider{clientSet: k8sClientSet}
}

type listPodsCommand struct {
	namespace      string
	labelSelectors map[string]string
	fieldSelectors map[string]string
}

func (p *provider) getPod(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	pod, err := p.clientSet.
		CoreV1().
		Pods(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf(
			"%w: failed to get pod %s in namespace %s",
			err, name, namespace,
		)
	}

	return pod, nil
}

func (p *provider) listPods(
	ctx context.Context, cmd *listPodsCommand,
) (*corev1.PodList, error) {
	opts := metav1.ListOptions{
		LabelSelector: buildSelector(cmd.labelSelectors),
		FieldSelector: buildSelector(cmd.fieldSelectors),
	}
	resp, err := p.clientSet.
		CoreV1().
		Pods(cmd.namespace).
		List(ctx, opts)
	if err != nil {
		return nil, errors.Wrapf(
			err, "failed to list pods with opts %+v",
			opts,
		)
	}

	return resp, nil
}

func buildSelector(labels map[string]string) string {
	selectors := make([]string, len(labels))

	i := 0
	for k, v := range labels {
		selectors[i] = fmt.Sprintf("%s=%s", k, v)
		i++
	}

	return strings.Join(selectors, ",")
}
