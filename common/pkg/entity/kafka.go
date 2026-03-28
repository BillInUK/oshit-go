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

//type OAuth2VerifiedInfo struct {
//	Provider           string
//	UserID             string
//	UserName           string
//	RequestToken       string
//	RequestTokenSecret string
//	AccessToken        string
//	AccessTokenSecret  string
//	AdditionalData     map[string]string
//}
//
//type KafkaOAuth2VerifiedMsg struct {
//	MsgType    string
//	MsgContent OAuth2VerifiedInfo
//}
//
//func (msg KafkaOAuth2VerifiedMsg) GetMsgType() string {
//	return msg.MsgType
//}
//
//func (msg KafkaOAuth2VerifiedMsg) GetMsgContent() interface{} {
//	return msg.MsgContent
//}
//
//// KafkaBuyTokenMsg BuyToken 消息
//type KafkaBuyTokenMsg struct {
//	MsgType    string          `json:"MsgType"`
//	MsgContent BuyTokenContent `json:"MsgContent"`
//}
//
//// BuyTokenContent BuyToken 消息内容
//type BuyTokenContent struct {
//	Slot        uint64                    `json:"slot"`
//	Transaction quicknode.TransactionData `json:"transaction"`
//}
//
//// 确保实现了 KafkaMsg 接口（如果需要）
//func (msg KafkaBuyTokenMsg) GetMsgType() string {
//	return msg.MsgType
//}
//
//func (msg KafkaBuyTokenMsg) GetMsgContent() interface{} {
//	return msg.MsgContent
//}
