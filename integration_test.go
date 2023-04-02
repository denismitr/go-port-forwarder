package portforwarder

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPortForward_Integration(t *testing.T) {
	runIntegrationTests := os.Getenv("RUN_INTEGRATION_TESTS")
	if runIntegrationTests != "true" {
		t.Skipf("skipping integration tests")
	}

	masterURL := os.Getenv("K8S_MASTER_URL")
	kubeConfig := os.Getenv("K8S_CONFIG")
	conn := NewKubeConfigConnector(masterURL, kubeConfig)

	t.Run("integration test with nginx", func(t *testing.T) {
		pf, err := NewPortForwarder(conn)
		require.NoError(t, err)
		require.NotNil(t, pf)

		process, err := pf.PortForwardAPod(
			context.TODO(),
			80,
			"forwarder",
			map[string]string{"run": "nginx"},
		)
		require.NoError(t, err)
		require.NotNil(t, process)

		go func() {
			<-time.After(3 * time.Second)
			process.Stop()
		}()

		<-process.Ready()
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d", process.Port))
		require.NoError(t, err)
		defer resp.Body.Close()

		nginxHtml, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Truef(
			t,
			strings.Contains(string(nginxHtml),
				"Welcome to nginx!"),
			"No expected text in NGINX response: %s",
			string(nginxHtml),
		)

		<-process.Done()
		require.NoError(t, process.Err())
		t.Log("\nport forward for nginx 80 is done")
	})
}
