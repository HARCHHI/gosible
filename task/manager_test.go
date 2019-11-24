package task

import (
	"errors"
	// "fmt"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
)

type managerSuite struct {
	suite.Suite
	mockClient *mockClient
	mockGroup  *mockGroup
}

func (ms *managerSuite) SetupTest() {
	ms.mockClient = &mockClient{}
	ms.mockGroup = &mockGroup{}
}

// Test_peakConnInfo should make sure read connInfos is goroutine saved
func (ms *managerSuite) Test_peakConnInfo() {
	infoChan := make(chan *ConnInfo, 3)
	var nilCount int
	var infoCount int
	infos := []*ConnInfo{
		&ConnInfo{User: "user"},
		&ConnInfo{User: "user"},
	}
	manager := &Manager{
		infoLocker: &sync.Mutex{},
		peakCount:  0,
		connInfos:  infos,
	}
	group := sync.WaitGroup{}
	group.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			infoChan <- manager.peakConnInfo()
			group.Done()
		}()
	}
	for i := 0; i < 3; i++ {
		info := <-infoChan
		if info != nil {
			infoCount++
		} else {
			nilCount++
		}
	}
	group.Wait()
	ms.Equal(2, manager.peakCount)
	ms.Equal(2, infoCount)
	ms.Equal(1, nilCount)
}

// Test_runTask_1 should execute cmds in manager and send result to logChan
func (ms *managerSuite) Test_runTask_1() {
	connInfo := &ConnInfo{
		User: "user",
		Addr: "addr",
	}
	copyInfo := &CopyInfo{
		Source:      []string{"source"},
		Destination: "dest",
	}

	manager := &Manager{
		newClient:  func(*ConnInfo) (CloseableClient, error) { return ms.mockClient, nil },
		infoLocker: &sync.Mutex{},
		peakCount:  0,
		connInfos:  []*ConnInfo{connInfo},
		copyInfos:  []*CopyInfo{copyInfo},
		cmds:       []string{"cmd"},
		logChan:    make(chan *ExecLog),
	}

	ms.mockGroup.On("Done").Return()
	ms.mockClient.On("Exec", manager.cmds[0]).Return("result", nil)
	ms.mockClient.On("Copy", copyInfo.Source, copyInfo.Destination).Return(nil)
	ms.mockClient.On("Close").Return(nil)

	go manager.runTask(ms.mockGroup.Done)

	result := <-manager.LogChan()

	ms.Equal(&ExecLog{
		Device: connInfo.Addr,
		Log:    "result",
	}, result)
	ms.mockGroup.AssertCalled(ms.T(), "Done")
	ms.mockClient.AssertCalled(ms.T(), "Exec", manager.cmds[0])
	ms.mockClient.AssertCalled(ms.T(), "Copy", copyInfo.Source, copyInfo.Destination)
	ms.mockClient.AssertCalled(ms.T(), "Close")
}

// Test_runTask_2 should connect and copy file twice
func (ms *managerSuite) Test_runTask_2() {
	connInfo := &ConnInfo{
		User: "user",
		Addr: "addr",
	}
	copyInfo := &CopyInfo{
		Source:      []string{"source"},
		Destination: "dest",
	}

	manager := &Manager{
		newClient:  func(*ConnInfo) (CloseableClient, error) { return ms.mockClient, nil },
		infoLocker: &sync.Mutex{},
		peakCount:  0,
		connInfos:  []*ConnInfo{connInfo, connInfo},
		copyInfos:  []*CopyInfo{copyInfo},
		cmds:       []string{"cmd"},
		logChan:    make(chan *ExecLog),
	}

	ms.mockGroup.On("Done").Return()
	ms.mockClient.On("Copy", copyInfo.Source, copyInfo.Destination).Return(errors.New("err")).Twice()
	ms.mockClient.On("Exec", manager.cmds[0]).Return("result", nil)
	ms.mockClient.On("Close").Return(nil)

	go manager.runTask(ms.mockGroup.Done)

	result1 := <-manager.LogChan()
	result2 := <-manager.LogChan()

	ms.Equal(&ExecLog{
		Device: connInfo.Addr,
		Log:    "err\n",
	}, result1)
	ms.Equal(result1, result2)
	ms.mockGroup.AssertCalled(ms.T(), "Done")
	ms.mockClient.AssertCalled(ms.T(), "Copy", copyInfo.Source, copyInfo.Destination)
	ms.mockClient.AssertNumberOfCalls(ms.T(), "Copy", 2)
	ms.mockClient.AssertNotCalled(ms.T(), "Exec")
	ms.mockClient.AssertCalled(ms.T(), "Close")
}

// Test_runTask_3 should run task parallelly
func (ms *managerSuite) Test_runTask_3() {
	connInfo := &ConnInfo{
		User: "user",
		Addr: "addr",
	}
	copyInfo := &CopyInfo{
		Source:      []string{"source"},
		Destination: "dest",
	}

	manager := &Manager{
		newClient:  func(*ConnInfo) (CloseableClient, error) { return ms.mockClient, nil },
		infoLocker: &sync.Mutex{},
		peakCount:  0,
		connInfos:  []*ConnInfo{connInfo, connInfo},
		copyInfos:  []*CopyInfo{copyInfo},
		cmds:       []string{"cmd"},
		logChan:    make(chan *ExecLog),
	}

	ms.mockGroup.On("Done").Return()
	ms.mockClient.On("Copy", copyInfo.Source, copyInfo.Destination).Return(nil)
	ms.mockClient.On("Exec", manager.cmds[0]).Return("", errors.New("err")).Twice()
	ms.mockClient.On("Close").Return(nil)

	go manager.runTask(ms.mockGroup.Done)
	go manager.runTask(ms.mockGroup.Done)

	result1 := <-manager.LogChan()
	result2 := <-manager.LogChan()

	ms.Equal(&ExecLog{
		Device: connInfo.Addr,
		Log:    "err\n",
	}, result1)
	ms.Equal(result1, result2)
	ms.mockGroup.AssertCalled(ms.T(), "Done")
	ms.mockClient.AssertCalled(ms.T(), "Copy", copyInfo.Source, copyInfo.Destination)
	ms.mockClient.AssertNumberOfCalls(ms.T(), "Exec", 2)
	ms.mockClient.AssertCalled(ms.T(), "Close")
}

// Test_Start_1 should run task parallelly
func (ms *managerSuite) Test_Start_1() {
	connInfo := &ConnInfo{
		User: "user",
		Addr: "addr",
	}
	copyInfo := &CopyInfo{
		Source:      []string{"source"},
		Destination: "dest",
	}
	manager := &Manager{
		newClient:  func(*ConnInfo) (CloseableClient, error) { return ms.mockClient, nil },
		infoLocker: &sync.Mutex{},
		peakCount:  0,
		connInfos:  []*ConnInfo{connInfo, connInfo},
		copyInfos:  []*CopyInfo{copyInfo},
		cmds:       []string{},
		logChan:    make(chan *ExecLog, 3),
	}
	ms.mockClient.On("Copy", copyInfo.Source, copyInfo.Destination).Return(errors.New("err")).Twice()
	ms.mockClient.On("Close").Return(nil)
	manager.Start(2)

	ms.mockClient.AssertNumberOfCalls(ms.T(), "Copy", 2)
	ms.mockClient.AssertNumberOfCalls(ms.T(), "Close", 2)
}

func TestManager(t *testing.T) {
	suite.Run(t, new(managerSuite))
}
