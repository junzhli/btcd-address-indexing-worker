package account_test

import (
	"log"
	"os"
	"testing"

	gomark "github.com/golang/mock/gomock"
	"github.com/junzhli/btcd-address-indexing-worker/account"
	mockBtcd "github.com/junzhli/btcd-address-indexing-worker/btcd/mocks"
	"github.com/junzhli/btcd-address-indexing-worker/logger"
	mockMongo "github.com/junzhli/btcd-address-indexing-worker/mongo/mocks"
	mockRedis "github.com/junzhli/btcd-address-indexing-worker/redis/mocks"
)

type vars struct {
	mongo   *mockMongo.MockMongo
	btcd    *mockBtcd.MockBtcd
	redis   *mockRedis.MockRedis
	account account.Account
}

func initVars(t *testing.T) vars {
	mockCtrl := gomark.NewController(t)
	defer mockCtrl.Finish()

	// configuration
	lg := log.New(os.Stdout, "[Task <testing>] ", log.LstdFlags)
	lg2 := logger.New(lg)
	mongo := mockMongo.NewMockMongo(mockCtrl)
	btcd := mockBtcd.NewMockBtcd(mockCtrl)
	rs := mockRedis.NewMockRedis(mockCtrl)
	config := account.Config{
		Btcd:  btcd,
		Mongo: mongo,
		Redis: rs,
	}
	acc := account.New(lg, lg2, &config)

	return vars{
		mongo:   mongo,
		btcd:    btcd,
		redis:   rs,
		account: acc,
	}
}

// func TestAccountGetBalance(t *testing.T) {
// 	v := initVars(t)
// }

// func TestAccountGetTransactions(t *testing.T) {

// }

// func TestAccountGetUnspents(t *testing.T) {

// }
