package reporter

import (
	"fmt"
	"math/big"

	"github.com/hashicorp/go-hclog"
	"github.com/sx-network/sx-reporter/infra/secrets"
	"github.com/sx-network/sx-reporter/reporter/proto"
)

// Holds configuration options for the reporter service.
type ReporterConfig struct {
	MQConfig                   *MQConfig // Configuration for message queue.
	VerifyOutcomeURI           string    // URI for verifying outcomes.
	OutcomeVotingPeriodSeconds uint64    // Duration of the outcome voting period in seconds.
	OutcomeReporterAddress     string    // Address of the outcome reporter.
	SXNodeAddress              string    // Address of the SX node.
}

// Represents a transaction for reporting.
type ReportingTx struct {
	functionType string        // Type of the function for reporting.
	report       *proto.Report // Report data.
}

// Orchestrates various components of the reporter service.
type ReporterService struct {
	logger                                    hclog.Logger // Logger for reporting.
	secretsManager                            secrets.SecretsManager
	config                                    *ReporterConfig   // Configuration for the reporter.
	mqService                                 *MQService        // AMQP message consumer.
	txService                                 *TxService        // JSON-RPC transaction sender.
	eventListener                             *EventListener    // Listener for blockchain events.
	storeProcessor                            *StoreProcessor   // Processor for market items.
	reportingTxChan                           chan *ReportingTx // Channel for queuing reporting transactions.
	proto.UnimplementedDataFeedOperatorServer                   // DataFeed operator commands implementation.
	// lock                                      sync.Mutex        // Mutex for synchronization.
}

// NewReporterService returns a new instance of the reporter service initialized with the provided parameters.
// It configures and starts various components of the reporter service, including message queue service,
func NewReporterService(
	logger hclog.Logger,
	config *ReporterConfig,
	secretsManager *secrets.SecretsManager,
) (*ReporterService, error) {
	reporterService := &ReporterService{
		logger:          logger.Named("reporter"),
		config:          config,
		reportingTxChan: make(chan *ReportingTx, 100),
		secretsManager:  *secretsManager,
	}

	if config.MQConfig.AMQPURI != "" {
		if config.MQConfig.ExchangeName == "" {
			return nil, fmt.Errorf("reporter 'amqp_uri' provided but missing a valid 'amqp_exchange_name'")
		}

		if config.MQConfig.QueueConfig.QueueName == "" {
			return nil, fmt.Errorf("reporter 'amqp_uri' provided but missing a valid 'amqp_queue_name'")
		}

		mqService, err := newMQService(reporterService.logger, config.MQConfig, reporterService)
		if err != nil {
			return nil, err
		}

		reporterService.mqService = mqService
	}

	txService, err := newTxService(reporterService.logger)
	if err != nil {
		return nil, err
	}
	reporterService.txService = txService

	go reporterService.processTxsFromQueue()

	if config.VerifyOutcomeURI == "" {
		reporterService.logger.Warn("Reporter 'verify_outcome_api_url' is missing but required for outcome voting and reporting.. we will avoid participating in outcome voting and reporting...") //nolint:lll

		return reporterService, nil
	}

	eventListener, err := newEventListener(reporterService.logger, reporterService)
	if err != nil {
		return nil, err
	}
	reporterService.eventListener = eventListener

	storeProcessor, err := newStoreProcessor(reporterService.logger, reporterService)
	if err != nil {
		return nil, err
	}
	reporterService.storeProcessor = storeProcessor

	return reporterService, nil
}

// Queues a reporting transaction for processing with the specified function type, market hash, and outcome.
// It creates a ReportingTx instance with the provided parameters and sets the outcome based on the function type.
// If the function type is "ProposeOutcome", it sets the outcome directly.
// If the function type is "VoteOutcome", it verifies the market and sets the outcome accordingly.
// Finally, it logs a debug message and queues the reporting transaction for processing.
func (d *ReporterService) queueReportingTx(functionType string, marketHash string, outcome int32) {

	reportingTx := &ReportingTx{
		functionType: functionType,
		report: &proto.Report{
			MarketHash: marketHash,
		},
	}

	switch functionType {
	case ProposeOutcome:
		reportingTx.report.Outcome = outcome
	case VoteOutcome:
		verifyOutcome, err := d.verifyMarket(marketHash)
		reportingTx.report.Outcome = verifyOutcome
		if err != nil {
			d.logger.Error("Error encountered in verifying market, skipping vote tx", "err", err)
			return
		}
	default:
		if functionType != ReportOutcome {
			d.logger.Error("Unrecognized function type, skipping tx..", "functionType", functionType)
			return
		}
	}

	d.logger.Debug("queueing reporting tx for processing", "function", functionType, "marketHash", marketHash)
	d.reportingTxChan <- reportingTx
}

// Processes transactions from the reporting transaction channel.
// It continuously listens for reporting transactions from the channel and processes them one by one.
// For each reporting transaction received, it invokes the sendTxWithRetry method to attempt sending
// the transaction with retries in case of failures.
func (d *ReporterService) processTxsFromQueue() {
	for reportingTx := range d.reportingTxChan {
		d.logger.Debug("processing reporting tx", "function", reportingTx.functionType, "marketHash", reportingTx.report.MarketHash)
		d.sendTxWithRetry(reportingTx.functionType, reportingTx.report)
	}
}

// Retrieves the voting period from the chain and updates the configuration accordingly.
// It sends a call to the "_votingPeriod" function on the chain to fetch the voting period value.
// If the result is nil or cannot be converted to *big.Int, it logs an error and returns.
// Otherwise, it updates the OutcomeVotingPeriodSeconds field in the configuration with the retrieved value.
func (d *ReporterService) syncVotingPeriod() {
	result := d.sendCall("_votingPeriod")
	if result == nil {
		d.logger.Error("voting period returned nil")
		return
	}

	votingPeriodOnchain, ok := result.(*big.Int)
	if !ok {
		d.logger.Error("failed to convert result to *big.Int")
		return
	}
	d.logger.Debug("retrieved onchain voting period", votingPeriodOnchain)
	d.logger.Debug("update voting period", votingPeriodOnchain)
	d.config.OutcomeVotingPeriodSeconds = votingPeriodOnchain.Uint64()
}
