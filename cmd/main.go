package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ThreeDotsLabs/watermill-benchmark/pkg"
)

const (
	defaultMessageSize = "16,64,256"
)

var pubsubFlag = flag.String("pubsub", "", "")
var messagesCount = flag.Int("count", 0, "")
var messageSizes = flag.String("size", defaultMessageSize, "comma-separated list of message sizes")

func main() {
	flag.Parse()
	sizes := strings.Split(*messageSizes, ",")

	var subResults []pkg.Results

	for _, size := range sizes {
		s, err := strconv.Atoi(size)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Starting benchmark for PubSub %s (%d messages, %d bytes each)\n",
			*pubsubFlag, *messagesCount, s)

		_, subRes, err := pkg.RunBenchmark(*pubsubFlag, *messagesCount, s)
		if err != nil {
			log.Fatal(err)
		}

		subResults = append(subResults, subRes)
	}

	fmt.Printf("messages\tmessage size\trate (messages/s)\tthroughput (b/s)\n")
	for _, r := range subResults {
		fmt.Printf("%d\t%d\t%f\t%f\n", r.Count, r.MessageSize, r.MeanRate, r.MeanThroughput)
	}
}
