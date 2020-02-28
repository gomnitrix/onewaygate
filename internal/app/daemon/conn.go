package daemon

import (
	"net"
	"sync"
	"time"
)

type ConnectionManager struct {
	sync.WaitGroup
	Counter   int
	mux       sync.Mutex
	idleConns map[string]net.Conn
}

func newConnectionManager() *ConnectionManager {
	cm := &ConnectionManager{}
	cm.WaitGroup = sync.WaitGroup{}
	cm.idleConns = make(map[string]net.Conn)
	return cm
}

func (cm *ConnectionManager) add(delta int) {
	cm.Counter += delta
	cm.WaitGroup.Add(delta)
}

func (cm *ConnectionManager) done() {
	cm.Counter--
	cm.WaitGroup.Done()
}

func (cm *ConnectionManager) close(t time.Duration) {
	cm.mux.Lock()
	dt := time.Now().Add(t)
	for _, c := range cm.idleConns {
		c.SetDeadline(dt)
	}
	cm.idleConns = nil
	cm.mux.Unlock()
	cm.WaitGroup.Wait()
	return
}

func (cm *ConnectionManager) rmIdleConns(key string) {
	cm.mux.Lock()
	delete(cm.idleConns, key)
	cm.mux.Unlock()
}

func (cm *ConnectionManager) addIdleConns(key string, conn net.Conn) {
	cm.mux.Lock()
	cm.idleConns[key] = conn
	cm.mux.Unlock()
}
