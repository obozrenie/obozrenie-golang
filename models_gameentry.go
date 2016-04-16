package main

import (
	"net/url"
	"time"

	"github.com/skybon/semaphore"
)

const (
	ProxyInvalid   = ProxyID("")
	AdapterInvalid = AdapterID("")
)

/*
--------------------------
SETTINGS
--------------------------
*/

type SettingsMap map[string]string

// GameSettings contains user-definable settings.
type GameSettings interface {
	AllKeys() []string
	AllSettings() SettingsMap
	Get(string) (string, bool)
	Set(string, string)
	Remove(string)
	Clear()
}

type SimpleGameSettings struct {
	semaphore semaphore.Semaphore
	data      SettingsMap
}

func (s *SimpleGameSettings) safeExec(f func()) { s.semaphore.Exec(f) }

func (s *SimpleGameSettings) AllKeys() (output []string) {
	s.safeExec(func() {
		output = make([]string, 0, len(s.data))
		for k := range s.data {
			output = append(output, k)
		}
	})
	return output
}

func (s *SimpleGameSettings) AllSettings() (outMap SettingsMap) {
	s.safeExec(func() {
		outMap = make(SettingsMap, len(s.data))
		for k, v := range s.data {
			outMap[k] = v
		}
	})

	return outMap
}

func (s *SimpleGameSettings) Get(name string) (v string, exists bool) {
	s.safeExec(func() { v, exists = s.data[name] })
	return v, exists
}

func (s *SimpleGameSettings) Set(k string, v string) {
	s.safeExec(func() {
		s.data[k] = v
	})
	return
}

func (s *SimpleGameSettings) Remove(k string) {
	s.safeExec(func() {
		delete(s.data, k)
	})
}

func (s *SimpleGameSettings) Clear() {
	s.safeExec(func() { s.data = make(map[string]string) })
}

func MakeSimpleGameSettings() *SimpleGameSettings {
	return &SimpleGameSettings{semaphore: semaphore.MakeSemaphore(1), data: make(map[string]string)}
}

/*
--------------------------
SERVERS
--------------------------
*/

// ServerList contains user-definable master server list for query.
type ServerList []url.URL

// StatFunc is the procedure that is used for querying servers.
type StatFunc func(SettingsMap, ServerList) ([]ServerData, error)

type ServerSettings map[string]string

type PlayerData struct {
	Name string
	Info map[string]string
}

func NewPlayerData() PlayerData {
	return PlayerData{Info: map[string]string{}}
}

// ServerData is a basic structure containing single server entry.
type ServerData struct {
	Host       string
	Name       string
	Status     string
	Map        string
	Ping       int
	Secure     bool
	NumPlayers int
	MaxPlayers int
	Players    []PlayerData
	Settings   ServerSettings
}

func MakeServerData(data ServerData) ServerData {
	if data.Settings == nil {
		data.Settings = ServerSettings{}
	}
	return data
}

func NewServerData() ServerData { return ServerData{Settings: ServerSettings{}} }

type ServerCollection interface {
	ModDate() time.Time

	Find(func(int, ServerData) bool) []ServerData
	Insert([]ServerData) error
	Delete(func(int, ServerData) bool) []ServerData
}

// ServerCollection represents server collection.
type SimpleServerCollection struct {
	semaphore semaphore.Semaphore
	data      []ServerData
	modDt     time.Time
}

func (c *SimpleServerCollection) safeExec(f func()) { c.semaphore.Exec(f) }

func (c *SimpleServerCollection) bumpModDate() { c.modDt = time.Now() }

func (c *SimpleServerCollection) find(f func(int, ServerData) bool) (output []ServerData) {
	for i, v := range c.data {
		if f(i, v) {
			output = append(output, v)
		}
	}

	return output
}

func (c *SimpleServerCollection) insert(data []ServerData) (err error) {
	for _, v := range data {
		newEntry := MakeServerData(v)
		c.data = append(c.data, newEntry)
	}

	c.bumpModDate()

	return err
}

func (c *SimpleServerCollection) delete(f func(int, ServerData) bool) (output []ServerData) {
	newData := make([]ServerData, 0, len(c.data))

	for i, v := range c.data {
		if f(i, v) {
			output = append(output, v)
		} else {
			newData = append(newData, v)
		}
	}

	c.data = newData

	return output
}

func (c *SimpleServerCollection) ModDate() time.Time { return c.modDt }

func (c *SimpleServerCollection) Find(f func(int, ServerData) bool) (output []ServerData) {
	c.safeExec(func() { output = c.find(f) })

	return output
}

func (c *SimpleServerCollection) Insert(data []ServerData) (err error) {
	c.safeExec(func() { err = c.insert(data) })

	return err
}

func (c *SimpleServerCollection) Delete(f func(int, ServerData) bool) (output []ServerData) {
	c.safeExec(func() { output = c.delete(f) })

	return output
}

// MakeServerCollection creates an empty SimpleServerCollection.
func MakeServerCollection() *SimpleServerCollection {
	return &SimpleServerCollection{data: []ServerData{}, semaphore: semaphore.MakeSemaphore(1)}
}

/*
--------------------------
GAMES
--------------------------
*/

type QueryStatus int

const (
	QueryEmpty = QueryStatus(iota)
	QueryReady
	QueryWorking
	QueryError
)

// GameInfo is a structure that contains basic information desribing the game's internals. It is a programmer's responsibility to fill it in. User-definable settings should be placed in GameSettings instead.
type GameInfo struct {
	Name     string
	Proxy    ProxyID
	Adapter  AdapterID
	StatFunc StatFunc
}

// GameEntry is a structure containing all information about a game.
type GameEntry struct {
	Info     GameInfo
	Settings GameSettings
	Servers  ServerCollection
	Status   QueryStatus
}

// MakeGameEntry creates an empty game entry.
func MakeGameEntry() *GameEntry {
	return &GameEntry{Settings: MakeSimpleGameSettings(), Servers: MakeServerCollection(), Status: QueryEmpty}
}
