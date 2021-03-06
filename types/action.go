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
	"errors"
	"io"
	"math/big"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/fractalplatform/fractal/common"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// ErrInvalidSig invalid signature.
var ErrInvalidSig = errors.New("invalid action v, r, s values")

// ActionType type of Action.
type ActionType uint64

const (
	// Transfer represents the ordinary and contract transfer action.
	Transfer ActionType = iota
	// CreateContract repesents the create contract action.
	CreateContract
	// CreateAccount repesents the create account.
	CreateAccount
	UpdateAccount
	DeleteAccount
	// Asset operation
	IncreaseAsset
	IssueAsset
	SetAssetOwner
	// dpos
	Miner
	RegProducer
	UpdateProducer
	UnregProducer
	RemoveVoter
	VoteProducer
	ChangeProducer
	UnvoteProducer
)

type actionData struct {
	AType    ActionType
	Nonce    uint64
	AssetID  uint64
	From     common.Name
	To       common.Name
	GasLimit uint64
	Amount   *big.Int
	Payload  []byte

	// Signature values
	V *big.Int
	R *big.Int
	S *big.Int
}

// Action represents an entire action in the transaction.
type Action struct {
	data actionData
	// cache
	hash   atomic.Value
	sender atomic.Value
}

// NewAction initialize transaction's action.
func NewAction(actionType ActionType, from, to common.Name, nonce, assetID, gasLimit uint64, amount *big.Int, payload []byte) *Action {
	if len(payload) > 0 {
		payload = common.CopyBytes(payload)
	}
	data := actionData{
		AType:    actionType,
		Nonce:    nonce,
		AssetID:  assetID,
		From:     from,
		To:       to,
		GasLimit: gasLimit,
		Amount:   new(big.Int),
		Payload:  payload,
		V:        new(big.Int),
		R:        new(big.Int),
		S:        new(big.Int),
	}
	if amount != nil {
		data.Amount.Set(amount)
	}
	return &Action{data: data}
}

func (a *Action) Type() ActionType       { return a.data.AType }
func (a *Action) Nonce() uint64          { return a.data.Nonce }
func (a *Action) AssetID() uint64        { return a.data.AssetID }
func (a *Action) Sender() common.Name    { return a.data.From }
func (a *Action) Recipient() common.Name { return a.data.To }
func (a *Action) Data() []byte           { return common.CopyBytes(a.data.Payload) }
func (a *Action) Gas() uint64            { return a.data.GasLimit }
func (a *Action) Value() *big.Int        { return new(big.Int).Set(a.data.Amount) }

// EncodeRLP implements rlp.Encoder
func (a *Action) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &a.data)
}

// DecodeRLP implements rlp.Decoder
func (a *Action) DecodeRLP(s *rlp.Stream) error {
	return s.Decode(&a.data)
}

// ChainID returns which chain id this action was signed for (if at all)
func (a *Action) ChainID() *big.Int {
	return deriveChainID(a.data.V)
}

// Hash hashes the RLP encoding of action.
func (a *Action) Hash() common.Hash {
	if hash := a.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(a)
	a.hash.Store(v)
	return v
}

// WithSignature returns a new transaction with the given signature.
func (a *Action) WithSignature(signer Signer, sig []byte) error {
	r, s, v, err := signer.SignatureValues(sig)
	if err != nil {
		return err
	}
	a.data.R, a.data.S, a.data.V = r, s, v
	return nil
}

// RawSignatureValues return raw signature values.
func (a *Action) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return a.data.V, a.data.R, a.data.S
}

// RPCAction represents a action that will serialize to the RPC representation of a action.
type RPCAction struct {
	Type       uint64        `json:"type"`
	Nonce      uint64        `json:"nonce"`
	From       common.Name   `json:"from"`
	To         common.Name   `json:"to"`
	AssetID    uint64        `json:"assetID"`
	GasLimit   uint64        `json:"gas"`
	Amount     *big.Int      `json:"value"`
	Payload    hexutil.Bytes `json:"payload"`
	V          *hexutil.Big  `json:"v"`
	R          *hexutil.Big  `json:"r"`
	S          *hexutil.Big  `json:"s"`
	Hash       common.Hash   `json:"actionHash"`
	ActionIdex uint64        `json:"actionIndex"`
}

// NewRPCAction returns a action that will serialize to the RPC.
func (a *Action) NewRPCAction(index uint64) *RPCAction {
	v, r, s := a.RawSignatureValues()
	result := &RPCAction{
		Type:       uint64(a.Type()),
		Nonce:      a.Nonce(),
		From:       a.Sender(),
		To:         a.Recipient(),
		AssetID:    a.AssetID(),
		GasLimit:   a.Gas(),
		Amount:     a.Value(),
		Payload:    hexutil.Bytes(a.Data()),
		Hash:       a.Hash(),
		V:          (*hexutil.Big)(v),
		R:          (*hexutil.Big)(r),
		S:          (*hexutil.Big)(s),
		ActionIdex: index,
	}
	return result
}

// deriveChainID derives the chain id from the given v parameter
func deriveChainID(v *big.Int) *big.Int {
	v = new(big.Int).Sub(v, big.NewInt(35))
	return v.Div(v, big.NewInt(2))
}
