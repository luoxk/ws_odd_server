package handler

import (
	bbinWails "bbinWails/src"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"
	"ws_odd_server/config"
	"ws_odd_server/models"
	"ws_odd_server/proto"
	"ws_odd_server/ws"
)

var (
	hub = ws.NewClientHub()
)

type snapshot struct {
	Lists   []bbinWails.DataList
	Leagues []*bbinWails.DataLeagueInfo
}

var (
	snapshotCache snapshot
	snapshotMu    sync.RWMutex
)

type EventHandler struct {
	OnConnected    func(c *ws.Client)
	OnDisconnected func(c *ws.Client)
}

var handlers = &EventHandler{
	OnConnected: func(c *ws.Client) {
		log.Printf("Client %d connected", c.ID)

		go func() {
			snapshotMu.RLock()
			payload := proto.SyncPayload{
				Lists:   snapshotCache.Lists,
				Leagues: snapshotCache.Leagues,
			}
			snapshotMu.RUnlock()

			resp, _ := proto.EncodeFrame(proto.OpcodeSyncData, payload, config.AESKey)
			c.Send <- resp
		}()
	},
	OnDisconnected: func(c *ws.Client) {
		log.Printf("Client %d disconnected", c.ID)
	},
}

func startSyncTask() {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		if models.GetBotInstance() == nil {
			continue
		}
		lists := getDataList()
		leagues := getLeagueInfo()

		snapshotMu.Lock()
		snapshotCache.Lists = lists
		snapshotCache.Leagues = leagues
		snapshotMu.Unlock()

		payload := proto.SyncPayload{
			Lists:   lists,
			Leagues: leagues,
		}
		resp, _ := proto.EncodeFrame(proto.OpcodeSyncData, payload, config.AESKey)
		hub.Broadcast(resp)
	}
}

func init() {
	go startSyncTask()
}

func getDataList() []bbinWails.DataList {
	var res = make([]bbinWails.DataList, 0)
	inst := models.GetBotInstance()
	if inst != nil {
		res = append(res, inst.GetLiveDataList(bbinWails.Today)...)
		res = append(res, inst.GetLiveDataList(bbinWails.Live)...)
		res = append(res, inst.GetLiveDataList(bbinWails.Soon)...)
	}
	return res
}

func getLeagueInfo() []*bbinWails.DataLeagueInfo {
	var res = make([]*bbinWails.DataLeagueInfo, 0)
	inst := models.GetBotInstance()
	if inst != nil {
		res = append(res, inst.GetLiveDataLeagueInfo(bbinWails.Today)...)
		res = append(res, inst.GetLiveDataLeagueInfo(bbinWails.Live)...)
		res = append(res, inst.GetLiveDataLeagueInfo(bbinWails.Soon)...)
	}
	return res
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.Upgrade(w, r)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	client := hub.Add(conn)
	if handlers.OnConnected != nil {
		handlers.OnConnected(client)
	}
	defer func() {
		hub.Remove(client.ID)
		if handlers.OnDisconnected != nil {
			handlers.OnDisconnected(client)
		}
	}()

	go ws.StartWriter(client)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		frame, err := proto.ParseFrame(msg)
		if err != nil {
			log.Println("parse failed:", err)
			continue
		}

		switch frame.Opcode {
		case proto.OpcodeGetTickets:
			var req proto.GetTicketsRequest
			if err := proto.DecodePayload(frame.Data, frame.IV, config.AESKey, &req); err != nil {
				continue
			}

			go func() {
				res := proto.GetTicketsResponse{
					MatchID: req.MatchID,
					Odds:    []float64{1.88, 3.25, 4.10},
				}
				resp, _ := proto.EncodeFrame(proto.OpcodeGetTickets, res, config.AESKey)
				client.Send <- resp
			}()
		}
	}
}

func HandleDebugSnapshot(w http.ResponseWriter, r *http.Request) {
	snapshotMu.RLock()
	defer snapshotMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"lists":   snapshotCache.Lists,
		"leagues": snapshotCache.Leagues,
	})
}
