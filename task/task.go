package task

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"os"
	"path"
	"time"
)

// CloseableClient gosible client interface
type CloseableClient interface {
	Exec(cmd string) (string, error)
	Copy(filePath []string, destDir string) error
	Close() error
}

type sshSession interface {
	Close() error
	Run(cmd string) error
	StdinPipe() (io.WriteCloser, error)
	CombinedOutput(cmd string) ([]byte, error)
}

type sshClient interface {
	NewSession() (sshSession, error)
	Close() error
}

// ConnInfo info of ssh connection
type ConnInfo struct {
	Addr       string
	User       string
	Password   string
	PrivateKey []byte
	Proxy      *ConnInfo
}

type warpSSHClient struct {
	*ssh.Client
}

// Task copyable ssh client
type Task struct {
	sshClient sshClient
}

func (wc *warpSSHClient) NewSession() (sshSession, error) {
	return wc.Client.NewSession()
}

func newClientConfig(info *ConnInfo) (*ssh.ClientConfig, error) {
	authMethods := []ssh.AuthMethod{}

	if info.Password != "" {
		authMethods = append(authMethods, ssh.Password(info.Password))
	}

	if info.PrivateKey != nil && len(info.PrivateKey) != 0 {
		key, err := ssh.ParsePrivateKey(info.PrivateKey)
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(key))
		}
	}

	if len(authMethods) == 0 {
		return nil, errors.New("no auth info provided")
	}

	return &ssh.ClientConfig{
		User:            info.User,
		Auth:            authMethods,
		HostKeyCallback: func(string, net.Addr, ssh.PublicKey) error { return nil },
		Timeout:         time.Second * 30,
	}, nil
}

// NewClient create new gosible client
func NewClient(info *ConnInfo) (CloseableClient, error) {
	var client *ssh.Client
	var config *ssh.ClientConfig

	config, err := newClientConfig(info)
	if err != nil {
		return nil, err
	}

	if info.Proxy != nil {
		proxyConfig, err := newClientConfig(info.Proxy)
		if err != nil {
			return nil, err
		}
		proxy, err := ssh.Dial("tcp", info.Proxy.Addr, proxyConfig)
		if err != nil {
			return nil, err
		}
		conn, err := proxy.Dial("tcp", info.Addr)
		if err != nil {
			return nil, err
		}
		ncc, chans, reqs, err := ssh.NewClientConn(conn, info.Addr, config)
		if err != nil {
			return nil, err
		}
		client = ssh.NewClient(ncc, chans, reqs)
	} else {
		client, err = ssh.Dial("tcp", info.Addr, config)
		if err != nil {
			return nil, err
		}
	}

	return &Task{
		sshClient: &warpSSHClient{
			Client: client,
		},
	}, nil
}

// Close close gosible client
func (t *Task) Close() error {
	return t.sshClient.Close()
}

// Exec execute command on client
func (t *Task) Exec(cmd string) (string, error) {
	var result string
	session, err := t.sshClient.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	data, err := session.CombinedOutput(cmd)

	if data != nil {
		result += string(data) + "\n"
	}
	if err != nil {
		result += err.Error() + "\n"
	}

	return result, nil
}

// Copy copy file to client
func (t *Task) Copy(filePaths []string, destDir string) error {
	session, err := t.sshClient.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	var readers []io.Reader

	for _, filePath := range filePaths {
		r, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer r.Close()
		readers = append(readers, r)
	}

	go func() {
		w, err := session.StdinPipe()

		if err != nil {
			log.Printf("stdin err %+v", err)
		}
		defer w.Close()
		for i, r := range readers {
			fs, _ := os.Stat(filePaths[i])

			fmt.Fprintln(w, fmt.Sprintf("C0%o", fs.Mode().Perm()), fs.Size(), path.Base(filePaths[i]))
			io.Copy(w, r)
			fmt.Fprint(w, "\x00")
		}
	}()

	if _, err := session.CombinedOutput("/usr/bin/scp -qtr " + destDir); err != nil {
		return err
	}
	return nil
}
