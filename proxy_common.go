package main

import "github.com/skybon/semaphore"

type ProxyID string

type ProxyFunc func(GameInfo, SettingsMap) ([]string, error)

type ProxyCollection struct {
	data      map[ProxyID]ProxyFunc
	semaphore semaphore.Semaphore
}

func (c *ProxyCollection) All() {}

func (c *ProxyCollection) Insert(k ProxyID, v ProxyFunc) {
	c.semaphore.Exec(func() {
		c.data[k] = v
	})
}

func (c *ProxyCollection) Retrieve(k ProxyID) (v ProxyFunc, exists bool) {
	c.semaphore.Exec(func() {
		v, exists = c.data[k]
	})

	return v, exists
}

func MakeProxyCollection() *ProxyCollection {
	return &ProxyCollection{data: make(map[ProxyID]ProxyFunc), semaphore: semaphore.MakeSemaphore(1)}
}
