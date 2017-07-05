package abcicounter

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

const RewardPerTx uint64 = 5.0

type CounterApplication struct {
	types.BaseApplication
	logger tmlog.Logger

	hashCount int
	txCount   int
	serial    bool

	rewards map[string]uint64
	UID     string
}

func NewCounterApplication(serial bool, logger tmlog.Logger, UID string) *CounterApplication {

	return &CounterApplication{
		serial:  serial,
		logger:  logger,
		txCount: 0,
		rewards: make(map[string]uint64),
		UID:     UID,
	}
}

func (app *CounterApplication) Info() types.ResponseInfo {
	app.logger.Debug("Info()")
	return types.ResponseInfo{Data: cmn.Fmt("{\"hashes\":%v,\"txs\":%v}", app.hashCount, app.txCount)}
}

func (app *CounterApplication) SetOption(key string, value string) (log string) {
	if key == "serial" && value == "on" {
		app.serial = true
	}
	return ""
}

func (app *CounterApplication) EndBlock(height uint64) (resEndBlock types.ResponseEndBlock) {
	//validators set managed by multiverse :)
	res := types.ResponseEndBlock{Diffs: nil}
	return res
}

func (app *CounterApplication) DeliverTx(data []byte) types.Result {
	app.logger.Debug("DeliverTx()", "data", data)
	tx, res, err := app.validateTx(data)
	if err != nil {
		return res
	}
	app.txCount++
	if r, ok := app.rewards[tx.RewardDest]; ok == true {
		app.rewards[tx.RewardDest] = r + RewardPerTx
	} else {
		app.rewards[tx.RewardDest] = RewardPerTx
	}

	return types.OK
}

type Tx struct {
	Amount     uint64 `json:"amount"`
	RewardDest string `json:"reward_destination"`
}

func decodeTx(data []byte) (Tx, error) {
	res := Tx{}
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, err := base64.StdEncoding.Decode(decoded, data)
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(decoded[:n], &res)
	return res, err
}

func (app *CounterApplication) validateTx(data []byte) (Tx, types.Result, error) {
	tx, err := decodeTx(data)
	if err != nil {
		return tx, types.ErrEncodingError.SetLog(cmn.Fmt("decondig tx: %s", err)), err
	}
	if app.serial && tx.Amount < uint64(app.txCount) {
		return tx, types.ErrBadNonce.SetLog(cmn.Fmt("Invalid nonce. Expected >= %v, got %v", app.txCount, tx.Amount)), fmt.Errorf("Bad nonce")
	}
	return tx, types.OK, nil
}

func (app *CounterApplication) CheckTx(data []byte) types.Result {
	app.logger.Debug("CheckTx()", "data", data)
	tx, res, err := app.validateTx(data)
	if err != nil {
		return res
	}
	if tx.RewardDest != app.UID {
		return types.ErrBaseInvalidPubKey.SetLog(cmn.Fmt("Invalid pubkey (got:%s, want:%s). Reward me please!", tx.RewardDest, app.UID))
	}
	return res
}

func (app *CounterApplication) Commit() types.Result {
	app.logger.Debug("Commit()")
	app.hashCount++
	app.logger.Debug("Hash count", "count", app.hashCount)
	if app.txCount == 0 {
		return types.OK
	}
	//we built the sha256 hash of the rewards tree and the count

	hash := sha256.New()
	byteCount := make([]byte, 8)
	binary.BigEndian.PutUint64(byteCount, uint64(app.txCount))
	hash.Write(byteCount)
	fmt.Fprintf(hash, "%v", app.rewards)
	return types.NewResultOK(hash.Sum(nil), "")
}

func (app *CounterApplication) Query(reqQuery types.RequestQuery) types.ResponseQuery {
	app.logger.Debug("Query()", "query", reqQuery)
	switch reqQuery.Path {
	case "hash":
		return types.ResponseQuery{Value: []byte(cmn.Fmt("%v", app.hashCount))}
	case "tx":
		return types.ResponseQuery{Value: []byte(cmn.Fmt("%v", app.txCount))}
	default:
		return types.ResponseQuery{Log: cmn.Fmt("Invalid query path. Expected hash or tx, got %v", reqQuery.Path)}
	}
}
