package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
)

type qStatServerRule struct {
	Name  *string `xml:"name,attr"`
	Value *string `xml:",innerxml"`
}

func (r qStatServerRule) ToKV() (string, string, error) {
	k := r.Name
	v := r.Value
	if k == nil || v == nil {
		return "", "", errMalformedEntry
	}

	return *k, *v, nil
}

type qStatPlayer struct {
	Name  *string `xml:"name"`
	Score *string `xml:"score"`
	Ping  *string `xml:"ping"`
}

func (e qStatPlayer) ToPlayerData() (PlayerData, error) {
	qName := e.Name
	qScore := e.Score
	qPing := e.Ping

	if qName == nil {
		return PlayerData{}, errMalformedEntry
	}

	newEntry := NewPlayerData()
	newEntry.Name = *qName

	if qScore != nil {
		newEntry.Info["score"] = *qScore
	}

	if qPing != nil {
		newEntry.Info["ping"] = *qPing
	}

	return newEntry, nil
}

type qStatServer struct {
	ServerType    *string           `xml:"type,attr"`
	Hostname      *string           `xml:"address,attr"`
	Status        *string           `xml:"status,attr"`
	Name          *string           `xml:"name"`
	GameType      *string           `xml:"gametype"`
	Map           *string           `xml:"map"`
	NumPlayers    *int              `xml:"numplayers"`
	MaxPlayers    *int              `xml:"maxplayers"`
	NumSpectators *int              `xml:"numspectators"`
	MaxSpectators *int              `xml:"maxspectators"`
	Ping          *int              `xml:"ping"`
	Retries       *int              `xml:"retries"`
	Rules         []qStatServerRule `xml:"rules>rule"`
	Players       []qStatPlayer     `xml:"players>player"`
}

type qStatOutput struct {
	Servers []qStatServer `xml:"server"`
}

func (d *qStatOutput) Equal(cd qStatOutput) bool {
	dJSON, dErr := json.Marshal(d)
	cdJSON, cdErr := json.Marshal(cd)

	switch {
	case dErr != nil:
		panic(dErr)
	case cdErr != nil:
		panic(cdErr)
	}

	return bytes.Equal(dJSON, cdJSON)
}

func loadQStatXML(qstatString string) (output qStatOutput, err error) {
	err = xml.Unmarshal([]byte(qstatString), &output)
	return output, err
}

func (e qStatServer) ToServerData() (ServerData, error) {
	if e.Hostname == nil || e.Status == nil || e.Name == nil || e.Map == nil || e.NumPlayers == nil || e.MaxPlayers == nil {
		return ServerData{}, errMalformedEntry
	}

	newEntry := NewServerData()
	newEntry.Status = *e.Status
	newEntry.Host = *e.Hostname
	newEntry.Name = *e.Name
	newEntry.Map = *e.Map
	newEntry.NumPlayers = *e.NumPlayers
	newEntry.MaxPlayers = *e.MaxPlayers

	qRules := e.Rules
	if qRules != nil {
		for _, r := range qRules {
			k, v, rErr := r.ToKV()
			if rErr == nil {
				newEntry.Settings[k] = v
			}
		}
	}

	qPlayers := e.Players
	if qPlayers != nil {
		for _, p := range qPlayers {
			e, pErr := p.ToPlayerData()
			if pErr == nil {
				newEntry.Players = append(newEntry.Players, e)
			}
		}
	}

	return newEntry, nil
}

func AdaptQStatOutput(qstatStringSlice []string, i GameInfo, s SettingsMap) ([]ServerData, error) {
	output := []ServerData{}

	for _, qstatString := range qstatStringSlice {
		qstatData, parseErr := loadQStatXML(qstatString)
		if parseErr != nil {
			return nil, parseErr
		}

		for _, v := range qstatData.Servers {
			sData, sDataErr := v.ToServerData()
			if sDataErr == nil {
				output = append(output, sData)
			}
		}
	}

	return output, nil
}
