package pkg

import (
	cryptoRand "crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/oklog/ulid"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

func (ps PubSub) PublishMessages() (Results, error) {
	messagesLeft := ps.MessagesCount
	workers := 64

	wg := sync.WaitGroup{}
	wg.Add(workers)

	addMsg := make(chan *message.Message)

	for num := 0; num < workers; num++ {
		go func() {
			defer wg.Done()

			for msg := range addMsg {
				if err := ps.Publisher.Publish(ps.Topic, msg); err != nil {
					panic(err)
				}
			}
		}()
	}

	msgPayload, err := ps.payload()
	if err != nil {
		return Results{}, err
	}

	start := time.Now()

	var uuidFunc func() string
	if ps.UUIDFunc != nil {
		uuidFunc = ps.UUIDFunc
	} else {
		uuidFunc = watermill.NewULID
	}

	for ; messagesLeft > 0; messagesLeft-- {
		msg := message.NewMessage(uuidFunc(), msgPayload)
		addMsg <- msg
	}
	close(addMsg)

	wg.Wait()

	elapsed := time.Now().Sub(start)

	fmt.Printf("added %d messages in %s, %f msg/s\n", ps.MessagesCount, elapsed, float64(ps.MessagesCount)/elapsed.Seconds())

	return Results{
		Count:          ps.MessagesCount,
		MessageSize:    ps.MessageSize,
		MeanRate:       float64(ps.MessagesCount) / elapsed.Seconds(),
		MeanThroughput: float64(ps.MessagesCount*ps.MessageSize) / elapsed.Seconds(),
	}, nil
}

func newBinaryULID() string {
	bytes, err := ulid.MustNew(ulid.Now(), cryptoRand.Reader).MarshalBinary()
	if err != nil {
		panic(err)
	}
	return string(bytes)
}

func (ps PubSub) payload() ([]byte, error) {
	msgPayload := make([]byte, ps.MessageSize)
	_, err := cryptoRand.Read(msgPayload)
	if err != nil {
		return nil, err
	}

	return msgPayload, nil
}
