package server

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunServesUntilAskedToStop(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	address := listener.Addr().String()
	_, port, _ := net.SplitHostPort(address)
	listener.Close()

	// Keeps a signal sent before Run listens for it from ending the test binary
	ignored := make(chan os.Signal, 8)
	signal.Notify(ignored, syscall.SIGTERM)
	defer signal.Stop(ignored)

	arguments := os.Args
	os.Args = []string{arguments[0], "-host", "127.0.0.1", "-port", port}
	defer func() { os.Args = arguments }()

	stopped := make(chan struct{})
	go func() {
		Run()
		close(stopped)
	}()

	require.Eventually(t, func() bool {
		response, err := http.Get("http://" + address + "/api/health")
		if err != nil {
			return false
		}
		response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 5*time.Second, 20*time.Millisecond, "the server answers")

	assert.Eventually(t, func() bool {
		syscall.Kill(os.Getpid(), syscall.SIGTERM)
		select {
		case <-stopped:
			return true
		case <-time.After(100 * time.Millisecond):
			return false
		}
	}, 5*time.Second, 10*time.Millisecond, "SIGTERM stops the server without ending the process")

	_, err = http.Get("http://" + address + "/api/health")
	assert.Error(t, err, "the port is released")
}
