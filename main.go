package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"
	"sync"
)

var input = flag.String("input", "measurements.txt", "path of the input file to evaluate")
var output = flag.String("output", "output", "path of the output file")

const chunkSize = 64 * 1024 * 1024

type stats struct {
	min, max, sum float64
	count         int64
}

func readMeasurements(inputFile *os.File, chunkStream chan []byte, resultStream chan map[string]stats,
	wg *sync.WaitGroup) {
	buf := make([]byte, chunkSize)
	leftover := make([]byte, 0, chunkSize)
	for {
		bytesRead, err := inputFile.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			panic(err)
		}

		buf = buf[:bytesRead]
		chunkToSend := make([]byte, bytesRead)
		copy(chunkToSend, buf)

		lastNewLineIndex := bytes.LastIndex(buf, []byte{'\n'})

		chunkToSend = append(leftover, buf[:lastNewLineIndex+1]...)
		leftover = make([]byte, len(buf[lastNewLineIndex+1:]))
		copy(leftover, buf[lastNewLineIndex+1:])

		chunkStream <- chunkToSend
	}

	close(chunkStream)

	// wait for all chunks to be proccessed before closing the result stream
	wg.Wait()

	close(resultStream)
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

	chunkStream := make(chan []byte, 15)
	resultStream := make(chan map[string]stats, 10)
	var wg sync.WaitGroup
	stationTempMap := make(map[string]stats)

	// running gorutines no. of cpu cores-1 parallel process file read chunks
	for i := 0; i < runtime.NumCPU()-1; i++ {
		wg.Add(1)

		go func() {
			for chunk := range chunkStream {
				processReadChunk(chunk, resultStream)
			}

			wg.Done()
		}()
	}

	// decouple producer and consumer of file read using channel
	go readMeasurements(inputFile, chunkStream, resultStream, &wg)

	for data := range resultStream {
		for city, tempInfo := range data {
			if val, ok := stationTempMap[city]; ok {
				if tempInfo.min < val.min {
					val.min = tempInfo.min
				}

				if tempInfo.max > val.max {
					val.max = tempInfo.max
				}

				val.sum += tempInfo.sum
				val.count += tempInfo.count

				stationTempMap[city] = val
			} else {
				stationTempMap[city] = tempInfo
			}
		}
	}

	formatAndWriteMesaurements(outputFile, stationTempMap)
}

func processReadChunk(buf []byte, resultStream chan<- map[string]stats) {
	stationTempMap := make(map[string]stats)
	var city string
	var start int

	stringBuf := string(buf)

	for index, char := range stringBuf {
		switch char {
		case ';':
			city = stringBuf[start:index]
			start = index + 1
		case '\n':
			if index-start > 0 && len(city) != 0 {
				temp, _ := strconv.ParseFloat(stringBuf[start:index], 64)
				start = index + 1
				if val, ok := stationTempMap[city]; ok {
					if temp < val.min {
						val.min = temp
					}

					if temp > val.max {
						val.max = temp
					}

					val.sum += temp
					val.count++

					stationTempMap[city] = val
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
	}

	resultStream <- stationTempMap
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
