package btcd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/junzhli/btcd-address-indexing-worker/logger"
)

type request struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	METHOD  string        `json:"method"`
	PARAMS  []interface{} `json:"params"`
}

// ResponseError illustrates about 'error' in Btcd JSON-RPC response message
type responseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// response illustrates about Btcd JSON-RPC response message
type response struct {
	Result json.RawMessage `json:"result"`
	Error  responseError   `json:"error"`
}

// Btcd is used for queries, in compliance with Btcd JSON-RPC spec
type Btcd interface {
	SearchRawTransactions(addr string, startIdx int64, max int64) (*[]ResponseSearchRawTransactions, error)
	GetInfo() (*map[string]interface{}, error)
}

type btcd struct {
	endpoint  string
	username  string
	password  string
	netClient *http.Client
}

// GetInfo returns some network level information like block height...
//
// {
//	"version": 200000,
//	"protocolversion": 70002,
//	"blocks": 602288,
//	"timeoffset": 0,
//	"connections": 8,
//	"proxy": "",
//	"difficulty": 13691480038694.451,
//	"testnet": false,
//	"relayfee": 0.00001,
//	"errors": ""
// }
//
//
func (b btcd) GetInfo() (*map[string]interface{}, error) {
	payload := request{
		JSONRPC: "1.0",
		ID:      "0",
		METHOD:  "getinfo",
		PARAMS:  []interface{}{},
	}
	pl, err := json.Marshal(payload)
	if err != nil {
		logger.LogOnError(err, "Failed to create payload")
		return nil, err
	}

	res, err := processRequest(&b, pl)
	if res.Error != (responseError{}) {
		logger.LogOnError(err, "Failed to create payload")
		return nil, JSONRPCError{Code: res.Error.Code, Message: res.Error.Message}
	}

	var result map[string]interface{}
	json.Unmarshal([]byte(res.Result), &result)

	return &result, err
}

type scriptPubKey struct {
	Asm       string   `json:"asm"`
	Hex       string   `json:"hex"`
	Type      string   `json:"type"`
	Addresses []string `json:"addresses"`
}

type vout struct {
	Value        float64      `json:"value"`
	ScriptPubKey scriptPubKey `json:"scriptPubKey"`
}

type prevOut struct {
	Addresses []string `json:"addresses"`
}

type vin struct {
	Txid      string  `json:"txid"`
	VoutIndex uint64  `json:"vout"`
	PrevOut   prevOut `json:"prevOut"`
}

type ResponseSearchRawTransactions struct {
	Txid          string `json:"txid"`
	Vins          []vin  `json:"vin"`
	Vouts         []vout `json:"vout"`
	Confirmations uint64 `json:"confirmations"`
	Blocktime     uint64 `json:"blocktime"`
}

// SearchRawTransactions get relevant transactions with given bitcoin address
func (b btcd) SearchRawTransactions(addr string, startIdx int64, max int64) (*[]ResponseSearchRawTransactions, error) {
	// params:
	// 1. address (string, required) - bitcoin address
	// 2. verbose (int, optional, default=true) - specifies the transaction is returned as a JSON object instead of hex-encoded string
	// 3. skip (int, optional, default=0) - the number of leading transactions to leave out of the final response
	// 4. count (int, optional, default=100) - the maximum number of transactions to return
	// 5. vinextra (int, optional, default=0) - Specify that extra data from previous output will be returned in vin
	// 6. reverse (boolean, optional, default=false) - Specifies that the transactions should be returned in reverse chronological order
	payload := request{
		JSONRPC: "1.0",
		ID:      "0",
		METHOD:  "searchrawtransactions",
		PARAMS:  []interface{}{addr, 1, startIdx, max, 1000000000, false},
	}
	pl, err := json.Marshal(payload)
	if err != nil {
		logger.LogOnError(err, "Failed to create payload")
		return nil, err
	}

	res, err := processRequest(&b, pl)
	if res.Error != (responseError{}) {
		if res.Error.Code == -5 {
			return nil, errors.New(ErrorNoDataReturned)
		}
		logger.LogOnError(err, "Failed to create payload")
		return nil, JSONRPCError{Code: res.Error.Code, Message: res.Error.Message}
	}

	var result []ResponseSearchRawTransactions
	if err := json.Unmarshal([]byte(res.Result), &result); err != nil {
		logger.LogOnError(err, "Failed to parse response - phase 1")
		return nil, err
	}

	return &result, err
}

func processRequest(b *btcd, payload []byte) (*response, error) {
	req, err := http.NewRequest("POST", b.endpoint, bytes.NewBuffer(payload))
	if err != nil {
		logger.LogOnError(err, "Failed to initialize request to btcd")
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(b.username, b.password)
	resp, err := b.netClient.Do(req)
	if err != nil {
		logger.LogOnError(err, "Failed to fetch response")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := InvalidResponseCodeError{
			Code: resp.StatusCode,
		}
		logger.LogOnError(err, "Invalid response code")
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.LogOnError(err, "Failed to parse response - phase 1")
		return nil, err
	}

	var result response
	if err := json.Unmarshal(body, &result); err != nil {
		logger.LogOnError(err, "Failed to parse response - phase 2")
		return nil, err
	}

	return &result, nil
}

// New creates a instance of Btcd
func New(ep string, user string, pass string, timeout time.Duration) Btcd {
	// TODO must properly handle connection with self-certificate
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   timeout * time.Second, // TODO consider to taker longer time to prevent timeout
	}
	return btcd{
		endpoint:  ep,
		username:  user,
		password:  pass,
		netClient: client,
	}
}
