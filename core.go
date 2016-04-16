package main

const (
	ProxyQStatOutput = ProxyID("qstat_output")
	ProxyNetHTTP     = ProxyID("net_http")
	AdapterQStatXML  = AdapterID("qstat_xml")
)

func GetProxyID(id string) ProxyID {
	switch id {
	case string(ProxyQStatOutput):
		return ProxyQStatOutput
	case string(ProxyNetHTTP):
		return ProxyNetHTTP
	}

	return ProxyInvalid
}

func GetAdapterID(b string) AdapterID {
	switch b {
	case string(AdapterQStatXML):
		return AdapterQStatXML
	}

	return AdapterInvalid
}

// Core class of Obozrenie.
type Core struct {
	GameTable GameTable
	Proxies   *ProxyCollection
	Adapters  *AdapterCollection
}

func (c *Core) statMasterTarget(gameID GameID, cb func([]ServerData, error)) {
	var proxyFunc ProxyFunc
	var adapterFunc AdaptFunc
	var err error
	var result []ServerData

	queryLocked, err := c.GameTable.TryLockQuery(gameID)
	if err != nil {
		if cb != nil {
			cb(nil, err)
		}
	} else if queryLocked {
		var e *GameEntry
		if err == nil {
			e, err = c.GameTable.CopyGameEntry(gameID, false)
		}

		if err == nil {
			proxyID := e.Info.Proxy

			var pExists bool
			proxyFunc, pExists = c.Proxies.Retrieve(proxyID)

			if !pExists {
				err = errNoProxy
			}
		}

		var data []string
		if err == nil {
			data, err = proxyFunc(e.Info, e.Settings.AllSettings())
		}

		if err == nil && data == nil {
			err = errEmptyProxyData
		}

		if err == nil {
			result, err = adapterFunc(data, e.Info, e.Settings.AllSettings())
		}

		if err == nil {
			c.GameTable.InsertServers(gameID, result)
			c.GameTable.SetQueryStatus(gameID, QueryReady)
		} else {
			c.GameTable.SetQueryStatus(gameID, QueryError)
		}

		if cb != nil {
			cb(result, err)
		}
	}
}

// UpdateServerList refreshes server list for selected game.
func (c *Core) UpdateServerList(gameID GameID, cb func([]ServerData, error)) {
	go c.statMasterTarget(gameID, cb)
}

// StartGame executes launcher pattern for selected game and server.
func (c *Core) StartGame(gameID GameID, server string, password string) error {
	if gameID == "" {
		return errInvalidGameID
	}

	host, port, _, addrErr := ParseHostPort(server)
	if addrErr != nil {
		return addrErr
	}
	_ = host == port

	return nil
}

// StartCore creates the core instance.
func StartCore() *Core {
	c := Core{GameTable: MakeMemGameTable(), Proxies: MakeProxyCollection(), Adapters: MakeAdapterCollection()}

	c.Proxies.Insert(ProxyQStatOutput, GetQStatOutput)
	c.Adapters.Insert(AdapterQStatXML, AdaptQStatOutput)

	return &c
}
