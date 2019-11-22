package mongo

import (
	"time"

	"github.com/go-bongo/bongo"
)

// Unspent is used for the store of User's unspent transaction information, kept in UserHistory
type Unspent struct {
	Transaction  string
	VOutIdx      uint64
	ScriptPubKey string
	Amount       uint64
	BlockTime    uint64
}

// UserHistory keeps all revelant information about balance, transaction history, unspent...
// Note that it can also be used for the struct of user data in redis, well implemented in json representation as bytes array
type UserHistory struct {
	Address      string            `json:"a"`
	Timestamp    time.Time         `json:"t"`
	Subtotal     int64             `json:"sbtl"`
	Spents       map[string]bool   `json:"spts"`
	UnspentAmts  map[string]uint64 `json:"usptsams"`
	Unspents     []Unspent         `json:"uspts"`
	Shadowspents []string          `json:"sdspts"`
	Transactions []string          `json:"txs"`
	Skipped      uint64            `json:"skd"`
}

// userHistoryModel offers UserHistory with additional implementation in compliance with mongo model spec
type userHistoryModel struct {
	bongo.DocumentBase `bson:",inline"`
	Address            string
	Timestamp          time.Time
	Subtotal           int64
	Spents             map[string]bool
	UnspentAmts        map[string]uint64
	Unspents           []Unspent
	Shadowspents       []string
	Transactions       []string
	Skipped            uint64
}

func newUserHistoryModel(d *UserHistory) userHistoryModel {
	return userHistoryModel{
		Address:      d.Address,
		Timestamp:    d.Timestamp,
		Subtotal:     d.Subtotal,
		Spents:       d.Spents,
		UnspentAmts:  d.UnspentAmts,
		Unspents:     d.Unspents,
		Shadowspents: d.Shadowspents,
		Transactions: d.Transactions,
		Skipped:      d.Skipped,
	}
}
