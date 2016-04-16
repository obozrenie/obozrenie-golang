package main

type gameEntryPost struct {
	ID       *GameID           `json:"id"`
	Proxy    *string           `json:"proxy"`
	Adapter  *string           `json:"adapter"`
	Name     *string           `json:"name"`
	Settings map[string]string `json:"settings"`
}

type gameEntryEditPost struct {
	Password  string          `json:"password"`
	Data      []gameEntryPost `json:"games"`
	IDs       []string        `json:"ids"`
	NotifyURL string          `json:"notify_url"`
}
