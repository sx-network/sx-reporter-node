package reporter

import (
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
)

// Processes market items for reporting.
type StoreProcessor struct {
	logger          hclog.Logger
	reporterService *ReporterService
	store           *MarketItemStore
}

// Represents a store for market items.
type MarketItemStore struct {
	logger      hclog.Logger
	marketItems map[string]uint64
	sync.Mutex
}

// Creates and initializes a new StoreProcessor instance with the provided logger and reporter service.
// It starts the processing loop for handling market items.
func newStoreProcessor(logger hclog.Logger, reporterService *ReporterService) (*StoreProcessor, error) {
	storeProcessor := &StoreProcessor{
		logger:          logger.Named("storeProcessor"),
		reporterService: reporterService,
		store: &MarketItemStore{
			marketItems: make(map[string]uint64),
		},
	}
	storeProcessor.store.logger = storeProcessor.logger.Named("store")

	go storeProcessor.startProcessingLoop()

	return storeProcessor, nil
}

// Starts the processing loop for the StoreProcessor.
// It continuously checks for market items in the store and processes them if they are ready for reporting.
// The loop runs indefinitely with a sleep interval of 5 seconds between iterations.
// For each market item, it compares the stored timestamp plus the outcome voting period with the current time.
// If the item is ready for processing, it logs the processing action and queues a reporting transaction.
// If the item is not yet ready, it logs the remaining time until it becomes ready and continues to the next item.
func (s *StoreProcessor) startProcessingLoop() {
	for {
		time.Sleep(5 * time.Second)
		for marketHash, timestamp := range s.store.marketItems {
			s.logger.Debug("current outcome voting period seconds", s.reporterService.config.OutcomeVotingPeriodSeconds)
			if timestamp+s.reporterService.config.OutcomeVotingPeriodSeconds <= uint64(time.Now().Unix()) {
				s.logger.Debug(
					"processing market item",
					"market", marketHash,
					"block ts", timestamp,
					"current ts", time.Now().Unix())
				s.reporterService.queueReportingTx(ReportOutcome, marketHash, -1)
			} else {
				s.logger.Debug(
					"market item not yet ready for processing",
					"market", marketHash,
					"block ts", timestamp,
					"current ts", time.Now().Unix(),
					"remaining s", timestamp+s.reporterService.config.OutcomeVotingPeriodSeconds-uint64(time.Now().Unix()))
				continue
			}
		}
	}
}

// Adds a new market item to the MarketItemStore.
// It locks the store, adds the market item with its corresponding block timestamp,
// and logs the addition of the item.
func (m *MarketItemStore) add(marketHash string, blockTimestamp uint64) {
	m.Lock()
	defer m.Unlock()

	m.marketItems[marketHash] = blockTimestamp
	m.logger.Debug("added to store", "market", marketHash, "blockTimestamp", blockTimestamp)
}

// Removes a market item from the MarketItemStore.
// It locks the store, deletes the market item with the specified market hash,
// and logs the removal of the item.
func (m *MarketItemStore) remove(marketHash string) {
	m.Lock()
	defer m.Unlock()

	delete(m.marketItems, marketHash)
	m.logger.Debug("removed from store", "market", marketHash)
}
