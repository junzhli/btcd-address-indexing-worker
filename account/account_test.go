package account_test

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	gomark "github.com/golang/mock/gomock"
	"github.com/junzhli/btcd-address-indexing-worker/account"
	"github.com/junzhli/btcd-address-indexing-worker/btcd"
	mockBtcd "github.com/junzhli/btcd-address-indexing-worker/btcd/mocks"
	"github.com/junzhli/btcd-address-indexing-worker/logger"
	"github.com/junzhli/btcd-address-indexing-worker/mongo"
	mockMongo "github.com/junzhli/btcd-address-indexing-worker/mongo/mocks"
	rs "github.com/junzhli/btcd-address-indexing-worker/redis"
	mockRedis "github.com/junzhli/btcd-address-indexing-worker/redis/mocks"
	"github.com/junzhli/btcd-address-indexing-worker/redis/utils"
)

const address = "15a7wZQhCeQ457KzxRZbeJ8jobb6yMVubR"

type vars struct {
	mongo   *mockMongo.MockMongo
	btcd    *mockBtcd.MockBtcd
	redis   *mockRedis.MockRedis
	account account.Account
}

func txsIsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func unspentsIsEqual(a []*mongo.Unspent, b []mongo.Unspent) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if *v != b[i] {
			return false
		}
	}
	return true
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

func initMocks(env *vars) {
	// mocks function returns in sequence
	stateKey := utils.GenStateKey(address, rs.CommandAll)
	env.redis.EXPECT().Get(stateKey).Return(rs.StateNew, nil).Times(1)

	rawTxs := `[
			{
				"hex": "0100000001b9e1a16124f3ecf3cd0c8d5317e68dc70b655bed957bca8344d820c021dd71d8010000006a4730440220174f03086b54518633d918e9ddb1b5263fe8384470cd081bf92d16fbe5bde87302204818d62565c179de1b9860edd04e666b25f12b12f4461ff2b3985aac30e21045012103e1dfb8175d7be1e64a41e5ff8da17ee90a3c91af33c0797a660339145313ef8effffffff02e072a705000000001976a9144a3681ee9e3451bd4c24f68eafb65ec832e9ec0e88ac3e689409000000001976a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac00000000",
				"txid": "5cf66fd258a5d2be04051134b11bd794a65e149bbbfc1f32e8f18841997e936d",
				"hash": "",
				"size": "",
				"vsize": "",
				"weight": "",
				"version": 1,
				"locktime": 0,
				"vin": [
					{
						"txid": "d871dd21c020d84483ca7b95ed5b650bc78de617538d0ccdf3ecf32461a1e1b9",
						"vout": 1,
						"scriptSig": {
							"asm": "30440220174f03086b54518633d918e9ddb1b5263fe8384470cd081bf92d16fbe5bde87302204818d62565c179de1b9860edd04e666b25f12b12f4461ff2b3985aac30e2104501 03e1dfb8175d7be1e64a41e5ff8da17ee90a3c91af33c0797a660339145313ef8e",
							"hex": "4730440220174f03086b54518633d918e9ddb1b5263fe8384470cd081bf92d16fbe5bde87302204818d62565c179de1b9860edd04e666b25f12b12f4461ff2b3985aac30e21045012103e1dfb8175d7be1e64a41e5ff8da17ee90a3c91af33c0797a660339145313ef8e"
						},
						"prevOut": {
							"addresses": [
								"1A5ehPU5W3VxkuvKWLSyYdAfK2YMdsJiaq"
							],
							"value": 2.55630958
						},
						"sequence": 4294967295
					}
				],
				"vout": [
					{
						"value": 0.9486,
						"n": 0,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 4a3681ee9e3451bd4c24f68eafb65ec832e9ec0e OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a9144a3681ee9e3451bd4c24f68eafb65ec832e9ec0e88ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"17mQJSt7v2w2FTrP8MnBjTBPffgVgBdkJ3"
							]
						}
					},
					{
						"value": 1.60720958,
						"n": 1,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 3224060e14d6cf0d2e225a2a2f3aa8779de4226b OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"15a7wZQhCeQ457KzxRZbeJ8jobb6yMVubR"
							]
						}
					}
				],
				"blockhash": "00000000000000000025bbf5ebe2ab7e424d10afb4857270695ae6403b068a06",
				"confirmations": 58145,
				"time": 1540994884,
				"blocktime": 1540994884
			},
			{
				"hex": "01000000016d937e994188f1e8321ffcbb9b145ea694d71bb134110504bed2a558d26ff65c010000006a47304402205e8f0d570d3cdbccdcd9c9f34de4881058467837e947efa2d0c517ddd06c2292022043899065bfa8795a851c54344e8940323c9a00304ef0332f1d19ba28a256cc6501210330a8a1ab91531b57d4883181f98038dc3bc2a2b4a8cb18dc8e57a09c3d2932bfffffffff02106549000000000017a91466d7080ddfe69e5803d5b40548f8c1175d84f80387de3f4a09000000001976a9149d84a76f0a5715043c19d295f1351e75deb211c888ac00000000",
				"txid": "f74918c59110c5389c5b935d01e54428eb33e6180deb90abef22cd8d8100e3ff",
				"hash": "",
				"size": "",
				"vsize": "",
				"weight": "",
				"version": 1,
				"locktime": 0,
				"vin": [
					{
						"txid": "5cf66fd258a5d2be04051134b11bd794a65e149bbbfc1f32e8f18841997e936d",
						"vout": 1,
						"scriptSig": {
							"asm": "304402205e8f0d570d3cdbccdcd9c9f34de4881058467837e947efa2d0c517ddd06c2292022043899065bfa8795a851c54344e8940323c9a00304ef0332f1d19ba28a256cc6501 0330a8a1ab91531b57d4883181f98038dc3bc2a2b4a8cb18dc8e57a09c3d2932bf",
							"hex": "47304402205e8f0d570d3cdbccdcd9c9f34de4881058467837e947efa2d0c517ddd06c2292022043899065bfa8795a851c54344e8940323c9a00304ef0332f1d19ba28a256cc6501210330a8a1ab91531b57d4883181f98038dc3bc2a2b4a8cb18dc8e57a09c3d2932bf"
						},
						"prevOut": {
							"addresses": [
								"15a7wZQhCeQ457KzxRZbeJ8jobb6yMVubR"
							],
							"value": 1.60720958
						},
						"sequence": 4294967295
					}
				],
				"vout": [
					{
						"value": 0.0481,
						"n": 0,
						"scriptPubKey": {
							"asm": "OP_HASH160 66d7080ddfe69e5803d5b40548f8c1175d84f803 OP_EQUAL",
							"hex": "a91466d7080ddfe69e5803d5b40548f8c1175d84f80387",
							"reqSigs": 1,
							"type": "scripthash",
							"addresses": [
								"3B4nSkwKYhW9ojUArcJTqRrF5SXKEpafv7"
							]
						}
					},
					{
						"value": 1.55860958,
						"n": 1,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 9d84a76f0a5715043c19d295f1351e75deb211c8 OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a9149d84a76f0a5715043c19d295f1351e75deb211c888ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"1FMt15jFr5S7Bbu9rjcVS8YzCEYogKvaHz"
							]
						}
					}
				],
				"blockhash": "0000000000000000001611ba8bd688c885731ae1aeb928078af3d83af79d6dcf",
				"confirmations": 58144,
				"time": 1540996105,
				"blocktime": 1540996105
			},
			{
				"hex": "0100000002a8d1d5d799e28ff633dcef6ba2b148d24caf622f81ae6e7ad3180444b1e3faf8010000006a47304402206af5f8fd0b8edb757370b342e52139607ab0b17ead47a2c1aa869de0cdd422490220196082ef7019b449061ad6b0b2f7ffbec2e0d1c15e62ebf302cb4aae1a29cf200121021cdcd04f2cc3cae0cbe2f8b8beef70d14de9601c44ed4c29a2608eec8ab54862ffffffff2f046f56536cc110895862e5055c7be553949ceebb401007101d9ad9299ebf8b010000006a473044022010c56e6f9d6b797d50d98e084483c4b4717916411496a32d9b1a55992f528b4a02206452ce7049a669c83917d77135e60783cd5b7f5d10975f2d022cc420c3761f5d012103af061ee5118bcf2835d7a7761608e7d172c93377aa3516d97a6fd350a4ab30f1ffffffff023001600f000000001976a91497e222cce73e42c6e3baa643500cffaa9d2090a388ac3c6c112d000000001976a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac00000000",
				"txid": "e47ff4d45664d31e5c2f7886be56c55d96ff09b7bc39a3eb6de759f219f77f07",
				"hash": "",
				"size": "",
				"vsize": "",
				"weight": "",
				"version": 1,
				"locktime": 0,
				"vin": [
					{
						"txid": "f8fae3b1440418d37a6eae812f62af4cd248b1a26befdc33f68fe299d7d5d1a8",
						"vout": 1,
						"scriptSig": {
							"asm": "304402206af5f8fd0b8edb757370b342e52139607ab0b17ead47a2c1aa869de0cdd422490220196082ef7019b449061ad6b0b2f7ffbec2e0d1c15e62ebf302cb4aae1a29cf2001 021cdcd04f2cc3cae0cbe2f8b8beef70d14de9601c44ed4c29a2608eec8ab54862",
							"hex": "47304402206af5f8fd0b8edb757370b342e52139607ab0b17ead47a2c1aa869de0cdd422490220196082ef7019b449061ad6b0b2f7ffbec2e0d1c15e62ebf302cb4aae1a29cf200121021cdcd04f2cc3cae0cbe2f8b8beef70d14de9601c44ed4c29a2608eec8ab54862"
						},
						"prevOut": {
							"addresses": [
								"1FvcnF2yymDXH5upTFaoMZXJQF8Bq8PyDD"
							],
							"value": 10
						},
						"sequence": 4294967295
					},
					{
						"txid": "8bbf9e29d99a1d10071040bbee9c9453e57b5c05e562588910c16c53566f042f",
						"vout": 1,
						"scriptSig": {
							"asm": "3044022010c56e6f9d6b797d50d98e084483c4b4717916411496a32d9b1a55992f528b4a02206452ce7049a669c83917d77135e60783cd5b7f5d10975f2d022cc420c3761f5d01 03af061ee5118bcf2835d7a7761608e7d172c93377aa3516d97a6fd350a4ab30f1",
							"hex": "473044022010c56e6f9d6b797d50d98e084483c4b4717916411496a32d9b1a55992f528b4a02206452ce7049a669c83917d77135e60783cd5b7f5d10975f2d022cc420c3761f5d012103af061ee5118bcf2835d7a7761608e7d172c93377aa3516d97a6fd350a4ab30f1"
						},
						"prevOut": {
							"addresses": [
								"1JsdbDL8GBX8aZRqAXEBZaibZiyLpBTJdM"
							],
							"value": 0.1411654
						},
						"sequence": 4294967295
					}
				],
				"vout": [
					{
						"value": 2.5795,
						"n": 0,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 97e222cce73e42c6e3baa643500cffaa9d2090a3 OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a91497e222cce73e42c6e3baa643500cffaa9d2090a388ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"1Er5wAgRyVNx5Ce4KfNXDWFDJa3jQUmEyu"
							]
						}
					},
					{
						"value": 7.5611654,
						"n": 1,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 3224060e14d6cf0d2e225a2a2f3aa8779de4226b OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"15a7wZQhCeQ457KzxRZbeJ8jobb6yMVubR"
							]
						}
					}
				],
				"blockhash": "00000000000000000000af959a7326d0bc24e4266ce70bb5bc90b095d8972c0b",
				"confirmations": 57807,
				"time": 1541182871,
				"blocktime": 1541182871
			},
			{
				"hex": "0100000001077ff719f259e76deba339bcb709ff965dc556be86782f5c1ed36456d4f47fe4010000006b483045022100a31ca601f0bb40aba858e195ad43d76448a7a124f454176f6c927f6b2d1b7147022068f39ac4bea12d5f6c8e6d12b20b9cbbbcefc5e0fd2ffad4f43d6eb01c89cc8d01210330a8a1ab91531b57d4883181f98038dc3bc2a2b4a8cb18dc8e57a09c3d2932bfffffffff022086850b000000001976a914e6ba00ffc9393b821a8d95ea1bace8c613da505288accc228b21000000001976a914223a0a2a0326d738bae6f2648451346f35f9b12388ac00000000",
				"txid": "36b9485a9e0583e467e00a2d7809b2af94153f2871fab2c8925c1013f0e69548",
				"hash": "",
				"size": "",
				"vsize": "",
				"weight": "",
				"version": 1,
				"locktime": 0,
				"vin": [
					{
						"txid": "e47ff4d45664d31e5c2f7886be56c55d96ff09b7bc39a3eb6de759f219f77f07",
						"vout": 1,
						"scriptSig": {
							"asm": "3045022100a31ca601f0bb40aba858e195ad43d76448a7a124f454176f6c927f6b2d1b7147022068f39ac4bea12d5f6c8e6d12b20b9cbbbcefc5e0fd2ffad4f43d6eb01c89cc8d01 0330a8a1ab91531b57d4883181f98038dc3bc2a2b4a8cb18dc8e57a09c3d2932bf",
							"hex": "483045022100a31ca601f0bb40aba858e195ad43d76448a7a124f454176f6c927f6b2d1b7147022068f39ac4bea12d5f6c8e6d12b20b9cbbbcefc5e0fd2ffad4f43d6eb01c89cc8d01210330a8a1ab91531b57d4883181f98038dc3bc2a2b4a8cb18dc8e57a09c3d2932bf"
						},
						"prevOut": {
							"addresses": [
								"15a7wZQhCeQ457KzxRZbeJ8jobb6yMVubR"
							],
							"value": 7.5611654
						},
						"sequence": 4294967295
					}
				],
				"vout": [
					{
						"value": 1.933,
						"n": 0,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 e6ba00ffc9393b821a8d95ea1bace8c613da5052 OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a914e6ba00ffc9393b821a8d95ea1bace8c613da505288ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"1N2yEu8NvmjBKK7612hHK4DcL6Rw4xWHxS"
							]
						}
					},
					{
						"value": 5.6276654,
						"n": 1,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 223a0a2a0326d738bae6f2648451346f35f9b123 OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a914223a0a2a0326d738bae6f2648451346f35f9b12388ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"147yWC7KkMu6rt3Qopvbb19Qncja6EASpT"
							]
						}
					}
				],
				"blockhash": "00000000000000000013f52b83acfc98bfcc9c39a63074ea01d89fa08d481c6a",
				"confirmations": 57742,
				"time": 1541219381,
				"blocktime": 1541219381
			},
			{
				"hex": "01000000021c27b4899c3f563cdfce18ac03b9c50a9164e3cf709b6ed3871c07c7cd93e3a4010000006a47304402202cdcc3f66427f8653630aae1401c52560a32ba11b9ad876963881db9d6f04671022040b7a390fa4e199d0e6b36c20dfb48a815ea7267e730b63cff0d8cee4613c00a0121035c9dee33eb95f7daa64235a10c8558dc663e24ffcd52592c86e35bae54fcfd18ffffffff5d063407277a02aa5bc2df396207116a79ddeca5f2b35d8802a907b559864be3010000006a473044022014e34ded273748d18a5bafec75940d20ce24fb25f4765f18dd665169cb8c8ea802200e1429f62633c6885f126f86eb8625e87fe983914578cbc6440c7344a1717755012103af061ee5118bcf2835d7a7761608e7d172c93377aa3516d97a6fd350a4ab30f1ffffffff0220be82030000000017a914a2db3426a04543907230779fcff0caad105d3cdc87ca056000000000001976a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac00000000",
				"txid": "dc1a9641ca1e77b29327023cb9349ba9c5da698cc604a89b7e435122593f349b",
				"hash": "",
				"size": "",
				"vsize": "",
				"weight": "",
				"version": 1,
				"locktime": 0,
				"vin": [
					{
						"txid": "a4e393cdc7071c87d36e9b70cfe364910ac5b903ac18cedf3c563f9c89b4271c",
						"vout": 1,
						"scriptSig": {
							"asm": "304402202cdcc3f66427f8653630aae1401c52560a32ba11b9ad876963881db9d6f04671022040b7a390fa4e199d0e6b36c20dfb48a815ea7267e730b63cff0d8cee4613c00a01 035c9dee33eb95f7daa64235a10c8558dc663e24ffcd52592c86e35bae54fcfd18",
							"hex": "47304402202cdcc3f66427f8653630aae1401c52560a32ba11b9ad876963881db9d6f04671022040b7a390fa4e199d0e6b36c20dfb48a815ea7267e730b63cff0d8cee4613c00a0121035c9dee33eb95f7daa64235a10c8558dc663e24ffcd52592c86e35bae54fcfd18"
						},
						"prevOut": {
							"addresses": [
								"1N3xKc3vzrmmd3vJiqnQCrxpGhjkVbjA2n"
							],
							"value": 0.65179286
						},
						"sequence": 4294967295
					},
					{
						"txid": "e34b8659b507a902885db3f2a5ecdd796a11076239dfc25baa027a270734065d",
						"vout": 1,
						"scriptSig": {
							"asm": "3044022014e34ded273748d18a5bafec75940d20ce24fb25f4765f18dd665169cb8c8ea802200e1429f62633c6885f126f86eb8625e87fe983914578cbc6440c7344a171775501 03af061ee5118bcf2835d7a7761608e7d172c93377aa3516d97a6fd350a4ab30f1",
							"hex": "473044022014e34ded273748d18a5bafec75940d20ce24fb25f4765f18dd665169cb8c8ea802200e1429f62633c6885f126f86eb8625e87fe983914578cbc6440c7344a1717755012103af061ee5118bcf2835d7a7761608e7d172c93377aa3516d97a6fd350a4ab30f1"
						},
						"prevOut": {
							"addresses": [
								"1JsdbDL8GBX8aZRqAXEBZaibZiyLpBTJdM"
							],
							"value": 0.00063652
						},
						"sequence": 4294967295
					}
				],
				"vout": [
					{
						"value": 0.589,
						"n": 0,
						"scriptPubKey": {
							"asm": "OP_HASH160 a2db3426a04543907230779fcff0caad105d3cdc OP_EQUAL",
							"hex": "a914a2db3426a04543907230779fcff0caad105d3cdc87",
							"reqSigs": 1,
							"type": "scripthash",
							"addresses": [
								"3GY7zAK6EppPVUYsreCm7n4L5Xn75Wix5t"
							]
						}
					},
					{
						"value": 0.06292938,
						"n": 1,
						"scriptPubKey": {
							"asm": "OP_DUP OP_HASH160 3224060e14d6cf0d2e225a2a2f3aa8779de4226b OP_EQUALVERIFY OP_CHECKSIG",
							"hex": "76a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac",
							"reqSigs": 1,
							"type": "pubkeyhash",
							"addresses": [
								"15a7wZQhCeQ457KzxRZbeJ8jobb6yMVubR"
							]
						}
					}
				],
				"blockhash": "0000000000000000001226b0448196207bbcd3051f47ac79b62cd2091a493220",
				"confirmations": 57577,
				"time": 1541316014,
				"blocktime": 1541316014
			}
		]`
	var txHistory []btcd.ResponseSearchRawTransactions
	json.Unmarshal([]byte(rawTxs), &txHistory)
	firstCall := env.btcd.EXPECT().SearchRawTransactions(address, int64(0), int64(2000)).Return(&txHistory, nil).Times(1)
	secondCall := env.mongo.EXPECT().PutUserHistory(gomock.Any()).Return(nil).Times(1).After(firstCall)
	cacheKey := utils.GenCacheKey(address, rs.CommandAll)
	thirdCall := env.redis.EXPECT().Set(cacheKey, gomock.Any(), gomock.Any()).Return(nil).Times(1).After(secondCall)
	env.redis.EXPECT().Del(stateKey).Return(nil).Times(1).After(thirdCall)
}

func TestAccountGetBalance(t *testing.T) {
	v := initVars(t)
	initMocks(&v)

	balance, err := v.account.GetAddressBalance(address)
	if err != nil {
		t.Fail()
		return
	}

	if balance != 0.06292938 {
		t.Fail()
	}
}

func TestAccountGetTransactions(t *testing.T) {
	v := initVars(t)
	initMocks(&v)

	txs, err := v.account.GetAddressTransactions(address)
	if err != nil {
		t.Fail()
		return
	}

	expected := make([]string, 0)
	expected = append(expected,
		"5cf66fd258a5d2be04051134b11bd794a65e149bbbfc1f32e8f18841997e936d",
		"f74918c59110c5389c5b935d01e54428eb33e6180deb90abef22cd8d8100e3ff",
		"e47ff4d45664d31e5c2f7886be56c55d96ff09b7bc39a3eb6de759f219f77f07",
		"36b9485a9e0583e467e00a2d7809b2af94153f2871fab2c8925c1013f0e69548",
		"dc1a9641ca1e77b29327023cb9349ba9c5da698cc604a89b7e435122593f349b")

	if !txsIsEqual(txs, expected) {
		t.Fail()
	}
}

func TestAccountGetUnspents(t *testing.T) {
	v := initVars(t)
	initMocks(&v)

	outputs, err := v.account.GetAddressUnspentOutputs(address)
	if err != nil {
		t.Fail()
		return
	}

	expected := make([]mongo.Unspent, 0)
	expected = append(expected, mongo.Unspent{
		Transaction:  "dc1a9641ca1e77b29327023cb9349ba9c5da698cc604a89b7e435122593f349b",
		VOutIdx:      1,
		ScriptPubKey: "76a9143224060e14d6cf0d2e225a2a2f3aa8779de4226b88ac",
		Amount:       6292938,
		BlockTime:    1541316014,
	})
	if !unspentsIsEqual(outputs, expected) {
		t.Fail()
		return
	}
}
