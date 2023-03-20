package portforwarder

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func createK8SClientSet(k8sCfg *rest.Config) (*kubernetes.Clientset, error) {
	k8sClientSet, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create kubernetes client set")
	}

	return k8sClientSet, nil
}

func parseKubeConfig(masterURL, config string) (*rest.Config, error) {
	k8sCfg, err := clientcmd.BuildConfigFromKubeconfigGetter(
		masterURL,
		func() (*clientcmdapi.Config, error) {
			b, err := base64.StdEncoding.DecodeString(config)
			if err != nil {
				return nil, errors.Wrapf(
					err,
					"failed to base64 decode config for masterURL %s",
					masterURL,
				)
			}

			return clientcmd.Load(b)
		},
	)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to create kubernetes config")
	}

	return k8sCfg, nil
}
