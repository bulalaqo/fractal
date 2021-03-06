// Copyright 2018 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
)

const (
	// ReceiptStatusFailed is the status code of a action if execution failed.
	ReceiptStatusFailed = uint64(0)

	// ReceiptStatusSuccessful is the status code of a action if execution succeeded.
	ReceiptStatusSuccessful = uint64(1)
)

// ActionResult represents the results the transaction action.
type ActionResult struct {
	Status  uint64
	Index   uint64
	GasUsed uint64
	Error   string
}

// EncodeRLP implements rlp.Encoder
func (a *ActionResult) EncodeRLP() ([]byte, error) {
	return rlp.EncodeToBytes(a)
}

// DecodeRLP implements rlp.Decoder
func (a *ActionResult) DecodeRLP(data []byte) error {
	return rlp.DecodeBytes(data, a)
}

// RPCActionResult that will serialize to the RPC representation of a ActionResult.
type RPCActionResult struct {
	ActionType uint64 `json:"actionType"`
	Status     uint64 `json:"status"`
	Index      uint64 `json:"index"`
	GasUsed    uint64 `json:"gasUsed"`
	Error      string `json:"error"`
}

// NewRPCActionResult returns a ActionResult that will serialize to the RPC.
func (a *ActionResult) NewRPCActionResult(aType ActionType) *RPCActionResult {
	return &RPCActionResult{
		ActionType: uint64(aType),
		Status:     a.Status,
		Index:      a.Index,
		GasUsed:    a.GasUsed,
		Error:      a.Error,
	}
}

// Receipt represents the results of a transaction.
type Receipt struct {
	PostState         []byte
	ActionResults     []*ActionResult
	CumulativeGasUsed uint64
	Bloom             Bloom
	Logs              []*Log
	TxHash            common.Hash
	TotalGasUsed      uint64
}

// NewReceipt creates a barebone transaction receipt, copying the init fields.
func NewReceipt(root []byte, cumulativeGasUsed, totalGasUsed uint64) *Receipt {
	return &Receipt{PostState: common.CopyBytes(root), CumulativeGasUsed: cumulativeGasUsed, TotalGasUsed: totalGasUsed}
}

// EncodeRLP implements rlp.Encoder
func (r *Receipt) EncodeRLP() ([]byte, error) {
	return rlp.EncodeToBytes(r)
}

// DecodeRLP implements rlp.Decoder
func (r *Receipt) DecodeRLP(data []byte) error {
	return rlp.DecodeBytes(data, r)
}

// Size returns the approximate memory used by all internal contents
func (r *Receipt) Size() common.StorageSize {
	bytes, _ := r.EncodeRLP()
	return common.StorageSize(len(bytes))
}

// Hash hashes the RLP encoding of Receipt.
func (r *Receipt) Hash() common.Hash {
	return rlpHash(r)
}

// RPCReceipt that will serialize to the RPC representation of a Receipt.
type RPCReceipt struct {
	BlockHash         common.Hash        `json:"blockHash"`
	BlockNumber       uint64             `json:"blockNumber"`
	Hash              common.Hash        `json:"txHash"`
	TransactionIndex  uint64             `json:"transactionIndex"`
	PostState         hexutil.Bytes      `json:"postState"`
	ActionResults     []*RPCActionResult `json:"actionResults"`
	CumulativeGasUsed uint64             `json:"cumulativeGasUsed"`
	TotalGasUsed      uint64             `json:"totalGasUsed"`
	Bloom             Bloom              `json:"logsBloom"`
	Logs              []*Log             `json:"logs"`
}

// NewRPCReceipt returns a Receipt that will serialize to the RPC.
func (r *Receipt) NewRPCReceipt(blockHash common.Hash, blockNumber uint64, index uint64, tx *Transaction) *RPCReceipt {
	result := &RPCReceipt{
		BlockHash:         blockHash,
		BlockNumber:       blockNumber,
		Hash:              tx.Hash(),
		TransactionIndex:  index,
		PostState:         hexutil.Bytes(r.PostState),
		CumulativeGasUsed: r.CumulativeGasUsed,
		TotalGasUsed:      r.TotalGasUsed,
		Bloom:             r.Bloom,
		Logs:              r.Logs,
	}

	var rpcActionResults []*RPCActionResult
	for i, a := range tx.GetActions() {
		rpcActionResults = append(rpcActionResults, r.ActionResults[i].NewRPCActionResult(a.Type()))
	}
	result.ActionResults = rpcActionResults
	return result
}
