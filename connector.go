package portforwarder

import (
	"fmt"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type KubeConnector struct {
	masterURL, kubeConfig string
	restConfig            *rest.Config
}

func NewKubeConfigConnector(masterURL, kubeConfig string) *KubeConnector {
	return &KubeConnector{
		masterURL:  masterURL,
		kubeConfig: kubeConfig,
	}
}

func (c *KubeConnector) Connect() (*rest.Config, *kubernetes.Clientset, error) {
	restCfg, err := parseKubeConfig(
		c.masterURL, c.kubeConfig,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get rest config for kubernetes: %w", err)
	}

	k8sClientSet, err := createK8SClientSet(restCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes client set: %w", err)
	}

	return restCfg, k8sClientSet, nil
}
