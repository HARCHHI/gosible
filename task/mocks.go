package task

import (
	// "crypto/rand"
	// "crypto/rsa"
	// "crypto/x509"
	// "encoding/pem"
	// "errors"
	// sshServer "github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/mock"
	// "github.com/stretchr/testify/suite"
	// "testing"
	"io"
)

type mockSSHClient struct {
	mock.Mock
}

type mockSSHSession struct {
	mock.Mock
}

func (mc *mockSSHClient) NewSession() (sshSession, error) {
	args := mc.Called()
	return args.Get(0).(sshSession), args.Error(1)
}

func (mc *mockSSHClient) Close() error {
	args := mc.Called()
	return args.Error(0)
}

func (ms *mockSSHSession) Close() error {
	args := ms.Called()
	return args.Error(0)
}

func (ms *mockSSHSession) Run(cmd string) error {
	args := ms.Called(cmd)

	return args.Error(0)
}

func (ms *mockSSHSession) StdinPipe() (io.WriteCloser, error) {
	args := ms.Called()

	return args.Get(0).(io.WriteCloser), args.Error(1)
}

func (ms *mockSSHSession) CombinedOutput(cmd string) ([]byte, error) {
	args := ms.Called(cmd)

	return args.Get(0).([]byte), args.Error(1)
}
