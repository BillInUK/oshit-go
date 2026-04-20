package entity

import (
	"github.com/gagliardetto/solana-go/rpc"
	"time"
)

type KafkaMsg interface {
	GetMsgType() string
	GetMsgContent() interface{}
}

type NewScannedTx struct {
	Service    string
	SubService string
	TxSig      rpc.TransactionSignature
	DecodedTx  DecodedSolanaTransaction
}

type NewExpiredTx struct {
	Service    string
	SubService string
	TxID       string
}

type KafkaTxMsg struct {
	MsgType    string
	MsgContent NewScannedTx
}

func (msg KafkaTxMsg) GetMsgType() string {
	return msg.MsgType
}

func (msg KafkaTxMsg) GetMsgContent() interface{} {
	return msg.MsgContent
}

type KafkaExpiredTxMsg struct {
	MsgType    string
	MsgContent NewExpiredTx
}

func (msg KafkaExpiredTxMsg) GetMsgType() string {
	return msg.MsgType
}

func (msg KafkaExpiredTxMsg) GetMsgContent() interface{} {
	return msg.MsgContent
}

type KafkaNewHeightMsg struct {
	MsgType    string
	MsgContent uint64
}

func (msg KafkaNewHeightMsg) GetMsgType() string {
	return msg.MsgType
}

func (msg KafkaNewHeightMsg) GetMsgContent() interface{} {
	return msg.MsgContent
}

type KafkaNewSnapShotMsg struct {
	MsgType    string
	MsgContent time.Time
}

func (msg KafkaNewSnapShotMsg) GetMsgType() string {
	return msg.MsgType
}

func (msg KafkaNewSnapShotMsg) GetMsgContent() interface{} {
	return msg.MsgContent
}

type KafkaNewStakeSnapShotMsg struct {
	MsgType    string
	MsgContent time.Time
}

func (msg KafkaNewStakeSnapShotMsg) GetMsgType() string {
	return msg.MsgType
}

func (msg KafkaNewStakeSnapShotMsg) GetMsgContent() interface{} {
	return msg.MsgContent
}
