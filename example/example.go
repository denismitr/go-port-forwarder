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

	process, err := pf.PortForwardAPod(
		context.TODO(),
		80,
		"forwarder",
		map[string]string{"run": "nginx"},
	)
	if err != nil {
		panic(err)
	}

	go func() {
		<-time.After(3 * time.Second)
		log.Printf("stopping the process")
		process.Stop()
	}()

	<-process.Ready()
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", process.Port))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	<-process.Done()
}
