package reporter

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/contracts/abis"
)

// Represents a listener for Ethereum events. It contains a logger for logging events,
// a reference to the reporter service for processing events, and an Ethereum client for interacting
// with the Ethereum blockchain.
type EventListener struct {
	logger          hclog.Logger
	reporterService *ReporterService
	client          *ethclient.Client
}

// Represents the host address for connecting to an Ethereum JSON-RPC WebSocket server.
const (
	JSONRPCWsHost = "ws://localhost:10002/ws"
)

// Creates a new event listener with the provided logger and reporter service.
// It initializes an EventListener instance with the logger and reporter service, and establishes a connection to the JSON-RPC WebSocket host.
// If an error occurs while dialing the WebSocket RPC URL, it logs the error and returns nil along with the error.
// Otherwise, it starts the event listener's listening loop in a separate goroutine and returns the initialized EventListener instance.
func newEventListener(logger hclog.Logger, reporterService *ReporterService) (*EventListener, error) {

	eventListener := &EventListener{
		logger:          logger.Named("eventListener"),
		reporterService: reporterService,
	}

	client, err := ethclient.Dial(JSONRPCWsHost)
	if err != nil {
		logger.Error("error while dialing ws rpc url", "err", err)
		return nil, err
	}
	eventListener.client = client

	go eventListener.startListeningLoop()

	return eventListener, nil
}

// Initiates the event listener's loop for subscribing to and handling blockchain events.
// It starts by parsing the contract ABI and subscribing to ProposeOutcome and OutcomeReported events.
// Upon receiving events, it unpacks the event data, handles type assertions, and processes the events accordingly.
// If any errors occur during event handling or subscription, it logs the errors and attempts to reconnect after a delay.
func (e EventListener) startListeningLoop() {
	contractAbi, err := abi.JSON(strings.NewReader(abis.OutcomeReporterJSONABI))
	if err != nil {
		e.logger.Error("error while parsing OutcomeReporter contract ABI", "err", err)

		return
	}

	outcomeReporterAddress := common.HexToAddress(e.reporterService.config.OutcomeReporterAddress)

	proposeOutcomeSub, proposeOutcomeLogs, err := e.subscribeToProposeOutcome(contractAbi, outcomeReporterAddress)
	if err != nil {
		panic(fmt.Errorf("fatal error while subscribing to ProposeOutcome logs: %w", err))
	}

	outcomeReportedSub, outcomeReportedLogs, err := e.subscribeToOutcomeReported(contractAbi, outcomeReporterAddress)
	if err != nil {
		panic(fmt.Errorf("fatal error while subscribing to OutcomeReported logs: %w", err))
	}

	e.logger.Debug("listening for events...")

	for {
		select {
		case err := <-proposeOutcomeSub.Err():
			e.logger.Error("error listening to ProposeOutcome events, re-connecting after 5 seconds..", "err", err)
			time.Sleep(5 * time.Second)
			proposeOutcomeSub, proposeOutcomeLogs, err = e.subscribeToProposeOutcome(contractAbi, outcomeReporterAddress)
			if err != nil {
				e.logger.Error("fatal error while re-subscribing to ProposeOutcome logs", "err", err)
			}

		case err := <-outcomeReportedSub.Err():
			e.logger.Error("error listening to OutcomeReported events, re-connecting after 5 seconds..", "err", err)

			time.Sleep(5 * time.Second)
			outcomeReportedSub, outcomeReportedLogs, err = e.subscribeToOutcomeReported(contractAbi, outcomeReporterAddress)
			if err != nil {
				e.logger.Error("fatal error while re-subscribing to OutcomeReported logs", "err", err)
			}

		case vLog := <-proposeOutcomeLogs:
			results, err := contractAbi.Unpack("ProposeOutcome", vLog.Data)

			if err != nil {
				e.logger.Error("error unpacking ProposeOutcome event", "err", err)
			}

			if len(results) == 0 {
				e.logger.Error("unexpected empty results for ProposeOutcome event")
			}

			marketHash, ok := results[0].([32]byte)
			if !ok { // type assertion failed
				e.logger.Error("type assertion failed for [32]byte", "marketHash", results[0], "got type", reflect.TypeOf(results[0]).String())
			}

			outcome, ok := results[1].(uint8)
			if !ok { // type assertion failed
				e.logger.Error("type assertion failed for int", "outcome", results[1], "got type", reflect.TypeOf(results[1]).String())
			}

			blockTimestamp, ok := results[2].(*big.Int)
			if !ok { // type assertion failed
				e.logger.Error("type assertion failed for int", "timestamp", results[2], "got type", reflect.TypeOf(results[2]).String())
			}

			marketHashStr := fmt.Sprintf("0x%s", hex.EncodeToString(marketHash[:]))
			e.logger.Debug("received ProposeOutcome event", "marketHash", marketHashStr, "outcome", outcome, "blockTime", blockTimestamp)

			e.reporterService.syncVotingPeriod()
			e.reporterService.queueReportingTx(VoteOutcome, marketHashStr, -1)
			e.reporterService.storeProcessor.store.add(marketHashStr, uint64(blockTimestamp.Int64()))
		case vLog := <-outcomeReportedLogs:
			results, err := contractAbi.Unpack("OutcomeReported", vLog.Data)

			if err != nil {
				e.logger.Error("error unpacking OutcomeReported event", "err", err)
			}

			if len(results) == 0 {
				e.logger.Error("unexpected empty results for OutcomeReported event")
			}

			marketHash, ok := results[0].([32]byte)
			if !ok { // type assertion failed
				e.logger.Error("type assertion failed for [32]byte", "marketHash", results[0], "got type", reflect.TypeOf(results[0]).String())
			}

			outcome, ok := results[1].(uint8)
			if !ok { // type assertion failed
				e.logger.Error("type assertion failed for int", "outcome", results[1], "got type", reflect.TypeOf(results[1]).String())
			}

			marketHashStr := fmt.Sprintf("0x%s", hex.EncodeToString(marketHash[:]))
			e.logger.Debug("received OutcomeReported event", "marketHash", marketHashStr, "outcome", outcome)

			e.reporterService.storeProcessor.store.remove(marketHashStr)
		}
	}
}

// Subscribes to the ProposeOutcome event from the given contract ABI and outcome reporter address.
// It constructs a filter query based on the event ID and address, then subscribes to logs matching the query.
// It returns a subscription handle, a channel for receiving logs, and any error encountered during subscription.
func (e EventListener) subscribeToProposeOutcome(contractAbi abi.ABI, outcomeReporterAddress common.Address) (ethereum.Subscription, <-chan types.Log, error) {
	proposeOutcomeEvent := contractAbi.Events["ProposeOutcome"].ID

	var proposeOutcomeEventTopics [][]common.Hash

	proposeOutcomeEventTopics = append(proposeOutcomeEventTopics, []common.Hash{proposeOutcomeEvent})

	proposeOutcomeQuery := ethereum.FilterQuery{
		Addresses: []common.Address{outcomeReporterAddress},
		Topics:    proposeOutcomeEventTopics,
	}

	proposeOutcomeLogs := make(chan types.Log)
	proposeOutcomeSub, err := e.client.SubscribeFilterLogs(context.Background(), proposeOutcomeQuery, proposeOutcomeLogs)
	if err != nil {
		e.logger.Error("error in SubscribeFilterLogs call", "err", err)

		return nil, nil, err
	}

	return proposeOutcomeSub, proposeOutcomeLogs, nil
}

// Subscribes to the OutcomeReported event from the given contract ABI and outcome reporter address.
// It constructs a filter query based on the event ID and address, then subscribes to logs matching the query.
// It returns a subscription handle, a channel for receiving logs, and any error encountered during subscription.
func (e EventListener) subscribeToOutcomeReported(contractAbi abi.ABI, outcomeReporterAddress common.Address) (ethereum.Subscription, <-chan types.Log, error) {
	outcomeReportedEvent := contractAbi.Events["OutcomeReported"].ID

	var outcomeReportedEventTopics [][]common.Hash
	outcomeReportedEventTopics = append(outcomeReportedEventTopics, []common.Hash{outcomeReportedEvent})

	outcomeReportedQuery := ethereum.FilterQuery{
		Addresses: []common.Address{outcomeReporterAddress},
		Topics:    outcomeReportedEventTopics,
	}

	outcomeReportedLogs := make(chan types.Log)
	outcomeReportedSub, err := e.client.SubscribeFilterLogs(context.Background(), outcomeReportedQuery, outcomeReportedLogs)
	if err != nil {
		e.logger.Error("error in SubscribeFilterLogs call", "err", err)

		return nil, nil, err
	}

	return outcomeReportedSub, outcomeReportedLogs, nil
}
