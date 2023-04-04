# GO Embedded Port Forwarder for Kubernetes

## Usage
#### Example
```go
process, err := pf.PortForwardAPod(
    context.TODO(),
    &portforwarder.TargetPod{
        Namespace:     "my-namespace",
        Port:          8888,
        LabelSelector: map[string]string{"run": "nginx"},
    },
)
if err != nil {
    panic(err)
}

go func() {
    // force stop forwarding 
	// but when timeout is required - use context instead
    <-time.After(5 * time.Second)
    process.Stop() 
}()

<-process.Started() // signals that port forwarding has started
resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", process.Port))
if err != nil {
    panic(err)
}

<-process.Finished() // signals that port forward has finished
```