package task

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	sshServer "github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/suite"
	"testing"
)

type taskSuite struct {
	suite.Suite
	mockSSHServer  *sshServer.Server
	mockSSHClient  *mockSSHClient
	mockSSHSession *mockSSHSession
	addr           string
}

func (s *taskSuite) SetupTest() {
	s.addr = ":5555"
	s.mockSSHServer = &sshServer.Server{
		Addr: s.addr,
	}
	s.mockSSHClient = &mockSSHClient{}
	s.mockSSHSession = &mockSSHSession{}
}

// Test_newClientConfig_1 should append all authMethods for ssh.ClientConfig
func (s *taskSuite) Test_newClientConfig_1() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key),
		},
	)

	info := &ConnInfo{
		Addr:       "addr",
		User:       "user",
		Password:   "pwd",
		PrivateKey: pemdata,
	}
	config, err := newClientConfig(info)
	s.Equal(nil, err)
	s.Equal(config.User, info.User)
	s.Len(config.Auth, 2)
}

// Test_newClientConfig_2 should throw error when config has no authMethod
func (s *taskSuite) Test_newClientConfig_2() {
	info := &ConnInfo{
		Addr: "addr",
		User: "user",
	}
	config, err := newClientConfig(info)

	s.Nil(config)
	s.Equal(errors.New("no auth info provided"), err)
}

// Test_NewClient_1 should create new client with input connInfo
func (s *taskSuite) Test_NewClient_1() {
	pwdChan := make(chan string)
	s.mockSSHServer.PasswordHandler = func(ctx sshServer.Context, pwd string) bool {
		go func() {
			pwdChan <- pwd
		}()
		return true
	}
	go s.mockSSHServer.ListenAndServe()
	defer s.mockSSHServer.Close()
	info := &ConnInfo{
		Addr:     "127.0.0.1:5555",
		Password: "pwd",
		User:     "user",
	}

	client, err := NewClient(info)
	s.Nil(err)
	s.NotNil(client)
	s.Equal(info.Password, <-pwdChan)
}

// Test_NewClient_1 should create client through proxy
func (s *taskSuite) Test_NewClient_2() {
	pwdChan := make(chan string, 2)
	s.mockSSHServer.PasswordHandler = func(ctx sshServer.Context, pwd string) bool {
		go func() {
			pwdChan <- pwd
		}()
		return true
	}
	s.mockSSHServer.ChannelHandlers = map[string]sshServer.ChannelHandler{
		"direct-tcpip": sshServer.DirectTCPIPHandler,
		"session":      sshServer.DefaultSessionHandler,
	}
	s.mockSSHServer.LocalPortForwardingCallback = func(
		ctx sshServer.Context,
		destinationHost string,
		destinationPort uint32,
	) bool {
		return true
	}
	go s.mockSSHServer.ListenAndServe()
	defer s.mockSSHServer.Close()
	info := &ConnInfo{
		Proxy: &ConnInfo{
			Addr:     "127.0.0.1:5555",
			Password: "9453",
			User:     "user",
		},
		Addr:     "127.0.0.1:5555",
		Password: "pwd",
		User:     "user",
	}

	client, err := NewClient(info)
	s.Nil(err)
	s.NotNil(client)
	s.Equal(info.Proxy.Password, <-pwdChan)
	s.Equal(info.Password, <-pwdChan)
}

// Test_Close should call sshClient.close
func (s *taskSuite) Test_Close() {
	task := &Task{
		sshClient: s.mockSSHClient,
	}
	s.mockSSHClient.On("Close").Return(nil)
	err := task.Close()
	s.Nil(err)
	s.mockSSHClient.AssertCalled(s.T(), "Close")
}

// Test_Exec_1 should execute command on ssh session
func (s *taskSuite) Test_Exec_1() {
	cmd := "cmd"
	s.mockSSHClient.On("NewSession").Return(s.mockSSHSession, nil)
	s.mockSSHSession.On("Close").Return(nil)
	s.mockSSHSession.On("CombinedOutput", cmd).Return([]byte(cmd), nil)
	task := &Task{
		sshClient: s.mockSSHClient,
	}
	result, err := task.Exec(cmd)
	s.Nil(err)
	s.Equal(fmt.Sprintf("%s\n", cmd), result)
	s.mockSSHClient.AssertCalled(s.T(), "NewSession")
	s.mockSSHSession.AssertCalled(s.T(), "Close")
	s.mockSSHSession.AssertCalled(s.T(), "CombinedOutput", cmd)
}

// Test_Exec_2 should append error log on result
func (s *taskSuite) Test_Exec_2() {
	cmd := "cmd"
	errString := "errString"
	s.mockSSHClient.On("NewSession").Return(s.mockSSHSession, nil)
	s.mockSSHSession.On("Close").Return(nil)
	s.mockSSHSession.On("CombinedOutput", cmd).Return([]byte(cmd), errors.New(errString))
	task := &Task{
		sshClient: s.mockSSHClient,
	}
	result, err := task.Exec(cmd)
	s.Nil(err)
	s.Equal(fmt.Sprintf("%s\n%s\n", cmd, errString), result)
	s.mockSSHClient.AssertCalled(s.T(), "NewSession")
	s.mockSSHSession.AssertCalled(s.T(), "Close")
	s.mockSSHSession.AssertCalled(s.T(), "CombinedOutput", cmd)
}

// Test_Exec_3 should escape when create session failed
func (s *taskSuite) Test_Exec_3() {
	cmd := "cmd"
	errString := "errString"
	s.mockSSHClient.On("NewSession").Return(s.mockSSHSession, errors.New(errString))
	s.mockSSHSession.On("Close").Return(nil)
	task := &Task{
		sshClient: s.mockSSHClient,
	}
	result, err := task.Exec(cmd)
	s.Equal(errors.New(errString), err)
	s.mockSSHSession.AssertNotCalled(s.T(), "NewSession")
	s.Equal("", result)
}

func TestTask(t *testing.T) {
	suite.Run(t, new(taskSuite))
}
