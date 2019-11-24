package task

import (
	"sync"
)

// CopyInfo detail for copy file
type CopyInfo struct {
	Source      []string
	Destination string
}

// ExecLog command/scp execute log by devices
type ExecLog struct {
	Device string
	Log    string
}

// Manager manager for task
type Manager struct {
	newClient  func(*ConnInfo) (CloseableClient, error)
	connInfos  []*ConnInfo
	infoLocker *sync.Mutex
	peakCount  int
	copyInfos  []*CopyInfo
	cmds       []string
	logChan    chan *ExecLog
}

func (manager *Manager) peakConnInfo() *ConnInfo {
	manager.infoLocker.Lock()
	defer manager.infoLocker.Unlock()
	if manager.peakCount >= len(manager.connInfos) {
		return nil
	}
	target := manager.connInfos[manager.peakCount]
	manager.peakCount++
	return target
}

func (manager *Manager) runTask(done func()) {
	var result string
	info := manager.peakConnInfo()

	defer func() {
		if info != nil {
			manager.logChan <- &ExecLog{
				Device: info.Addr,
				Log:    result,
			}
			go manager.runTask(done)
		} else {
			done()
		}
	}()

	if info == nil {
		return
	}
	c, err := manager.newClient(info)
	if err != nil {
		result += err.Error() + "\n"
		return
	}
	defer c.Close()
	for _, copyInfo := range manager.copyInfos {
		err = c.Copy(copyInfo.Source, copyInfo.Destination)
		if err != nil {
			result += err.Error() + "\n"
			return
		}
	}

	for _, cmd := range manager.cmds {
		output, err := c.Exec(cmd)
		if err != nil {
			result += err.Error() + "\n"
			return
		}
		result += output
	}
}

// LogChan get log channel
func (manager *Manager) LogChan() <-chan *ExecLog {
	return manager.logChan
}

// Start all tasks parallelly
func (manager *Manager) Start(parallelCount int) {
	group := sync.WaitGroup{}
	group.Add(parallelCount)
	for i := 0; i < parallelCount; i++ {
		go manager.runTask(group.Done)
	}
	group.Wait()
	close(manager.logChan)
}

// NewManager create manager for specific copy and cmd
func NewManager(connInfos []*ConnInfo, copyInfos []*CopyInfo, cmds []string) *Manager {
	return &Manager{
		newClient:  NewClient,
		infoLocker: &sync.Mutex{},
		peakCount:  0,
		connInfos:  connInfos,
		copyInfos:  copyInfos,
		cmds:       cmds,
		logChan:    make(chan *ExecLog),
	}
}
