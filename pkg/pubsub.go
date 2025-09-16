package pkg

import (
	stdSQL "database/sql"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitemodernc"
	"github.com/ThreeDotsLabs/watermill-sqlite/wmsqlitezombiezen"

	"github.com/Shopify/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v3/pkg/amqp"
	"github.com/ThreeDotsLabs/watermill-googlecloud/pkg/googlecloud"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill-nats/v2/pkg/nats"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill-sql/v4/pkg/sql"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	driver "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	_ "modernc.org/sqlite"
	"zombiezen.com/go/sqlite"
)

const (
	defaultMessagesCount = 1000000
)

var logger = watermill.NopLogger{}

type PubSub struct {
	Publisher  message.Publisher
	Subscriber message.Subscriber

	MessagesCount int
	MessageSize   int

	Topic string

	UUIDFunc func() string
}

func NewPubSub(name string, topic string, messagesCount int, messageSize int) (PubSub, error) {
	definition, ok := pubSubDefinitions[name]
	if !ok {
		return PubSub{}, fmt.Errorf("unknown PubSub: %s", name)
	}

	pub, sub := definition.Constructor()

	if messagesCount == 0 {
		if definition.MessagesCount != 0 {
			messagesCount = definition.MessagesCount
		} else {
			messagesCount = defaultMessagesCount
		}
	}

	return PubSub{
		Publisher:  pub,
		Subscriber: sub,

		MessagesCount: messagesCount,
		MessageSize:   messageSize,
		Topic:         topic,

		UUIDFunc: definition.UUIDFunc,
	}, nil
}

func (ps PubSub) Close() error {
	if err := ps.Publisher.Close(); err != nil {
		return err
	}
	return ps.Subscriber.Close()
}

type PubSubDefinition struct {
	MessagesCount int
	UUIDFunc      func() string
	Constructor   func() (message.Publisher, message.Subscriber)
}

var pubSubDefinitions = map[string]PubSubDefinition{
	"gochannel": {
		Constructor: func() (message.Publisher, message.Subscriber) {
			pubsub := gochannel.NewGoChannel(gochannel.Config{
				Persistent: true,
			}, logger)
			return pubsub, pubsub
		},
	},
	"kafka": {
		MessagesCount: 5000000,
		Constructor:   kafkaConstructor([]string{"kafka:9092"}),
	},
	"kafka-multinode": {
		MessagesCount: 5000000,
		Constructor:   kafkaConstructor([]string{"kafka1:9091", "kafka2:9092", "kafka3:9093", "kafka4:9094", "kafka5:9095"}),
	},
	"nats-jetstream": {
		MessagesCount: 5000000,
		Constructor: func() (message.Publisher, message.Subscriber) {
			natsURL := os.Getenv("WATERMILL_NATS_URL")
			if natsURL == "" {
				natsURL = "nats://nats:4222"
			}

			jsConfig := nats.JetStreamConfig{
				AutoProvision: false,
			}

			ackAsyncString := os.Getenv("WATERMILL_NATS_ACK_ASYNC")
			if strings.EqualFold(ackAsyncString, "true") {
				jsConfig.AckAsync = true
			}

			pub, err := nats.NewPublisher(nats.PublisherConfig{
				URL:       natsURL,
				Marshaler: &nats.NATSMarshaler{},
				JetStream: jsConfig,
			}, logger)
			if err != nil {
				panic(err)
			}

			sub, err := nats.NewSubscriber(nats.SubscriberConfig{
				URL:              natsURL,
				QueueGroupPrefix: "benchmark",
				SubscribersCount: subscribersCount(),
				Unmarshaler:      &nats.NATSMarshaler{},
				JetStream:        jsConfig,
			}, logger)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"googlecloud": {
		Constructor: func() (message.Publisher, message.Subscriber) {
			pub, err := googlecloud.NewPublisher(
				googlecloud.PublisherConfig{
					ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
					Marshaler: googlecloud.DefaultMarshalerUnmarshaler{},
				}, logger,
			)
			if err != nil {
				panic(err)
			}

			sub := NewMultiplier(
				func() (message.Subscriber, error) {
					subscriber, err := googlecloud.NewSubscriber(
						googlecloud.SubscriberConfig{
							ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
							GenerateSubscriptionName: func(topic string) string {
								return topic
							},
							Unmarshaler: googlecloud.DefaultMarshalerUnmarshaler{},
						},
						logger,
					)
					if err != nil {
						return nil, err
					}

					return subscriber, nil
				}, subscribersCount(),
			)

			return pub, sub
		},
	},
	"mysql": {
		MessagesCount: 30000,
		UUIDFunc:      newBinaryULID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			conf := driver.NewConfig()
			conf.Net = "tcp"
			conf.User = "root"
			conf.Addr = "mysql"
			conf.DBName = "watermill"

			db, err := stdSQL.Open("mysql", conf.FormatDSN())
			if err != nil {
				panic(err)
			}

			err = db.Ping()
			if err != nil {
				panic(err)
			}

			pub, err := sql.NewPublisher(
				db,
				sql.PublisherConfig{
					AutoInitializeSchema: true,
					SchemaAdapter:        MySQLSchema{},
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

			sub, err := sql.NewSubscriber(
				db,
				sql.SubscriberConfig{
					SchemaAdapter:    MySQLSchema{},
					OffsetsAdapter:   sql.DefaultMySQLOffsetsAdapter{},
					ConsumerGroup:    watermill.NewULID(),
					InitializeSchema: true,
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"postgresql": {
		MessagesCount: 30000,
		UUIDFunc:      watermill.NewUUID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			dsn := "postgres://watermill:password@postgres:5432/watermill?sslmode=disable"
			db, err := stdSQL.Open("postgres", dsn)
			if err != nil {
				panic(err)
			}

			err = db.Ping()
			if err != nil {
				panic(err)
			}

			pub, err := sql.NewPublisher(
				db,
				sql.PublisherConfig{
					AutoInitializeSchema: true,
					SchemaAdapter:        PostgreSQLSchema{},
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

			sub, err := sql.NewSubscriber(
				db,
				sql.SubscriberConfig{
					SchemaAdapter:    PostgreSQLSchema{},
					OffsetsAdapter:   sql.DefaultPostgreSQLOffsetsAdapter{},
					ConsumerGroup:    watermill.NewULID(),
					InitializeSchema: true,
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"postgresql-queue": {
		MessagesCount: 30000,
		UUIDFunc:      watermill.NewUUID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			dsn := "postgres://watermill:password@postgres:5432/watermill?sslmode=disable"
			db, err := stdSQL.Open("postgres", dsn)
			if err != nil {
				panic(err)
			}

			err = db.Ping()
			if err != nil {
				panic(err)
			}

			pub, err := sql.NewPublisher(
				db,
				sql.PublisherConfig{
					AutoInitializeSchema: true,
					SchemaAdapter: sql.PostgreSQLQueueSchema{
						GeneratePayloadType: func(topic string) string {
							return "BYTEA"
						},
					},
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

			sub, err := sql.NewSubscriber(
				db,
				sql.SubscriberConfig{
					SchemaAdapter: sql.PostgreSQLQueueSchema{
						GeneratePayloadType: func(topic string) string {
							return "BYTEA"
						},
					},
					OffsetsAdapter:   sql.PostgreSQLQueueOffsetsAdapter{},
					InitializeSchema: true,
				},
				logger,
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"amqp": {
		MessagesCount: 100000,
		Constructor: func() (message.Publisher, message.Subscriber) {
			config := amqp.NewDurablePubSubConfig(
				"amqp://rabbitmq:5672",
				func(topic string) string {
					return topic
				},
			)

			pub, err := amqp.NewPublisher(config, logger)
			if err != nil {
				panic(err)
			}

			sub := NewMultiplier(
				func() (message.Subscriber, error) {
					sub, err := amqp.NewSubscriber(config, logger)
					if err != nil {
						panic(err)
					}
					return sub, nil
				}, subscribersCount(),
			)

			return pub, sub
		},
	},
	"redis": {
		Constructor: func() (message.Publisher, message.Subscriber) {
			subClient := redis.NewClient(&redis.Options{
				Addr: "redis:6379",
				DB:   0,
			})
			subscriber, err := redisstream.NewSubscriber(
				redisstream.SubscriberConfig{
					Client:        subClient,
					Unmarshaller:  redisstream.DefaultMarshallerUnmarshaller{},
					ConsumerGroup: "test_consumer_group",
				},
				watermill.NewStdLogger(false, false),
			)
			if err != nil {
				panic(err)
			}

			pubClient := redis.NewClient(&redis.Options{
				Addr: "redis:6379",
				DB:   0,
			})
			publisher, err := redisstream.NewPublisher(
				redisstream.PublisherConfig{
					Client:     pubClient,
					Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
				},
				watermill.NewStdLogger(false, false),
			)
			if err != nil {
				panic(err)
			}

			return publisher, subscriber
		},
	},
	"sqlite-memory-modernc": {
		MessagesCount: 100_000,
		UUIDFunc:      watermill.NewUUID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			db, err := stdSQL.Open("sqlite", ":memory:?journal_mode=WAL&busy_timeout=1000&cache=shared")
			if err != nil {
				panic(err)
			}
			db.SetMaxOpenConns(1) // required

			pub, err := wmsqlitemodernc.NewPublisher(
				db,
				wmsqlitemodernc.PublisherOptions{
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			sub, err := wmsqlitemodernc.NewSubscriber(
				db,
				wmsqlitemodernc.SubscriberOptions{
					BatchSize:        700,
					PollInterval:     time.Millisecond * 10,
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"sqlite-memory-zombiezen": {
		MessagesCount: 100_000,
		UUIDFunc:      watermill.NewUUID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			connectionDSN := "file:benchmark?mode=memory&cache=shared&journal_mode=WAL&busy_timeout=1000"
			conn, err := sqlite.OpenConn(connectionDSN)
			if err != nil {
				panic(err)
			}

			pub, err := wmsqlitezombiezen.NewPublisher(
				conn,
				wmsqlitezombiezen.PublisherOptions{
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			sub, err := wmsqlitezombiezen.NewSubscriber(
				connectionDSN,
				wmsqlitezombiezen.SubscriberOptions{
					BatchSize:        700,
					PollInterval:     time.Millisecond * 10,
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"sqlite-file-modernc": {
		MessagesCount: 100_000,
		UUIDFunc:      watermill.NewUUID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			file, err := os.CreateTemp("", "sqlite-file-modernc")
			if err != nil {
				panic(err)
			}
			defer os.Remove(file.Name())

			db, err := stdSQL.Open("sqlite", file.Name()+"?journal_mode=WAL&busy_timeout=1000&cache=shared")
			if err != nil {
				panic(err)
			}
			db.SetMaxOpenConns(1) // required

			pub, err := wmsqlitemodernc.NewPublisher(
				db,
				wmsqlitemodernc.PublisherOptions{
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			sub, err := wmsqlitemodernc.NewSubscriber(
				db,
				wmsqlitemodernc.SubscriberOptions{
					BatchSize:        700,
					PollInterval:     time.Millisecond * 10,
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
	"sqlite-file-zombiezen": {
		MessagesCount: 100_000,
		UUIDFunc:      watermill.NewUUID,
		Constructor: func() (message.Publisher, message.Subscriber) {
			file, err := os.CreateTemp("", "sqlite-file-zombiezen")
			if err != nil {
				panic(err)
			}
			defer os.Remove(file.Name())

			connectionDSN := file.Name() + "?cache=shared&journal_mode=WAL&busy_timeout=1000"
			conn, err := sqlite.OpenConn(connectionDSN)
			if err != nil {
				panic(err)
			}

			pub, err := wmsqlitezombiezen.NewPublisher(
				conn,
				wmsqlitezombiezen.PublisherOptions{
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			sub, err := wmsqlitezombiezen.NewSubscriber(
				connectionDSN,
				wmsqlitezombiezen.SubscriberOptions{
					BatchSize:        700,
					PollInterval:     time.Millisecond * 10,
					InitializeSchema: true,
					Logger:           logger,
				},
			)
			if err != nil {
				panic(err)
			}

			return pub, sub
		},
	},
}

func subscribersCount() int {
	var (
		mult = 1
		err  error
	)
	if ev := os.Getenv("SUBSCRIBER_CPU_MULTIPLIER"); ev != "" {
		mult, err = strconv.Atoi(ev)
		if err != nil {
			panic(fmt.Sprintf("invalid SUBSCRIBER_CPU_MULTIPLIER: %s", err.Error()))
		}
	}
	return runtime.NumCPU() * mult
}

func kafkaConstructor(brokers []string) func() (message.Publisher, message.Subscriber) {
	return func() (message.Publisher, message.Subscriber) {
		publisher, err := kafka.NewPublisher(
			kafka.PublisherConfig{
				Brokers:   brokers,
				Marshaler: kafka.DefaultMarshaler{},
			},
			logger,
		)
		if err != nil {
			panic(err)
		}

		saramaConfig := kafka.DefaultSaramaSubscriberConfig()
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

		subscriber, err := kafka.NewSubscriber(
			kafka.SubscriberConfig{
				Brokers:               brokers,
				Unmarshaler:           kafka.DefaultMarshaler{},
				OverwriteSaramaConfig: saramaConfig,
				ConsumerGroup:         "benchmark",
			},
			logger,
		)
		if err != nil {
			panic(err)
		}

		return publisher, subscriber
	}
}

type MySQLSchema struct {
	sql.DefaultMySQLSchema
}

func (m MySQLSchema) SchemaInitializingQueries(params sql.SchemaInitializingQueriesParams) ([]sql.Query, error) {
	createMessagesTable := strings.Join([]string{
		"CREATE TABLE IF NOT EXISTS " + m.MessagesTable(params.Topic) + " (",
		"`offset` BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,",
		"`uuid` BINARY(16) NOT NULL,",
		"`created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,",
		"`payload` BLOB DEFAULT NULL,",
		"`metadata` JSON DEFAULT NULL",
		");",
	}, "\n")

	return []sql.Query{{Query: createMessagesTable}}, nil
}

type PostgreSQLSchema struct {
	sql.DefaultPostgreSQLSchema
}

func (p PostgreSQLSchema) SchemaInitializingQueries(params sql.SchemaInitializingQueriesParams) ([]sql.Query, error) {
	createMessagesTable := `
		CREATE TABLE IF NOT EXISTS ` + p.MessagesTable(params.Topic) + ` (
			"offset" BIGSERIAL,
			"uuid" UUID NOT NULL,
			"created_at" TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			"payload" BYTEA DEFAULT NULL,
			"metadata" JSON DEFAULT NULL,
			"transaction_id" xid8 NOT NULL,
			PRIMARY KEY ("transaction_id", "offset")
		);
	`

	return []sql.Query{{Query: createMessagesTable}}, nil
}
