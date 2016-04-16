package main

type gamesRenderJSON struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Proxy    ProxyID     `json:"proxy"`
	Adapter  AdapterID   `json:"adapter"`
	Settings SettingsMap `json:"settings"`
}

type jsonResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Content interface{} `json:"content"`
}
