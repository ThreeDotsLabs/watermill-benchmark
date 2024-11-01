package pkg

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type Results struct {
	Count          int
	MessageSize    int
	MeanRate       float64
	MeanThroughput float64
}

// RunBenchmark runs benchmark on chosen pubsub and returns publishing and subscribing results.
func RunBenchmark(pubSubName string, messagesCount int, messageSize int) (Results, Results, error) {
	topic := "benchmark_" + watermill.NewShortUUID()

	if err := initialise(pubSubName, topic); err != nil {
		return Results{}, Results{}, err
	}

	pubsub, err := NewPubSub(pubSubName, topic, messagesCount, messageSize)
	if err != nil {
		return Results{}, Results{}, err
	}

	pubResults, err := pubsub.PublishMessages()
	if err != nil {
		return Results{}, Results{}, err
	}

	var c *Counter

	doneChannel := make(chan struct{})

	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for {
			select {
			case <-doneChannel:
				return
			case <-ticker.C:
				if c != nil {
					fmt.Printf("processed: %d\n", c.count)
				}
			}
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(pubsub.MessagesCount)

	c = NewCounter()

	go func() {
		err := pubsub.ConsumeMessages(&wg, c)
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	mean := c.MeanPerSecond()

	doneChannel <- struct{}{}

	if err := pubsub.Close(); err != nil {
		return Results{}, Results{}, err
	}

	subResults := Results{
		Count:          int(c.Count()),
		MessageSize:    pubsub.MessageSize,
		MeanRate:       mean,
		MeanThroughput: mean * float64(pubsub.MessageSize),
	}

	return pubResults, subResults, nil
}

// It is required to create a subscriber for some PubSubs for initialisation.
func initialise(pubSubName string, topic string) error {
	pubsub, err := NewPubSub(pubSubName, topic, 0, 0)
	if err != nil {
		return err
	}

	if si, ok := pubsub.Subscriber.(message.SubscribeInitializer); ok && strings.Contains(pubSubName, "nats") {
		err = si.SubscribeInitialize(topic)
		if err != nil {
			return err
		}
	}
	if _, err := pubsub.Subscriber.Subscribe(context.Background(), topic); err != nil {
		return err
	}

	err = pubsub.Close()
	if err != nil {
		return err
	}

	return nil
}
