package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

var input = flag.String("input", "measurements.txt", "path of the input file to evaluate")
var output = flag.String("output", "output", "path of the output file")

const batchSize = 100

type stats struct {
	min, max, sum float64
	count         int64
}

func produceMeasurements(inputFile *os.File, batchCh chan<- []string) {
	scanner := bufio.NewScanner(inputFile)
	batch := make([]string, batchSize)
	count := 0
	for scanner.Scan() {
		text := scanner.Text()
		batch[count] = text
		count++

		if count == batchSize {
			localCopy := make([]string, batchSize)
			copy(localCopy, batch)
			batchCh <- localCopy
			count = 0
		}
	}

	if count != 0 {
		batchCh <- batch[:count]
	}

	close(batchCh)
}

func evaluate(filePath, outputFilePath string) {
	inputFile, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer inputFile.Close()

	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		panic(err)
	}

	defer outputFile.Close()

	batchCh := make(chan []string, 100)

	stationTempMap := make(map[string]stats)

	// decouple producer and consumer of file read using channel
	go produceMeasurements(inputFile, batchCh)

	for batch := range batchCh {
		for _, text := range batch {
			// refactoring: instead of using strings.Split, use strings.Index as
			// split creates a new slice for each split part, leading to more memory usage
			// and subsequent GC overhead
			index := strings.Index(text, ";")

			// handline case when we are not getting the line in correct format/improper text
			if index == -1 {
				continue
			}
			city := text[:index]
			tempString := text[index+1:]
			temp, _ := strconv.ParseFloat(tempString, 64)
			if v, ok := stationTempMap[city]; ok {
				if temp < v.min {
					v.min = temp
				}

				if temp > v.max {
					v.max = temp
				}

				v.sum += temp
				v.count++

				stationTempMap[city] = v
			} else {
				stationTempMap[city] = stats{
					min:   temp,
					max:   temp,
					sum:   temp,
					count: 1,
				}
			}
		}
	}

	formatAndWriteMesaurements(outputFile, stationTempMap)
}

func formatAndWriteMesaurements(outputFile *os.File, stationTempMap map[string]stats) {
	stations := make([]string, len(stationTempMap))
	count := 0
	for city := range stationTempMap {
		stations[count] = city
		count++
	}

	sort.Strings(stations)

	w := bufio.NewWriter(outputFile)

	for i := range stations {
		station := stationTempMap[stations[i]]
		_, err := w.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f\n",
			stations[i], station.min, station.sum/float64(station.count), station.max))
		if err != nil {
			panic(err)
		}
	}

	w.Flush()
}

func main() {
	flag.Parse()
	evaluate(*input, *output)
}
