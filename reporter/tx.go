package reporter

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/helper/types"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/reporter/proto"
	"github.com/umbracle/ethgo"
	ethgoabi "github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
	"golang.org/x/crypto/sha3"
)

// Constants defining the JSON-RPC host address and smart contract function signatures.
const (
	JSONRPCHost              = "https://rpc.sx-rollup-testnet.t.raas.gelato.cloud"
	proposeOutcomeSCFunction = "function proposeOutcome(bytes32 marketHash, uint8 outcome)"
	voteOutcomeSCFunction    = "function voteOutcome(bytes32 marketHash, uint8 outcome)"
	reportOutcomeSCFunction  = "function reportOutcome(bytes32 marketHash)"
)

// Represents a service for interacting with transactions.
type TxService struct {
	logger hclog.Logger
	client *jsonrpc.Client
}

// Constants representing transaction types.
const (
	ProposeOutcome string = "proposeOutcome"
	VoteOutcome    string = "voteOutcome"
	ReportOutcome  string = "reportOutcome"
)

// Initializes a new TxService instance with the provided logger
// and returns it along with any error encountered during initialization.
func newTxService(logger hclog.Logger) (*TxService, error) {

	client, err := jsonrpc.NewClient(JSONRPCHost)
	if err != nil {
		logger.Error("failed to initialize new ethgo client")

		return nil, err
	}

	txService := &TxService{
		logger: logger.Named("tx"),
		client: client,
	}

	return txService, nil
}

func (d *ReporterService) GetPrivateKeyFromSecretsManager(keyName string) (*ecdsa.PrivateKey, error) {
	// // Convert byte slice to hex-encoded string
	// privateKeyString := hex.EncodeToString(privKeyBytes)

	// Retrieve the private key bytes from the SecretsManager
	privKeyBytes, err := d.secretsManager.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve private key from SecretsManager: %v", err)
	}

	privKey, err := BytesToECDSAPrivateKey(privKeyBytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

// Sends a transaction to the blockchain with retry logic
// in case of failures. It constructs the transaction based on the provided
// function type and report data. The transaction is attempted multiple times
// with increasing gas price and nonce until it succeeds or reaches the maximum
// number of tries. If the transaction fails due to a low nonce error, it retries
// with a higher nonce.
func (d *ReporterService) sendTxWithRetry(
	functionType string,
	report *proto.Report,
) {

	d.logger.Debug("functionType", functionType)
	d.logger.Debug("validator key", secrets.ReporterKey)

	privateKey, err := d.GetPrivateKeyFromSecretsManager(secrets.ReporterKey)
	if err != nil {
		d.txService.logger.Error("private key error")
	}

	const (
		maxTxTries    = 4
		txGasPriceWei = 1000000000
		txGasLimitWei = 1000000
	)

	var functionSig string

	var functionName string

	var functionArgs []interface{}

	switch functionType {
	case ProposeOutcome:
		functionSig = proposeOutcomeSCFunction
		functionName = ProposeOutcome

		functionArgs = append(make([]interface{}, 0), types.StringToHash(report.MarketHash), report.Outcome)
	case VoteOutcome:
		functionSig = voteOutcomeSCFunction
		functionName = VoteOutcome

		functionArgs = append(make([]interface{}, 0), types.StringToHash(report.MarketHash), report.Outcome)
	case ReportOutcome:
		functionSig = reportOutcomeSCFunction
		functionName = ReportOutcome

		functionArgs = append(make([]interface{}, 0), types.StringToHash(report.MarketHash))
	}

	abiContract, err := ethgoabi.NewABIFromList([]string{functionSig})
	if err != nil {
		d.txService.logger.Error(
			"failed to retrieve ethgo ABI",
			"function", functionName,
			"err", err,
		)

		return
	}

	validatorAddress, err := GetValidatorAddressFromSecretManager(d.secretsManager)
	if err != nil {
		return
	}

	d.logger.Debug("validatorAddress", validatorAddress.String())

	c := contract.NewContract(
		ethgo.Address(ethgo.HexToAddress(d.config.SXNodeAddress)),
		abiContract,
		contract.WithSender(wallet.NewKey(privateKey)),
		contract.WithJsonRPC(d.txService.client.Eth()),
	)

	txn, err := c.Txn(
		functionName,
		functionArgs...,
	)

	if err != nil {
		d.txService.logger.Error(
			"failed to create txn via ethgo",
			"function", functionName,
			"functionArgs", functionArgs,
			"functionSig", abiContract,
			"err", err,
		)

		return
	}

	txTry := uint64(0)
	// currNonce := d.consensusInfo().Nonce

	currNonce, err := d.getCurrentNonce(validatorAddress.String())
	if err != nil {
		d.txService.logger.Error("nonce error")
	}

	for txTry < maxTxTries {

		d.txService.logger.Debug(
			"attempting tx with nonce",
			"function", functionName,
			"nonce", currNonce,
			"try #", txTry,
			"marketHash", report.MarketHash)

		//TODO: derive these gas params better, have it dynamic?
		txn.WithOpts(
			&contract.TxnOpts{
				GasPrice: txGasPriceWei + (txTry * txGasPriceWei),
				GasLimit: txGasLimitWei,
				Nonce:    currNonce,
			},
		)

		// TODO: consider adding directly to txpool txpool.AddTx() instead of over local jsonrpc
		// can use TxnPoolOperatorClient.AddTx()
		err = txn.Do()
		if err != nil {
			if strings.Contains(err.Error(), "nonce too low") {
				// if nonce too low, retry with higher nonce
				d.txService.logger.Debug(
					"encountered nonce too low error trying to send raw txn via ethgo, retrying...",
					"function", functionName,
					"try #", txTry,
					"nonce", currNonce,
					"marketHash", report.MarketHash,
				)

				nonce, err := d.getCurrentNonce(validatorAddress.String())
				if err != nil {
					d.txService.logger.Error("nonce error")
				}

				currNonce = Max(nonce, nonce+1)
				txTry++

				continue
			} else {
				// if any other error, just log and return for now
				d.txService.logger.Error(
					"failed to send raw txn via ethgo due to non-recoverable error",
					"function", functionName,
					"err", err,
					"try #", txTry,
					"nonce", currNonce,
					"marketHash", report.MarketHash,
				)

				return
			}
		}

		d.txService.logger.Debug(
			"sent tx",
			"function", functionName,
			"hash", txn.Hash(),
			"from", ethgo.Address(validatorAddress),
			"nonce", currNonce,
			"market", report.MarketHash,
			"outcome", report.Outcome,
		)

		// wait for tx to mine
		receipt := <-d.txService.waitTxConfirmed(txn.Hash())

		if receipt.Status == 1 {
			d.txService.logger.Debug(
				"got success receipt",
				"function", functionName,
				"nonce", currNonce,
				"txHash", txn.Hash(),
				"marketHash", report.MarketHash,
			)

			if functionName == ReportOutcome {
				d.storeProcessor.store.remove(report.MarketHash)
			}

			return
		} else {

			nonce, err := d.getCurrentNonce(validatorAddress.String())
			if err != nil {
				d.txService.logger.Error("nonce error")
			}

			currNonce = Max(nonce, nonce+1)
			d.txService.logger.Debug(
				"got failed receipt, retrying with nextNonce and more gas",
				"function", functionName,
				"try #", txTry,
				"nonce", currNonce,
				"txHash", txn.Hash(),
				"marketHash", report.MarketHash,
			)
			txTry++
		}
	}
	d.txService.logger.Debug("could not get success tx receipt even after max tx retries",
		"function", functionName,
		"try #", txTry,
		"nonce", currNonce,
		"txHash", txn.Hash(),
		"marketHash", report.MarketHash)

	if functionName == ReportOutcome {
		d.storeProcessor.store.remove(report.MarketHash)
	}
}

// Returns a channel that receives the receipt of a transaction
// once it has been confirmed on the blockchain. It continuously polls the
// blockchain for the transaction receipt using its hash until the receipt is
// available. Once the receipt is received, it is sent over the channel.
func (t *TxService) waitTxConfirmed(hash ethgo.Hash) <-chan *ethgo.Receipt {
	ch := make(chan *ethgo.Receipt)
	go func() {
		for {
			var receipt *ethgo.Receipt
			t.client.Call("eth_getTransactionReceipt", &receipt, hash)
			if receipt != nil {
				ch <- receipt
			}

			time.Sleep(time.Millisecond * 500)
		}
	}()

	return ch
}

func (d *ReporterService) getCurrentNonce(address string) (uint64, error) {
	// Get the transaction count for the given address
	txCount, err := d.GetTransactionCount(address)
	if err != nil {
		return 0, err
	}

	// Convert txCount to uint64
	nonce := txCount.Uint64()

	return nonce, nil
}

func (d *ReporterService) GetTransactionCount(address string) (*big.Int, error) {
	// Initialize JSON-RPC client
	client, err := rpc.DialContext(context.Background(), "https://rpc.toronto.sx.technology")
	if err != nil {
		d.txService.logger.Error("failed to connect to Ethereum node", "err", err)
		return nil, err
	}
	defer client.Close()

	// Convert address string to Ethereum common.Address
	addr := common.HexToAddress(address)

	// Call eth_getTransactionCount method
	var count hexutil.Uint64
	err = client.CallContext(context.Background(), &count, "eth_getTransactionCount", addr, "latest")
	if err != nil {
		d.txService.logger.Error("failed to call eth_getTransactionCount via JSON-RPC", "err", err)
		return nil, err
	}

	// Convert the result to big.Int
	txCount := new(big.Int).SetUint64(uint64(count))

	d.logger.Debug("txCount", txCount)

	return txCount, nil
}

func Max(a, b uint64) uint64 {
	if a > b {
		return a
	}

	return b
}

func BytesToECDSAPrivateKey(input []byte) (*ecdsa.PrivateKey, error) {
	// The key file on disk should be encoded in Base64,
	// so it must be decoded before it can be parsed by ParsePrivateKey
	decoded, err := hex.DecodeString(string(input))
	if err != nil {
		return nil, err
	}

	// Make sure the key is properly formatted
	if len(decoded) != 32 {
		// Key must be exactly 64 chars (32B) long
		return nil, fmt.Errorf("invalid key length (%dB), should be 32B", len(decoded))
	}

	// Convert decoded bytes to a private key
	key, err := ParseECDSAPrivateKey(decoded)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// S256 is the secp256k1 elliptic curve
var S256 = btcec.S256()

func ParseECDSAPrivateKey(buf []byte) (*ecdsa.PrivateKey, error) {
	prv, _ := btcec.PrivKeyFromBytes(S256, buf)

	return prv.ToECDSA(), nil
}

var ErrECDSAKeyNotFound = errors.New("ECDSA key not found in given path")

func GetValidatorAddressFromSecretManager(manager secrets.SecretsManager) (types.Address, error) {
	if !manager.HasSecret(secrets.ReporterKey) {
		return types.ZeroAddress, ErrECDSAKeyNotFound
	}

	keyBytes, err := manager.GetSecret(secrets.ReporterKey)
	if err != nil {
		return types.ZeroAddress, err
	}

	privKey, err := BytesToECDSAPrivateKey(keyBytes)
	if err != nil {
		return types.ZeroAddress, err
	}

	return PubKeyToAddress(&privKey.PublicKey), nil
}

// PubKeyToAddress returns the Ethereum address of a public key
func PubKeyToAddress(pub *ecdsa.PublicKey) types.Address {
	buf := Keccak256(MarshalPublicKey(pub)[1:])[12:]

	return types.BytesToAddress(buf)
}

// Keccak256 calculates the Keccak256
func Keccak256(v ...[]byte) []byte {
	h := sha3.NewLegacyKeccak256()
	for _, i := range v {
		h.Write(i)
	}

	return h.Sum(nil)
}

// MarshalPublicKey marshals a public key on the secp256k1 elliptic curve.
func MarshalPublicKey(pub *ecdsa.PublicKey) []byte {
	return elliptic.Marshal(S256, pub.X, pub.Y)
}
