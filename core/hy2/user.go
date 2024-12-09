package hy2

import (
	"net"
	"sync"

	"github.com/apernet/hysteria/core/v2/server"
	"github.com/perfect-panel/ppanel-node/api/panel"
	"github.com/perfect-panel/ppanel-node/common/counter"
	vCore "github.com/perfect-panel/ppanel-node/core"
)

var _ server.Authenticator = &PP{}

type PP struct {
	usersMap map[string]int
	mutex    sync.Mutex
}

func (v *PP) Authenticate(addr net.Addr, auth string, tx uint64) (ok bool, id string) {
	v.mutex.Lock()
	defer v.mutex.Unlock()
	if _, exists := v.usersMap[auth]; exists {
		return true, auth
	}
	return false, ""
}

func (h *Hysteria2) AddUsers(p *vCore.AddUsersParams) (added int, err error) {
	var wg sync.WaitGroup
	for _, user := range p.Users {
		wg.Add(1)
		go func(u panel.UserInfo) {
			defer wg.Done()
			h.Auth.mutex.Lock()
			h.Auth.usersMap[u.Uuid] = u.Id
			h.Auth.mutex.Unlock()
		}(user)
	}
	wg.Wait()
	return len(p.Users), nil
}

func (h *Hysteria2) DelUsers(users []panel.UserInfo, tag string) error {
	var wg sync.WaitGroup
	for _, user := range users {
		wg.Add(1)
		go func(u panel.UserInfo) {
			defer wg.Done()
			h.Auth.mutex.Lock()
			delete(h.Auth.usersMap, u.Uuid)
			h.Auth.mutex.Unlock()
		}(user)
	}
	wg.Wait()
	return nil
}

func (h *Hysteria2) GetUserTraffic(tag string, uuid string, reset bool) (up int64, down int64) {
	if v, ok := h.Hy2nodes[tag].TrafficLogger.(*HookServer).Counter.Load(tag); ok {
		c := v.(*counter.TrafficCounter)
		up = c.GetUpCount(uuid)
		down = c.GetDownCount(uuid)
		if reset {
			c.Reset(uuid)
		}
		return up, down
	}
	return 0, 0
}
