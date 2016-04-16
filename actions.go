package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/skybon/multilogger"
)

type serverActions struct {
	password string
	core     *Core
	logs     *multilogger.LogCollection
}

func (s *serverActions) renderLogResponse(status int, message string, content interface{}, severity multilogger.LogMessageType, w http.ResponseWriter) {
	s.logs.Add(PrettyLogMessage(status, message, severity))
	renderResponse(status, message, content, w)
}

func (s *serverActions) renderError(w http.ResponseWriter, err error) {
	renderResponse(500, err.Error(), nil, w)
}

func (s *serverActions) renderLogError(w http.ResponseWriter, err error) {
	s.renderLogResponse(500, err.Error(), nil, multilogger.MSG_MAJOR, w)
}

func (s *serverActions) renderPing(w http.ResponseWriter, r *http.Request) {
	s.renderLogResponse(200, "Ping.", nil, multilogger.MSG_MINOR, w)
}

func (s *serverActions) renderShutdown(w http.ResponseWriter, r *http.Request) {
	s.renderLogResponse(200, "Server shutting down", nil, multilogger.MSG_MAJOR, w)
}

func (s *serverActions) renderLogs(w http.ResponseWriter, r *http.Request) {
	renderResponse(200, "OK.", map[string][]multilogger.LogMessage{"logs": s.logs.Logs()}, w)
}

func (s *serverActions) renderBatchParseResult(w http.ResponseWriter, errorMap map[string]error, severity multilogger.LogMessageType) {
	outMap := make(map[string]string)

	keys := make([]string, 0, len(errorMap))
	for k, v := range errorMap {
		keys = append(keys, k)
		if v != nil {
			outMap[k] = v.Error()
		} else {
			outMap[k] = "OK."
		}
	}

	s.renderLogResponse(200, fmt.Sprintf("Processed entries with IDs: %s", strings.Join(keys, ", ")), map[string]interface{}{"input_log": outMap}, severity, w)
}

func (s *serverActions) checkPassword(w http.ResponseWriter, r *http.Request, password string, cb func(w http.ResponseWriter, r *http.Request)) {
	if password == s.password || s.password == "" {
		cb(w, r)
	} else {
		s.renderError(w, errWrongPassword)
	}
}

func (s *serverActions) checkRequestPassword(w http.ResponseWriter, r *http.Request, cb func(w http.ResponseWriter, r *http.Request)) {
	var inputData gameEntryEditPost
	json.Unmarshal([]byte(retrievePostJSON(r)), &inputData)

	s.checkPassword(w, r, inputData.Password, cb)
}

func (s *serverActions) mergeGameEntry(entry gameEntryPost) error {
	id := *entry.ID
	if !s.core.GameTable.CheckGameEntry(id) {
		return errUnknownGameID
	}

	info, _ := s.core.GameTable.GameInfo(id)
	if entry.Name != nil {
		info.Name = *entry.Name
	}

	proxyP := entry.Proxy
	if proxyP != nil {
		proxyID := GetProxyID(*proxyP)

		info.Proxy = proxyID
	}

	adapterP := entry.Adapter
	if adapterP != nil {
		adapterID := GetAdapterID(*adapterP)

		info.Adapter = adapterID
	}
	s.core.GameTable.SetGameInfo(id, info)

	if entry.Settings != nil {
		s.core.GameTable.ClearSettings(id)
		for k, v := range entry.Settings {
			s.core.GameTable.SetSetting(id, k, v)
		}
	}

	return nil
}

func (s *serverActions) upsertGameEntryBase(w http.ResponseWriter, r *http.Request, create bool) {
	var inputData gameEntryEditPost
	json.Unmarshal([]byte(retrievePostJSON(r)), &inputData)

	var gameData = inputData.Data

	if len(gameData) == 0 {
		s.renderError(w, errNoGamesSpecified)
		return
	}

	var errorMap = map[string]error{}

	for _, entry := range gameData {
		func(entry gameEntryPost) {
			var err error

			var entryID GameID

			entryIDP := entry.ID
			if entryIDP == nil {
				return
			}
			entryID = *entryIDP

			if create {
				err = s.core.GameTable.CreateGameEntry(entryID)
			} else {
				exists := s.core.GameTable.CheckGameEntry(entryID)
				if !exists {
					err = errNoSuchGame
				}
			}

			if err == nil {
				err = s.mergeGameEntry(entry)
			}
			errorMap[entryID.String()] = err
		}(entry)
	}

	s.renderBatchParseResult(w, errorMap, multilogger.MSG_MAJOR)
}

func (s *serverActions) createGameEntries(w http.ResponseWriter, r *http.Request) {
	s.upsertGameEntryBase(w, r, true)
}

func (s *serverActions) updateGameEntry(w http.ResponseWriter, r *http.Request) {
	s.upsertGameEntryBase(w, r, false)
}

func (s *serverActions) deleteGameEntry(w http.ResponseWriter, r *http.Request) {
	var inputData gameEntryEditPost
	json.Unmarshal([]byte(retrievePostJSON(r)), &inputData)

	ids := inputData.IDs

	if ids == nil {
		s.renderLogError(w, errinvalidIDList)
	} else {
		outMap := make(map[string]string)
		for _, id := range ids {
			err := s.core.GameTable.RemoveGameEntry(GameID(id))
			if err == nil {
				outMap[id] = "OK"
			} else {
				outMap[id] = err.Error()
			}
		}

		s.renderLogResponse(200, fmt.Sprintf("Processed entries with IDs: %s", strings.Join(ids, ", ")), map[string]interface{}{"delete_log": outMap}, multilogger.MSG_MAJOR, w)
	}
}

func (s *serverActions) readGameCollection(w http.ResponseWriter, r *http.Request) {
	var games = s.core.GameTable.AllGames()
	var output = make([]gamesRenderJSON, 0, len(games))

	for _, id := range games {
		var outEntry gamesRenderJSON
		outEntry.ID = id.String()
		info, _ := s.core.GameTable.GameInfo(id)
		outEntry.Name = info.Name
		outEntry.Proxy = info.Proxy
		outEntry.Adapter = info.Adapter
		outEntry.Settings, _ = s.core.GameTable.Settings(id)

		output = append(output, outEntry)
	}

	s.renderLogResponse(200, "Games read from Game Table successful.", map[string]interface{}{"games": output}, multilogger.MSG_MAJOR, w)
}

func (s *serverActions) cleanup() {
	s.logs.Close()
}

func makeActionInstance(password string) *serverActions {
	return &serverActions{password: password, logs: multilogger.MakeLogCollection(multilogger.LoggingModes{Mem: true}, nil), core: StartCore()}
}

func makeServeMux(actions *serverActions, exitChan chan struct{}) *http.ServeMux {
	var sMux = http.NewServeMux()

	sMux.HandleFunc(systemPrefix+"/quit", func(w http.ResponseWriter, r *http.Request) {
		actions.renderShutdown(w, r)
		close(exitChan)
	})
	sMux.HandleFunc(systemPrefix+"/ping", actions.renderPing)
	sMux.HandleFunc(systemPrefix+"/logs", func(w http.ResponseWriter, r *http.Request) { actions.checkRequestPassword(w, r, actions.renderLogs) })
	sMux.HandleFunc(gameCollPrefix+"/create", func(w http.ResponseWriter, r *http.Request) {
		actions.checkRequestPassword(w, r, actions.createGameEntries)
	})
	sMux.HandleFunc(gameCollPrefix+"/read", func(w http.ResponseWriter, r *http.Request) {
		actions.checkRequestPassword(w, r, actions.readGameCollection)
	})
	sMux.HandleFunc(gameCollPrefix+"/update", func(w http.ResponseWriter, r *http.Request) {
		actions.checkRequestPassword(w, r, actions.updateGameEntry)
	})
	sMux.HandleFunc(gameCollPrefix+"/delete", func(w http.ResponseWriter, r *http.Request) {
		actions.checkRequestPassword(w, r, actions.deleteGameEntry)
	})

	return sMux
}
