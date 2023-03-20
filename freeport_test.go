package portforwarder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_netFreePortProvider(t *testing.T) {
	t.Run("localhost on 0 port", func(t *testing.T) {
		fp := newNetFreePortProvider("tcp", "localhost", 0)
		port, err := fp.getFreePort()
		assert.NoError(t, err)
		assert.Truef(t, port > 0, "port should be greater than 0, got %d", port)
	})
}
