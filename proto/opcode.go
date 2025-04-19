package proto

import bbinWails "bbinWails/src"

const (
	OpcodePing       = 1
	OpcodeEcho       = 2
	OpcodeGetTickets = 1001
	OpcodeSyncData   = 1002
)

type SyncPayload struct {
	Lists   []bbinWails.DataList // 先用 any 或 interface{}，你可以用真实类型
	Leagues []*bbinWails.DataLeagueInfo
}

type GetTicketsRequest struct {
	MatchID string
	Items   []bbinWails.DataList
}

type GetTicketsResponse struct {
	MatchID string
	Result  any
}
