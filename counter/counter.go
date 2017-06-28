package counter

import (
	"encoding/binary"

	"github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	tmlog "github.com/tendermint/tmlibs/log"
)

type CounterApplication struct {
	types.BaseApplication
	logger tmlog.Logger

	hashCount int
	txCount   int
	serial    bool
}

func NewCounterApplication(serial bool, logger tmlog.Logger) *CounterApplication {
	return &CounterApplication{serial: serial, logger: logger}
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

func (app *CounterApplication) DeliverTx(tx []byte) types.Result {
	app.logger.Debug("DeliverTx()", "data", tx)
	if app.serial {
		if len(tx) > 8 {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Max tx size is 8 bytes, got %d", len(tx)))
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(tx):], tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue != uint64(app.txCount) {
			return types.ErrBadNonce.SetLog(cmn.Fmt("Invalid nonce. Expected %v, got %v", app.txCount, txValue))
		}
	}
	app.txCount++
	return types.OK
}

func (app *CounterApplication) CheckTx(tx []byte) types.Result {
	app.logger.Debug("CheckTx()", "data", tx)
	if app.serial {
		if len(tx) > 8 {
			return types.ErrEncodingError.SetLog(cmn.Fmt("Max tx size is 8 bytes, got %d", len(tx)))
		}
		tx8 := make([]byte, 8)
		copy(tx8[len(tx8)-len(tx):], tx)
		txValue := binary.BigEndian.Uint64(tx8)
		if txValue < uint64(app.txCount) {
			return types.ErrBadNonce.SetLog(cmn.Fmt("Invalid nonce. Expected >= %v, got %v", app.txCount, txValue))
		}
	}
	return types.OK
}

func (app *CounterApplication) Commit() types.Result {
	app.logger.Debug("Commit()")
	app.hashCount++
	if app.txCount == 0 {
		return types.OK
	}
	hash := make([]byte, 8)
	binary.BigEndian.PutUint64(hash, uint64(app.txCount))
	return types.NewResultOK(hash, "")
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
