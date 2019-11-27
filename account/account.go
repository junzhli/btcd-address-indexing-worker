package account

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/junzhli/btcd-address-indexing-worker/mongo"
	rs "github.com/junzhli/btcd-address-indexing-worker/redis"
	rsmgo "github.com/junzhli/btcd-address-indexing-worker/redis/mongo"
	"github.com/junzhli/btcd-address-indexing-worker/redis/utils"
	"github.com/junzhli/btcd-address-indexing-worker/utils/btcd"
	"github.com/junzhli/btcd-address-indexing-worker/utils/logger"

	"github.com/go-bongo/bongo"
	"github.com/go-redis/redis"
)

// Config includes all necessary arguments during operation
type Config struct {
	Btcd        btcd.Btcd
	MongoClient *bongo.Connection
	RedisClient *redis.Client
}

const maxRequestedTransactionsRecord = 2000
const requiredConfirmations = 6
const satoshi float64 = 100000000

type userData struct {
	Transactions []string
	Unspents     []*mongo.Unspent
	Spents       map[string]*bool
	Total        int64
}

// UserData is ideal data schema for 'GetAddressResult'
type UserData struct {
	Balance      float64         `json:"balance"`
	Transactions []string        `json:"transactions"`
	Unspents     []mongo.Unspent `json:"unspents"`
}

// ToJSON encodes itself to JSON string represented in []byte
func (u UserData) ToJSON() ([]byte, error) {
	res, err := json.Marshal(u)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func containsAddr(arr []string, target string) bool {
	for _, item := range arr {
		if target == item {
			return true
		}
	}
	return false
}

func createUserHistory(
	addr string,
	unspts []*mongo.Unspent,
	unsptAmts map[string]uint64,
	spts map[string]bool,
	shadowspts []string,
	txs []string,
	skpt uint64,
	subtotal int64,
) *mongo.UserHistory {
	_unspts := make([]mongo.Unspent, 0)

	for _, unspt := range unspts {
		_unspts = append(_unspts, *unspt)
	}

	return &mongo.UserHistory{
		Address:      addr,
		Timestamp:    time.Now(),
		Subtotal:     subtotal,
		UnspentAmts:  unsptAmts,
		Unspents:     _unspts,
		Spents:       spts,
		Shadowspents: shadowspts,
		Transactions: txs,
		Skipped:      skpt,
	}
}

// combineUserHistory combines a1 and a2 into an instance of UserHistory
// Note that newer data should put as parameter a2 following a1 so that it results in properly order
// a1 and a2 must not be null at the same time
func combineUserHistory(a1 *mongo.UserHistory, a2 *mongo.UserHistory) (*mongo.UserHistory, error) {
	if a1 == nil && a2 != nil {
		return a2, nil
	}

	if a2 == nil && a1 != nil {
		return a1, nil
	}

	cSpts := a1.Spents
	for key, val := range a2.Spents {
		if _, ok := cSpts[key]; ok {
			err := errors.New("Conflict occurred during the merge op from a2 into a1: " + key)
			return nil, err
		}

		cSpts[key] = val
	}

	cUsptAmts := a1.UnspentAmts
	for key, val := range a2.UnspentAmts {
		if _, ok := cUsptAmts[key]; ok {
			err := errors.New("Conflict occurred during the merge op from b2 into a1: " + key)
			return nil, err
		}

		cUsptAmts[key] = val
	}

	return &mongo.UserHistory{
		Address:      a1.Address,
		Timestamp:    a2.Timestamp,
		Subtotal:     a1.Subtotal + a2.Subtotal,
		Spents:       cSpts,
		UnspentAmts:  cUsptAmts,
		Unspents:     append(a1.Unspents, a2.Unspents...),
		Shadowspents: append(a1.Shadowspents, a2.Shadowspents...),
		Transactions: append(a1.Transactions, a2.Transactions...),
		Skipped:      a2.Skipped,
	}, nil
}

func mergeSpents(a []map[string]*bool, b map[string]bool) error {
	for key, val := range b {
		spt := val
		for _, spents := range a {
			if _, ok := spents[key]; ok {
				err := errors.New("Conflict occurred during the merge op from b into a: " + key)
				return err
			}
			spents[key] = &spt
		}
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

func referenceUnspents(a []mongo.Unspent) []*mongo.Unspent {
	res := make([]*mongo.Unspent, 0)
	for _, unspt := range a {
		_unspt := unspt
		res = append(res, &_unspt)
	}
	return res
}

func restoreSpentStates(result map[string]*bool, shadowSpts []string) error {
	for _, key := range shadowSpts {
		spent, ok := result[key]
		if !ok {
			err := errors.New("Key: " + key + " doesn't exist in the given map")
			return err
		}

		if *spent {
			err := errors.New("Double spent at key: " + key)
			return err
		}
		*spent = true
	}
	return nil
}

func removeStateKeyRedis(config *Config, key string) {
	err := config.RedisClient.Del(key).Err()
	if err != nil {
		logger.FailOnError(err, "Failed to remove key in redis: key => "+key)
	}
}

// manipulateUserData processes raw data from database and btcd jsonRpc service as follows
// it keeps data in database up to date by appending newly update instead of replacing the old one for consistency and performance improvement
// it only appends data confirmed at least n 'confirmations' which is defined in 'account.go' to database
// otherwise, other data are always gathering from btcd and then merge them into data from database processing on-the-air for serving real-time data
func manipulateUserData(lg *log.Logger, lg2 logger.CustomLogger, config *Config, targetAddr string) (*userData, error) {
	key := utils.GenStateKey(targetAddr, rs.CommandAll)
	defer removeStateKeyRedis(config, key)
	// pre-checks
	state, err := config.RedisClient.Get(key).Result()
	if err == redis.Nil {
		lg2.LogOnError(err, "Could not find key existing in redis: key => "+key)
		return nil, err
	}
	if err != nil {
		lg2.LogOnError(err, "Fails on checking whether the key exists on redis: key => "+key)
		return nil, err
	}

	subtotalAll := int64(0)
	transactionsAll := make([]string, 0)
	spentsAll := make(map[string]*bool, 0)
	unspentAmtsAll := make(map[string]uint64, 0)
	unspentsAll := make([]*mongo.Unspent, 0)
	skipped := uint64(0)

	// predb
	subtotalPreDB := int64(0)
	transactionsPreDB := make([]string, 0)
	spentsPreDB := make(map[string]*bool, 0)
	// unspentAmtsPreDB := make(map[string]float64, 0)
	// unspentsPreDB := make([]*mongo.Unspent, 0)

	var startTime time.Time
	var elapsedTime time.Duration
	var preDB *mongo.UserHistory
	fetchFromDB := false
	if state == rs.StateAlreadyExisting {
		new := false
		// redis
		startTime = time.Now()
		key = utils.GenCacheKey(targetAddr, rs.CommandAll)
		preDB, err = rsmgo.RestoreUserHistory(config.RedisClient, key)

		if err != nil {
			if err == redis.Nil {
				lg.Println("Cached data is unavailable")
				fetchFromDB = true
			} else {
				lg2.LogOnError(err, "Error occurred on the request of fetching cached data from redis")
				return nil, err
			}
		}
		elapsedTime = time.Since(startTime)
		lg.Println("Data accessed from redis takes " + elapsedTime.String())

		// database
		if fetchFromDB {
			startTime = time.Now()
			preDB, err = mongo.GetUserHistory(config.MongoClient, targetAddr)

			if err != nil {
				if err.Error() == mongo.ErrorNoUserInfo {
					new = true
				} else {
					lg2.LogOnError(err, "Fails on the request of cached user detailed transaction history")
					return nil, err
				}
			}
			elapsedTime = time.Since(startTime)
			lg.Println("Data accessed from database takes " + elapsedTime.String())
		}

		startTime = time.Now()
		if !new {
			subtotalPreDB = preDB.Subtotal
			subtotalAll += subtotalPreDB
			transactionsAll = append(transactionsAll, preDB.Transactions...)
			transactionsPreDB = append(transactionsPreDB, preDB.Transactions...)
			mergeSpents([]map[string]*bool{spentsAll, spentsPreDB}, preDB.Spents)
			mergeUnspentAmts(unspentAmtsAll, preDB.UnspentAmts)
			// mergeUnspentAmts(unspentAmtsPreDB, preDB.UnspentAmts)
			unspentsAll = referenceUnspents(preDB.Unspents)
			// copy(unspentsPreDB, unspentsAll)
			skipped = preDB.Skipped
			restoreSpentStates(spentsAll, preDB.Shadowspents)
		}
		elapsedTime = time.Since(startTime)
		lg.Println("Preparing data from database/redis takes " + elapsedTime.String())
	} else if state == rs.StateNew {
		lg.Println("New address detected... bypass query for database/redis")
	} else {
		lg2.FailOnError(errors.New("Unknown state"), "Unsupported state on redis")
	}

	// db, memory
	node := config.Btcd

	alldone := false
	transactionsDB := make([]string, 0)
	// transactionsNonDB := make([]string, 0)
	spentsDB := make(map[string]*bool, 0)
	spentsDBPersistent := make(map[string]bool, 0)
	spentsNonDB := make(map[string]*bool, 0)
	unspentAmtsDB := make(map[string]uint64, 0)
	unspentsDB := make([]*mongo.Unspent, 0)
	// unspentsNonDB := make([]*mongo.Unspent, 0)
	shadowSpentsDB := make([]string, 0)
	subtotalDB := int64(0)

	// process non db part and memory part
	startTime2 := time.Now()
	start := int64(skipped)
	for !alldone {
		startTime = time.Now()
		res, err := node.SearchRawTransactions(targetAddr, start, maxRequestedTransactionsRecord)
		elapsedTime = time.Since(startTime)
		lg.Println("Fetching data from btcd takes " + elapsedTime.String())
		if err != nil {
			if err.Error() == btcd.ErrorNoDataReturned {
				break
			}
			lg2.LogOnError(err, "Fails on the request of user detailed transaction history")
			return nil, err
		}

		txsLen := len(*res)
		if txsLen < maxRequestedTransactionsRecord {
			alldone = true
		}

		for _, tx := range *res {
			blocktime := tx.Blocktime
			cfms := tx.Confirmations // confirmations of the transaction
			if cfms == 0 {
				continue // skip all unconfirmed transactions
			}

			persistent := false
			if cfms > requiredConfirmations {
				persistent = true
			}

			if persistent {
				transactionsDB = append(transactionsDB, tx.Txid)
			}
			// } else {
			// transactionsNonDB = append(transactionsNonDB, tx.Txid)
			// }
			transactionsAll = append(transactionsAll, tx.Txid)

			for idx, vout := range tx.Vouts {
				// Unspent is used for the store of User's unspent transaction information, kept in UserHistory
				// type Unspent struct {
				// 	Transaction   string
				// 	VOut          uint64
				// 	ScriptPubKey  string
				// 	Amount        float64
				//  BlockTime     uint64
				// }
				if containsAddr(vout.ScriptPubKey.Addresses, targetAddr) {
					unspent := mongo.Unspent{
						Transaction:  tx.Txid,
						VOutIdx:      uint64(idx),
						ScriptPubKey: vout.ScriptPubKey.Hex,
						Amount:       uint64(math.Round(vout.Value * satoshi)),
						BlockTime:    blocktime,
					}

					key := unspent.Transaction + "+" + strconv.FormatUint(unspent.VOutIdx, 10)
					spent := false
					amt := int64(unspent.Amount)
					if persistent {
						unspentsDB = append(unspentsDB, &unspent)
						spentsDB[key] = &spent
						spentsDBPersistent[key] = spent

						unspentAmtsDB[key] = unspent.Amount
						subtotalDB += amt
					} else {
						// unspentsNonDB = append(unspentsNonDB, &unspent)
						spentsNonDB[key] = &spent
					}
					subtotalAll += amt
					unspentAmtsAll[key] = unspent.Amount
					unspentsAll = append(unspentsAll, &unspent)
					spentsAll[key] = &spent
				}
			}

			for _, vin := range tx.Vins {
				if containsAddr(vin.PrevOut.Addresses, targetAddr) {
					key := vin.Txid + "+" + strconv.FormatUint(vin.VoutIndex, 10)
					amt := int64(unspentAmtsAll[key])
					var spent *bool
					var ok bool
					if persistent {
						spent, ok = spentsPreDB[key]
						if !ok {
							spent, ok = spentsDB[key]
							if !ok {
								err := errors.New("Cannot find key " + key + " on 'spentsPreDB/spentsDB'")
								lg2.LogOnError(err, "Should exist this unspent key on 'spentsPreDB/spentsDB'. Corrupted database?")
								return nil, err
							}

							spentPersistent, ok := spentsDBPersistent[key]
							if ok && spentPersistent {
								err := errors.New("Double spent at key: " + key)
								lg2.LogOnError(err, "Should not spend spent fund! Corrupted database?")
								return nil, err
							}
							spentsDBPersistent[key] = true
						} else {
							shadowSpentsDB = append(shadowSpentsDB, key)
						}
						subtotalDB -= amt
					} else {
						spent, ok = spentsPreDB[key]
						if !ok {
							spent, ok = spentsDB[key]
							if !ok {
								spent, ok = spentsNonDB[key]
								if !ok {
									err := errors.New("Cannot find key " + key + " on 'spentsPreDB/spentsDB/spentsNonDB'")
									lg2.LogOnError(err, "Should exist this unspent key on 'spentsPreDB/spentsDB/spentsNonDB'. Corrupted database?")
									return nil, err
								}
							}
						}
					}
					if *spent {
						err := errors.New("Double spent at key: " + key)
						lg2.LogOnError(err, "Should not spend spent fund! Corrupted database?")
						return nil, err
					}
					*spent = true
					subtotalAll -= amt
				}
			}
		}

		if txsLen < maxRequestedTransactionsRecord {
			start += int64(txsLen)
		} else {
			start += maxRequestedTransactionsRecord
		}
	}
	elapsedTime2 := time.Since(startTime2)
	lg.Println("Data accessed from btcd totally takes " + elapsedTime2.String())

	skipped = uint64(len(transactionsDB) + len(transactionsPreDB))

	var usrHistory *mongo.UserHistory
	if len(transactionsDB) != 0 || fetchFromDB {
		if len(transactionsDB) != 0 {
			startTime = time.Now()
			usrHistory = createUserHistory(targetAddr, unspentsDB, unspentAmtsDB, spentsDBPersistent, shadowSpentsDB, transactionsDB, skipped, subtotalDB)
			elapsedTime = time.Since(startTime)
			lg.Println("The task requested to prepare for UserHistory takes " + elapsedTime.String())

			startTime = time.Now()
			err = mongo.PutUserHistory(config.MongoClient, usrHistory)
			elapsedTime = time.Since(startTime)
			lg.Println("Document creation on database takes " + elapsedTime.String())
			if err != nil {
				lg2.LogOnError(err, "Fails on the document creation in database")
				return nil, err
			}
		} else {
			lg.Println("No need to create/update document on database")
		}

		// redis
		startTime = time.Now()
		cachedUsrHistory, err := combineUserHistory(preDB, usrHistory)
		if err != nil {
			lg2.LogOnError(err, "Fails on the creation of cached data for redis")
			return nil, err
		}
		elapsedTime = time.Since(startTime)
		lg.Println("The creation of cached data for redis takes " + elapsedTime.String())

		startTime = time.Now()
		key = utils.GenCacheKey(targetAddr, rs.CommandAll)
		err = rsmgo.CacheUserHistory(config.RedisClient, key, cachedUsrHistory, 3600*time.Second)
		if err != nil {
			lg2.LogOnError(err, "Fails on updating cached data on redis... trying to remove cached data on redis")
			err = config.RedisClient.Del(key).Err()
			if err != nil {
				lg2.LogOnError(err, "Fails on removing cached data on redis")
				return nil, err
			}
			return nil, err
		}
		lg.Println("The creation of cached data on redis takes " + elapsedTime.String())
	}

	res := userData{
		Unspents:     unspentsAll,
		Spents:       spentsAll,
		Transactions: transactionsAll,
		Total:        subtotalAll,
	}
	return &res, nil
}

// GetAddressBalance returns the balance of the given account
func GetAddressBalance(lg *log.Logger, lg2 logger.CustomLogger, config *Config, addr string) (float64, error) {
	uData, err := manipulateUserData(lg, lg2, config, addr)
	if err != nil {
		return 0, err
	}
	return float64(uData.Total) / float64(satoshi), nil
}

// GetAddressTransactions returns the list of transaction ids with the given account
func GetAddressTransactions(lg *log.Logger, lg2 logger.CustomLogger, config *Config, addr string) ([]string, error) {
	uData, err := manipulateUserData(lg, lg2, config, addr)
	if err != nil {
		return nil, err
	}
	return uData.Transactions, nil
}

// GetAddressUnspentOutputs returns the unspent outputs of the given account
func GetAddressUnspentOutputs(lg *log.Logger, lg2 logger.CustomLogger, config *Config, addr string) ([]*mongo.Unspent, error) {
	uData, err := manipulateUserData(lg, lg2, config, addr)
	if err != nil {
		return nil, err
	}

	unspents := genUTXO(uData.Unspents, uData.Spents)
	return unspents, nil
}

// GetAddressResult returns details for the given address
func GetAddressResult(lg *log.Logger, lg2 logger.CustomLogger, config *Config, addr string) (*UserData, error) {
	uData, err := manipulateUserData(lg, lg2, config, addr)
	if err != nil {
		return nil, err
	}

	unspents := genUTXO(uData.Unspents, uData.Spents)
	_unspents := make([]mongo.Unspent, 0)
	for _, val := range unspents {
		_unspents = append(_unspents, *val)
	}

	res := UserData{
		Balance:      float64(uData.Total) / float64(satoshi),
		Transactions: uData.Transactions,
		Unspents:     _unspents,
	}
	return &res, nil
}

func genUTXO(outputs []*mongo.Unspent, spents map[string]*bool) []*mongo.Unspent {
	unspents := make([]*mongo.Unspent, 0)
	for _, unspt := range outputs {
		key := unspt.Transaction + "+" + strconv.FormatUint(unspt.VOutIdx, 10)
		spent := spents[key]
		if !(*spent) {
			unspents = append(unspents, unspt)
		}
	}
	return unspents
}
