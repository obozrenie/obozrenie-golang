package main

import "github.com/skybon/semaphore"

type GameID string

func (i *GameID) String() string { return string(*i) }

// GameTable is an interface to the storage that holds game entries
type GameTable interface {
	CheckGameEntry(GameID) bool
	AllGames() []GameID
	MatchGameEntries(func(GameID, *GameEntry) bool) []GameID
	CreateGameEntry(GameID) error
	RemoveGameEntry(GameID) error

	CopyGameEntry(GameID, bool) (*GameEntry, error)

	QueryStatus(GameID) (QueryStatus, error)
	SetQueryStatus(GameID, QueryStatus) error
	TryLockQuery(GameID) (bool, error)

	GameInfo(GameID) (GameInfo, error)
	SetGameInfo(GameID, GameInfo) error

	Settings(GameID) (SettingsMap, error)
	SetSetting(GameID, string, string) error
	GetSetting(GameID, string) (string, bool, error)
	RemoveSetting(GameID, string) error
	ClearSettings(GameID) error

	FindServers(GameID, func(int, ServerData) bool) ([]ServerData, error)
	AllServers(GameID) ([]ServerData, error)
	InsertServers(GameID, []ServerData) error
	DeleteServers(GameID, func(int, ServerData) bool) ([]ServerData, error)
	ClearServers(GameID) error
}

type MemGameTable struct {
	semaphore semaphore.Semaphore
	data      map[GameID]*GameEntry
}

func (t *MemGameTable) safeExec(f func()) { t.semaphore.Exec(f) }

func (t *MemGameTable) matchGames(f func(GameID, *GameEntry) bool) []GameID {
	output := make([]GameID, 0, len(t.data))
	for k, v := range t.data {
		if f(k, v) {
			output = append(output, k)
		}
	}

	return output
}

func (t *MemGameTable) createGameEntry(id GameID) error {
	_, exists := t.data[id]
	if exists {
		return errGameExists
	}

	t.data[id] = MakeGameEntry()

	return nil
}

func (t *MemGameTable) removeGameEntry(id GameID) error {
	_, exists := t.data[id]
	if !exists {
		return errUnknownGameID
	}

	delete(t.data, id)

	return nil
}

func (t *MemGameTable) gameEntrySnapshot(id GameID, servers bool) (*GameEntry, error) {
	g, gexists := t.data[id]
	if !gexists {
		return nil, errUnknownGameID
	}

	e := MakeGameEntry()
	e.Status = QueryEmpty
	e.Info = g.Info
	for k, v := range g.Settings.AllSettings() {
		e.Settings.Set(k, v)
	}
	if servers {
		e.Servers.Insert(g.Servers.Find(func(int, ServerData) bool { return true }))
	}

	return e, nil
}

func (t *MemGameTable) settings(id GameID) (output SettingsMap, err error) {
	g, gexists := t.data[id]
	if !gexists {
		return nil, errUnknownGameID
	}

	output = g.Settings.AllSettings()

	return output, nil
}

func (t *MemGameTable) getSetting(id GameID, settingID string) (v string, exists bool, err error) {
	g, gexists := t.data[id]
	if !gexists {
		return "", false, errUnknownGameID
	}

	v, exists = g.Settings.Get(settingID)

	return v, exists, err
}

func (t *MemGameTable) setSetting(id GameID, settingID string, v string) error {
	g, exists := t.data[id]
	if !exists {
		return errUnknownGameID
	}

	g.Settings.Set(settingID, v)
	t.data[id] = g

	return nil
}

func (t *MemGameTable) removeSetting(id GameID, settingID string) error {
	g, exists := t.data[id]
	if !exists {
		return errUnknownGameID
	}

	g.Settings.Remove(settingID)
	t.data[id] = g

	return nil
}

func (t *MemGameTable) clearSettings(id GameID) error {
	g, exists := t.data[id]
	if !exists {
		return errUnknownGameID
	}

	g.Settings.Clear()
	t.data[id] = g

	return nil
}

func (t *MemGameTable) findServers(id GameID, f func(int, ServerData) bool) (output []ServerData, err error) {
	g, exists := t.data[id]
	if !exists {
		return nil, errUnknownGameID
	}

	return g.Servers.Find(f), nil
}

func (t *MemGameTable) insertServers(id GameID, data []ServerData) (err error) {
	g, exists := t.data[id]
	if !exists {
		return errUnknownGameID
	}

	err = g.Servers.Insert(data)
	t.data[id] = g

	return err
}

func (t *MemGameTable) deleteServers(id GameID, f func(int, ServerData) bool) (deleted []ServerData, err error) {
	g, exists := t.data[id]
	if !exists {
		return nil, errUnknownGameID
	}
	deleted = g.Servers.Delete(f)
	t.data[id] = g

	return deleted, err
}

func (t *MemGameTable) MatchGameEntries(f func(GameID, *GameEntry) bool) (output []GameID) {
	t.safeExec(func() { output = t.matchGames(f) })

	return output
}

func (t *MemGameTable) AllGames() (output []GameID) {
	t.safeExec(func() { output = t.matchGames(func(GameID, *GameEntry) bool { return true }) })

	return output
}

func (t *MemGameTable) CheckGameEntry(id GameID) (exists bool) {
	t.safeExec(func() { _, exists = t.data[id] })

	return exists
}

func (t *MemGameTable) CreateGameEntry(id GameID) (err error) {
	t.safeExec(func() { err = t.createGameEntry(id) })

	return err
}

func (t *MemGameTable) RemoveGameEntry(id GameID) (err error) {
	t.safeExec(func() { err = t.removeGameEntry(id) })

	return err
}

func (t *MemGameTable) CopyGameEntry(id GameID, servers bool) (e *GameEntry, err error) {
	t.safeExec(func() { e, err = t.gameEntrySnapshot(id, servers) })

	return e, err
}

func (t *MemGameTable) queryStatus(id GameID) (output QueryStatus, err error) {
	g, exists := t.data[id]
	if !exists {
		return output, errUnknownGameID
	}

	return g.Status, nil
}

func (t *MemGameTable) QueryStatus(id GameID) (output QueryStatus, err error) {
	t.safeExec(func() {
		output, err = t.queryStatus(id)
	})

	return output, err
}

func (t *MemGameTable) SetQueryStatus(id GameID, status QueryStatus) (err error) {
	t.safeExec(func() {
		err = func(id GameID, status QueryStatus) error {
			g, exists := t.data[id]
			if !exists {
				return errUnknownGameID
			}

			g.Status = status
			t.data[id] = g

			return nil
		}(id, status)
	})

	return err
}

func (t *MemGameTable) tryLockQuery(id GameID) (success bool, err error) {
	g, exists := t.data[id]
	if !exists {
		return false, errUnknownGameID
	}

	if g.Status == QueryWorking {
		return false, nil
	}

	g.Status = QueryWorking

	return true, nil
}

func (t *MemGameTable) TryLockQuery(id GameID) (success bool, err error) {
	t.safeExec(func() { success, err = t.tryLockQuery(id) })

	return success, err
}

func (t *MemGameTable) GameInfo(id GameID) (output GameInfo, err error) {
	t.safeExec(func() {
		output, err = func(id GameID) (GameInfo, error) {
			g, exists := t.data[id]
			if !exists {
				return GameInfo{}, errUnknownGameID
			}

			return g.Info, nil
		}(id)
	})

	return output, err
}

func (t *MemGameTable) SetGameInfo(id GameID, info GameInfo) (err error) {
	t.safeExec(func() {
		err = func(id GameID, info GameInfo) error {
			g, exists := t.data[id]
			if !exists {
				return errUnknownGameID
			}

			g.Info = info
			t.data[id] = g

			return nil
		}(id, info)
	})

	return err
}

func (t *MemGameTable) Settings(id GameID) (output SettingsMap, err error) {
	t.safeExec(func() { output, err = t.settings(id) })

	return output, err
}

func (t *MemGameTable) GetSetting(id GameID, settingID string) (e string, exists bool, err error) {
	t.safeExec(func() { e, exists, err = t.getSetting(id, settingID) })

	return e, exists, err
}

func (t *MemGameTable) SetSetting(id GameID, settingID string, v string) (err error) {
	t.safeExec(func() { err = t.setSetting(id, settingID, v) })

	return err
}

func (t *MemGameTable) RemoveSetting(id GameID, settingID string) (err error) {
	t.safeExec(func() { err = t.removeSetting(id, settingID) })

	return err
}

func (t *MemGameTable) ClearSettings(id GameID) (err error) {
	t.safeExec(func() { err = t.clearSettings(id) })

	return err
}

func (t *MemGameTable) FindServers(id GameID, f func(int, ServerData) bool) (output []ServerData, err error) {
	t.safeExec(func() { output, err = t.findServers(id, f) })

	return output, err
}

func (t *MemGameTable) AllServers(id GameID) (output []ServerData, err error) {
	t.safeExec(func() { output, err = t.findServers(id, func(int, ServerData) bool { return true }) })

	return output, err
}

func (t *MemGameTable) InsertServers(id GameID, data []ServerData) (err error) {
	t.safeExec(func() { err = t.insertServers(id, data) })

	return err
}

func (t *MemGameTable) DeleteServers(id GameID, f func(int, ServerData) bool) (deleted []ServerData, err error) {
	t.safeExec(func() { deleted, err = t.deleteServers(id, f) })

	return deleted, err
}

func (t *MemGameTable) ClearServers(id GameID) (err error) {
	t.safeExec(func() { _, err = t.deleteServers(id, func(int, ServerData) bool { return true }) })

	return err
}

func MakeMemGameTable() *MemGameTable {
	return &MemGameTable{semaphore: semaphore.MakeSemaphore(1), data: map[GameID]*GameEntry{}}
}
