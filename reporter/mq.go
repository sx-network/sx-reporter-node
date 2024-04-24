package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sx-network/sx-reporter/infra/common"
	"github.com/sx-network/sx-reporter/reporter/proto"
)

const (
	// Sets the concurrency level for message queue consumers.
	mqConsumerConcurrency = 1
)

// Holds configuration settings for the message queue.
type MQConfig struct {
	AMQPURI      string       // AMQPURI is the URI for connecting to the AMQP broker.
	ExchangeName string       // ExchangeName is the name of the exchange.
	QueueConfig  *QueueConfig // QueueConfig holds configuration settings for the message queue.
}

// Represents a message queue service.
type MQService struct {
	logger          hclog.Logger // logger is the logger instance.
	config          *MQConfig    // config holds the configuration settings for the message queue.
	connection      Connection   // connection is the connection to the message queue.
	reporterService *ReporterService    // reporterService is the service responsible for reporting.
}

// Represents a connection to the message queue.
type Connection struct {
	Channel *amqp.Channel // Channel is the AMQP channel used for communication.
}

// Holds configuration settings for the message queue queue.
type QueueConfig struct {
	QueueName string // QueueName is the name of the queue.
}

// Initializes a message queue service with a logger, MQ configuration, and reporter service,
// establishes a connection to the AMQP URI, and starts consuming messages from the queue.
// Returns the initialized MQService instance or an error.
func newMQService(logger hclog.Logger, config *MQConfig, reporterService *ReporterService) (*MQService, error) {
	conn, err := getConnection(
		config.AMQPURI,
	)
	if err != nil {
		return nil, err
	}

	mq := &MQService{
		logger:          logger.Named("mq"),
		config:          config,
		connection:      conn,
		reporterService: reporterService,
	}

	go mq.startConsumeLoop()

	return mq, nil
}

// Establishes a connection to RabbitMQ using the provided URL.
// Returns a Connection instance representing the connection and an error if the connection fails.
func getConnection(rabbitMQURL string) (Connection, error) {
	conn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		return Connection{}, err
	}

	ch, err := conn.Channel()

	return Connection{
		Channel: ch,
	}, err
}

// Initiates a loop to continuously listen for messages from the message queue.
// It starts the message consumer with the specified concurrency and handles any errors that occur during consumption.
// If a termination signal is received, it restarts the consumer after a brief delay.
func (mq *MQService) startConsumeLoop() {
	mq.logger.Debug("listening for MQ messages...")

	ctx, _ := context.WithCancel(context.Background())

	reports, errors, err := mq.startConsumer(ctx, mqConsumerConcurrency)

	if err != nil {
		mq.logger.Error("error while starting mq consumer", "err", err)
		panic(err)
	}

	for {
		select {
		case report := <-reports:
			mq.reporterService.queueReportingTx(ProposeOutcome, report.MarketHash, report.Outcome)
		case err = <-errors:
			mq.logger.Error("error while consuming from message queue", "err", err)
			mq.logger.Debug("Restarting consumer...")
			time.Sleep(2 * time.Second)
			reports, errors, err = mq.startConsumer(ctx, mqConsumerConcurrency)
			if err != nil {
				mq.logger.Error("Got Error during consumer restart", err)
			}
		case <-common.GetTerminationSignalCh():
			mq.logger.Debug("got sigterm, shuttown down mq consumer")
			mq.logger.Debug("Restarting consumer...")
			time.Sleep(2 * time.Second)
			reports, errors, err = mq.startConsumer(ctx, mqConsumerConcurrency)
			if err != nil {
				mq.logger.Error("Got Error during consumer restart", err)
			}

		}
	}
}

// Initiates message consumption from the queue, receiving deliveries on the 'deliveries' channel.
// It creates the queue if it doesn't exist, binds it to the exchange, and sets prefetching to optimize concurrency.
// Returns parsed deliveries within the reports channel and any encountered errors within the errors channel.
func (mq *MQService) startConsumer(
	ctx context.Context, concurrency int,
) (<-chan *proto.Report, <-chan error, error) {
	mq.logger.Debug("Starting MQConsumerService...")

	// create the queue if it doesn't already exist
	_, err := mq.connection.Channel.QueueDeclare(mq.config.QueueConfig.QueueName, true, false, false, false, nil)
	if err != nil {
		return nil, nil, err
	}

	// bind the queue to the routing key
	err = mq.connection.Channel.QueueBind(mq.config.QueueConfig.QueueName, "", mq.config.ExchangeName, false, nil)
	if err != nil {
		return nil, nil, err
	}

	// prefetch 4x as many messages as we can handle at once
	prefetchCount := concurrency * 4

	err = mq.connection.Channel.Qos(prefetchCount, 0, false)
	if err != nil {
		return nil, nil, err
	}

	uuid := uuid.New().String()
	deliveries, err := mq.connection.Channel.Consume(
		mq.config.QueueConfig.QueueName, // queue
		uuid,                            // consumer
		false,                           // auto-ack
		false,                           // exclusive
		false,                           // no-local
		false,                           // no-wait
		nil,                             // args
	)

	if err != nil {
		return nil, nil, err
	}

	reports := make(chan *proto.Report)
	errors := make(chan error)

	for i := 0; i < concurrency; i++ {
		go func() {
			for delivery := range deliveries {
				report, err := mq.parseDelivery(delivery)
				if err != nil {
					errors <- err
					//delivery.Nack(false, true) //nolint:errcheck
					// nacking will avoid removing from queue, so we ack even so we've encountered an error
					delivery.Ack(false) //nolint:errcheck
				} else {
					delivery.Ack(false) //nolint:errcheck
					reports <- report
				}
			}
		}()
	}

	// stop the consumer upon sigterm
	go func() {
		<-ctx.Done()
		// stop consumer quickly
		mq.connection.Channel.Cancel(uuid, false) //nolint:errcheck
	}()

	return reports, errors, nil
}

// Unmarshals the delivery body into a report or returns an error if parsing fails.
// It checks for the presence of a message body and handles JSON unmarshaling errors.
func (mq *MQService) parseDelivery(delivery amqp.Delivery) (*proto.Report, error) {
	if delivery.Body == nil {
		return &proto.Report{}, fmt.Errorf("no message body")
	}

	var reportOutcome proto.Report
	if err := json.Unmarshal(delivery.Body, &reportOutcome); err != nil {
		return &proto.Report{}, fmt.Errorf("error during report outcome json unmarshaling, %w", err)
	}

	mq.logger.Debug("MQ message received", "marketHash", reportOutcome.MarketHash)

	return &reportOutcome, nil
}
