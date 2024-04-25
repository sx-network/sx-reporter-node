package reporter

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"github.com/DataDog/datadog-agent/comp/netflow/common"
	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/helper/types"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/reporter/proto"
	"github.com/umbracle/ethgo"
	ethgoabi "github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/contract"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/wallet"
)

// Constants defining the JSON-RPC host address and smart contract function signatures.
const (
	JSONRPCHost              = "http://localhost:10002"
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
	// Retrieve the private key bytes from the SecretsManager
	privKeyBytes, err := d.secretsManager.GetSecret(keyName)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve private key from SecretsManager: %v", err)
	}

	// Parse the PEM encoded private key
	block, _ := pem.Decode(privKeyBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM encoded private key")
	}

	// Decode the PEM block to obtain the DER encoded private key
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DER encoded private key: %v", err)
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

	privateKey, err := d.GetPrivateKeyFromSecretsManager(secrets.ValidatorKey)
	if err != nil {
		d.txService.logger.Error("private keyy error")
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

	currNonce, err := d.getCurrentNonce(d.config.SXNodeAddress)
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

				nonce, err := d.getCurrentNonce(d.config.SXNodeAddress)
				if err != nil {
					d.txService.logger.Error("nonce error")
				}

				currNonce = common.Max(nonce, nonce+1)
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
			// "from", ethgo.Address(d.consensusInfo().ValidatorAddress),
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

			nonce, err := d.getCurrentNonce(d.config.SXNodeAddress)
			if err != nil {
				d.txService.logger.Error("nonce error")
			}

			currNonce = common.Max(nonce, nonce+1)
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
    txCount, err := d.getTransactionCount(address)
    if err != nil {
        return 0, err
    }

    // Convert txCount to uint64
    nonce := txCount.Uint64()

    return nonce, nil
}
