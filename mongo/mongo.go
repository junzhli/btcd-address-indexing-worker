package mongo

import (
	"errors"

	"github.com/junzhli/btcd-address-indexing-worker/utils/logger"

	"github.com/go-bongo/bongo"
	"gopkg.in/mgo.v2/bson"
)

const dbUser string = "users"

// Mongo deals with mongo database stuffs
type Mongo interface {
	PutUserHistory(doc *UserHistory) error
	GetUserHistory(addr string) (*UserHistory, error)
}

type mongo struct {
	conn *bongo.Connection
}

// PutUserHistory stores provided document to database
func (m *mongo) PutUserHistory(doc *UserHistory) error {
	_doc := newUserHistoryModel(doc)
	return m.conn.Collection(dbUser).Save(&_doc)
}

// GetUserHistory fetches revalant address history from database
func (m *mongo) GetUserHistory(addr string) (*UserHistory, error) {
	var histories []userHistoryModel
	err := m.conn.Collection(dbUser).Find(bson.M{"address": addr}).Query.Sort("timestamp").All(&histories)
	if err != nil {
		logger.LogOnError(err, "Failed to fetch user history from database")
		return nil, err
	}

	if len(histories) == 0 {
		err := errors.New(ErrorNoUserInfo)
		return nil, err
	}

	lastIdx := len(histories) - 1
	timeStp := histories[lastIdx].Timestamp
	var subtotl int64
	spts := make(map[string]bool, 0)
	unsptAmts := make(map[string]uint64, 0)
	unspts := make([]Unspent, 0)
	shadowspts := make([]string, 0)
	txs := make([]string, 0)
	skipped := histories[lastIdx].Skipped

	for _, history := range histories {
		subtotl += history.Subtotal
		mergeSpents(spts, history.Spents)
		mergeUnspentAmts(unsptAmts, history.UnspentAmts)
		unspts = append(unspts, history.Unspents...)
		shadowspts = append(shadowspts, history.Shadowspents...)
		txs = append(txs, history.Transactions...)
	}

	return &UserHistory{
		Address:      addr,
		Timestamp:    timeStp,
		Subtotal:     subtotl,
		Spents:       spts,
		UnspentAmts:  unsptAmts,
		Unspents:     unspts,
		Shadowspents: shadowspts,
		Transactions: txs,
		Skipped:      skipped,
	}, nil
}

// New creates an instance of Mongo
func New(conn *bongo.Connection) Mongo {
	return &mongo{
		conn,
	}
}

func mergeSpents(a map[string]bool, b map[string]bool) error {
	for key, val := range b {
		if _, ok := a[key]; ok {
			err := errors.New("Conflict occurred during the merge op from b into a: " + key)
			return err
		}

		a[key] = val
	}
	return nil
}

func mergeUnspentAmts(a map[string]uint64, b map[string]uint64) error {
	for key, val := range b {
		if _, ok := a[key]; ok {
			err := errors.New("Conflict occurred during the merge op from b into a: " + key)
			return err
		}

		a[key] = val
	}
	return nil
}
