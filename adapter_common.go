package main

import "github.com/skybon/semaphore"

type AdapterID string

type AdaptFunc func([]string, GameInfo, SettingsMap) ([]ServerData, error)

type AdapterCollection struct {
	data      map[AdapterID]AdaptFunc
	semaphore semaphore.Semaphore
}

func (c *AdapterCollection) Insert(k AdapterID, v AdaptFunc) {
	c.semaphore.Exec(func() {
		c.data[k] = v
	})
}

func (c *AdapterCollection) Retrieve(k AdapterID) (v AdaptFunc, exists bool) {
	c.semaphore.Exec(func() {
		v, exists = c.data[k]
	})

	return v, exists
}

func MakeAdapterCollection() *AdapterCollection {
	return &AdapterCollection{data: make(map[AdapterID]AdaptFunc), semaphore: semaphore.MakeSemaphore(1)}
}
