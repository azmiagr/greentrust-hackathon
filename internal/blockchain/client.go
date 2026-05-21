package blockchain

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"greentrust-hackathon/internal/blockchain/greenpassportchain"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type IssuePassportInput struct {
	ProfileID      string
	PassportID     string
	DocumentHashes []string
	GRSScore       float64
	IssuedAt       time.Time
}

type IssuePassportResult struct {
	TxHash      string
	BlockNumber uint64
}

type Client interface {
	IssuePassport(ctx context.Context, input IssuePassportInput) (*IssuePassportResult, error)
}

type EthereumClient struct {
	rpc        *ethclient.Client
	contract   *greenpassportchain.GreenPassportRegistry
	privateKey *ecdsa.PrivateKey
	chainID    *big.Int
	timeout    time.Duration
}

func NewEthereumClientFromEnv() (Client, error) {
	rpcURL := os.Getenv("BLOCKCHAIN_RPC_URL")
	privateKeyHex := os.Getenv("BLOCKCHAIN_PRIVATE_KEY")
	contractAddress := os.Getenv("GREEN_PASSPORT_CONTRACT_ADDRESS")
	chainIDRaw := os.Getenv("BLOCKCHAIN_CHAIN_ID")

	if rpcURL == "" || privateKeyHex == "" || contractAddress == "" || chainIDRaw == "" {
		return nil, errors.New("blockchain env is incomplete")
	}

	rpc, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, err
	}

	chainID, ok := new(big.Int).SetString(chainIDRaw, 10)
	if !ok {
		return nil, errors.New("invalid BLOCKCHAIN_CHAIN_ID")
	}

	contract, err := greenpassportchain.NewGreenPassportRegistry(common.HexToAddress(contractAddress), rpc)
	if err != nil {
		return nil, err
	}

	timeoutSeconds := 90
	if raw := os.Getenv("BLOCKCHAIN_TX_TIMEOUT_SECONDS"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			timeoutSeconds = parsed
		}
	}

	return &EthereumClient{
		rpc:        rpc,
		contract:   contract,
		privateKey: privateKey,
		chainID:    chainID,
		timeout:    time.Duration(timeoutSeconds) * time.Second,
	}, nil
}

func (c *EthereumClient) IssuePassport(ctx context.Context, input IssuePassportInput) (*IssuePassportResult, error) {
	hashes := make([][32]byte, 0, len(input.DocumentHashes))
	for _, raw := range input.DocumentHashes {
		hash, err := sha256HexToBytes32(raw)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, hash)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(c.privateKey, c.chainID)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	auth.Context = ctx

	grsScore := uint16(math.Round(input.GRSScore))
	issuedAt := uint64(input.IssuedAt.Unix())

	tx, err := c.contract.IssuePassport(
		auth,
		input.ProfileID,
		input.PassportID,
		hashes,
		grsScore,
		issuedAt,
	)
	if err != nil {
		return nil, err
	}

	receipt, err := bind.WaitMined(ctx, c.rpc, tx)
	if err != nil {
		return nil, err
	}

	if receipt.Status != 1 {
		return nil, errors.New("blockchain transaction reverted")
	}

	return &IssuePassportResult{
		TxHash:      tx.Hash().Hex(),
		BlockNumber: receipt.BlockNumber.Uint64(),
	}, nil
}

func sha256HexToBytes32(raw string) ([32]byte, error) {
	var out [32]byte

	raw = strings.TrimPrefix(strings.TrimSpace(raw), "0x")
	decoded, err := hex.DecodeString(raw)
	if err != nil {
		return out, err
	}

	if len(decoded) != 32 {
		return out, errors.New("sha256 hash must be 32 bytes")
	}

	copy(out[:], decoded)
	return out, nil
}
