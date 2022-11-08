package types

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const (
	LegacyTxType     = "legacy"
	AccessListTxType = "access_list"
	DynamicFeeTxType = "dynamic_fee"
)

type CreateEthAccountRequest struct {
	KeyID string            `json:"keyId,omitempty" example:"my-key-account"`
	Tags  map[string]string `json:"tags,omitempty"`
}

type ImportEthAccountRequest struct {
	KeyID      string            `json:"keyId,omitempty" example:"my-imported-key-account"`
	PrivateKey hexutil.Bytes     `json:"privateKey" validate:"required" example:"0x56202652FDFFD802B7252A456DBD8F3ECC0352BBDE76C23B40AFE8AEBD714E2E" swaggertype:"string"`
	Tags       map[string]string `json:"tags,omitempty"`
}

type UpdateEthAccountRequest struct {
	Tags map[string]string `json:"tags,omitempty"`
}

type SignMessageRequest struct {
	Message hexutil.Bytes `json:"message" validate:"required" example:"0xfeade..." swaggertype:"string"`
}

type SignTypedDataHashRequest struct {
	Message hexutil.Bytes `json:"hash" validate:"required" example:"0xfeade..." swaggertype:"string"`
}

type SignTypedDataRequest struct {
	DomainSeparator DomainSeparator        `json:"domainSeparator" validate:"required"`
	Types           map[string][]Type      `json:"types" validate:"required"`
	Message         map[string]interface{} `json:"message" validate:"required"`
	MessageType     string                 `json:"messageType" validate:"required" example:"Mail"`
}

type DomainSeparator struct {
	Name              string `json:"name" validate:"required" example:"MyDApp"`
	Version           string `json:"version" validate:"required" example:"v1.0.0"`
	ChainID           int64  `json:"chainID" validate:"required" example:"1"`
	VerifyingContract string `json:"verifyingContract,omitempty" validate:"omitempty,isHexAddress" example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18"`
	Salt              string `json:"salt,omitempty" validate:"omitempty" example:"some-random-string"`
}

type Type struct {
	Name string `json:"name" validate:"required" example:"fieldName"`
	Type string `json:"type" validate:"required" example:"string"`
}

type SignETHTransactionRequest struct {
	TransactionType string           `json:"transactionType,omitempty" example:"dynamic_fee" enums:"legacy,access_list,dynamic_fee"`
	Nonce           hexutil.Uint64   `json:"nonce" example:"0x1" swaggertype:"string"`
	To              *common.Address  `json:"to,omitempty" example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18" swaggertype:"string"`
	Value           hexutil.Big      `json:"value,omitempty" example:"0xfeaeae" swaggertype:"string"`
	GasPrice        hexutil.Big      `json:"gasPrice,omitempty" example:"0x0" swaggertype:"string"`
	GasLimit        hexutil.Uint64   `json:"gasLimit" validate:"required" example:"0x5208" swaggertype:"string"`
	Data            hexutil.Bytes    `json:"data,omitempty" example:"0xfeaeee..." swaggertype:"string"`
	ChainID         hexutil.Big      `json:"chainID" validate:"required" example:"0x1 (mainnet)" swaggertype:"string"`
	GasFeeCap       *hexutil.Big     `json:"maxFeePerGas,omitempty" example:"0x5208" swaggertype:"string"`
	GasTipCap       *hexutil.Big     `json:"maxPriorityFeePerGas,omitempty" example:"0x5208" swaggertype:"string"`
	AccessList      types.AccessList `json:"accessList,omitempty" swaggertype:"array,object"`
}

type SignQuorumPrivateTransactionRequest struct {
	Nonce    hexutil.Uint64  `json:"nonce" example:"0x1" swaggertype:"string"`
	To       *common.Address `json:"to,omitempty" example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18" swaggertype:"string"`
	Value    hexutil.Big     `json:"value,omitempty" example:"0x1" swaggertype:"string"`
	GasPrice hexutil.Big     `json:"gasPrice" validate:"required" example:"0x0" swaggertype:"string"`
	GasLimit hexutil.Uint64  `json:"gasLimit" validate:"required" example:"0x5208" swaggertype:"string"`
	Data     hexutil.Bytes   `json:"data,omitempty" example:"0xfeaeee..." swaggertype:"string"`
}

type SignEEATransactionRequest struct {
	Nonce          hexutil.Uint64  `json:"nonce" example:"0x1" swaggertype:"string"`
	To             *common.Address `json:"to,omitempty" example:"0x905B88EFf8Bda1543d4d6f4aA05afef143D27E18" swaggertype:"string"`
	Value          hexutil.Big     `json:"value,omitempty" example:"0x1" swaggertype:"string"`
	GasPrice       hexutil.Big     `json:"gasPrice,omitempty" example:"0x0" swaggertype:"string"`
	GasLimit       hexutil.Uint64  `json:"gasLimit,omitempty" example:"0x5208" swaggertype:"string"`
	Data           hexutil.Bytes   `json:"data,omitempty" example:"0xfeaeee..." swaggertype:"string"`
	ChainID        hexutil.Big     `json:"chainID" validate:"required" example:"0x1 (mainnet)" swaggertype:"string"`
	PrivateFrom    string          `json:"privateFrom" validate:"required,base64,required_with=PrivateFor PrivacyGroupID" example:"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="`
	PrivateFor     []string        `json:"privateFor,omitempty" validate:"omitempty,min=1,unique,dive,base64" example:"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo=,B1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="`
	PrivacyGroupID string          `json:"privacyGroupId,omitempty" validate:"omitempty,base64" example:"A1aVtMxLCUHmBVHXoZzzBgPbW/wj5axDpW9X8l91SGo="`
}

type EthAccountResponse struct {
	PublicKey           hexutil.Bytes     `json:"publicKey" example:"0x1abae27a0cbfb02945720425d3b80c7e09728534" swaggertype:"string"`
	CompressedPublicKey hexutil.Bytes     `json:"compressedPublicKey" example:"0x6019a3c8..." swaggertype:"string"`
	CreatedAt           time.Time         `json:"createdAt" example:"2020-07-09T12:35:42.115395Z"`
	UpdatedAt           time.Time         `json:"updatedAt" example:"2020-07-09T12:35:42.115395Z"`
	DeletedAt           *time.Time        `json:"deletedAt,omitempty" example:"2020-07-09T12:35:42.115395Z"`
	KeyID               string            `json:"keyId" example:"my-key-id"`
	Tags                map[string]string `json:"tags,omitempty"`
	Address             common.Address    `json:"address" example:"0x664895b5fE3ddf049d2Fb508cfA03923859763C6" swaggertype:"string"`
	Disabled            bool              `json:"disabled" example:"false"`
}
