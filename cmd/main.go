package main

import (
	"bytes"
	"flag"
	"fmt"
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

	var pubResults []pkg.Results
	var subResults []pkg.Results

	for _, size := range sizes {
		s, err := strconv.Atoi(size)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Starting benchmark for PubSub %s (%d messages, %d bytes each)\n",
			*pubsubFlag, *messagesCount, s)

		pubRes, subRes, err := pkg.RunBenchmark(*pubsubFlag, *messagesCount, s)
		if err != nil {
			panic(err)
		}

		pubResults = append(pubResults, pubRes)
		subResults = append(subResults, subRes)
	}

	fmt.Printf("\n\n")
	fmt.Println(generateTable(pubResults, subResults))
}

func padRight(str string, length int) string {
	if len(str) >= length {
		return str
	}
	return str + strings.Repeat(" ", length-len(str))
}

func generateTable(pubResults, subResults []pkg.Results) string {
	headers := []string{
		"Message size (bytes)",
		"Publish (messages / s)",
		"Subscribe (messages / s)",
	}

	// Calculate max width for each column based on headers and data
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	// Check data widths for each column including thousand separators
	for i := range pubResults {
		// Column 1: Message size
		msgSizeWidth := len(fmt.Sprintf("%d", pubResults[i].MessageSize))
		if msgSizeWidth > colWidths[0] {
			colWidths[0] = msgSizeWidth
		}

		// Column 2: Publish rate
		pubRateWidth := len(fmt.Sprintf("%d,%03d", int(pubResults[i].MeanRate)/1000, int(pubResults[i].MeanRate)%1000))
		if pubRateWidth > colWidths[1] {
			colWidths[1] = pubRateWidth
		}

		// Column 3: Subscribe rate
		subRateWidth := len(fmt.Sprintf("%d,%03d", int(subResults[i].MeanRate)/1000, int(subResults[i].MeanRate)%1000))
		if subRateWidth > colWidths[2] {
			colWidths[2] = subRateWidth
		}
	}

	var buf bytes.Buffer

	// Write header
	buf.WriteString("|")
	for i, header := range headers {
		buf.WriteString(" " + padRight(header, colWidths[i]) + " |")
	}
	buf.WriteString("\n")

	// Write separator
	buf.WriteString("|")
	for _, width := range colWidths {
		buf.WriteString("-" + strings.Repeat("-", width) + "-|")
	}
	buf.WriteString("\n")

	// Write data rows
	for i := range pubResults {
		buf.WriteString("|")

		// Message size (left-aligned)
		msgSize := fmt.Sprintf("%d", pubResults[i].MessageSize)
		buf.WriteString(" " + padRight(msgSize, colWidths[0]) + " |")

		// Publish rate (left-aligned) with thousand separator
		pubRate := fmt.Sprintf("%d,%03d", int(pubResults[i].MeanRate)/1000, int(pubResults[i].MeanRate)%1000)
		buf.WriteString(" " + padRight(pubRate, colWidths[1]) + " |")

		// Subscribe rate (left-aligned) with thousand separator
		subRate := fmt.Sprintf("%d,%03d", int(subResults[i].MeanRate)/1000, int(subResults[i].MeanRate)%1000)
		buf.WriteString(" " + padRight(subRate, colWidths[2]) + " |")

		buf.WriteString("\n")
	}

	return buf.String()
}
