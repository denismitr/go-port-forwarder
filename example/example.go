package main

import (
	"context"
	"fmt"
	"github.com/denismitr/portforwarder"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	masterURL := os.Getenv("K8S_MASTER_URL")
	kubeConfig := os.Getenv("K8S_CONFIG")
	conn := portforwarder.NewKubeConfigConnector(masterURL, kubeConfig)
	pf, err := portforwarder.NewPortForwarder(conn)
	if err != nil {
		panic(err)
	}

	namespace := "forwarder"
	targetPort := uint(80)
	matchLabel := map[string]string{"run": "nginx"}

	process, err := pf.PortForwardAPod(
		context.TODO(),
		&portforwarder.TargetPod{
			Namespace:     namespace,
			Port:          targetPort,
			LabelSelector: matchLabel,
		},
	)
	if err != nil {
		panic(err)
	}

	go func() {
		<-time.After(3 * time.Second)
		log.Printf("stopping the process")
		process.Stop()
	}()

	<-process.Started()
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", process.Port))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	<-process.Finished()
}
