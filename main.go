package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var input = flag.String("input", "measurements.txt", "path of the input file to evaluate")
var output = flag.String("output", "output", "path of the output file")

type stats struct {
	min, max, mean float64
}

func produceMeasurements(inputFile *os.File, lineCh chan<- string) {
	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		text := scanner.Text()
		lineCh <- text
	}

	close(lineCh)
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

	lineCh := make(chan string, 100)

	stationTemps := make(map[string][]float64)

	// decouple producer and consumer of file read using channel
	go produceMeasurements(inputFile, lineCh)

	for text := range lineCh {
		// refactoring: instead of using strings.Split, use strings.Index as
		// split creates a new slice for each split part, leading to more memory usage
		// and subsequent GC overhead
		index := strings.Index(text, ";")
		city := text[:index]
		tempString := text[index+1:]
		temp, _ := strconv.ParseFloat(tempString, 64)
		stationTemps[city] = append(stationTemps[city], temp)
	}

	formatAndWriteMesaurements(outputFile, stationTemps)
}

func formatAndWriteMesaurements(outputFile *os.File, stationTemps map[string][]float64) {
	result := make(map[string]stats)
	var mx sync.Mutex
	var wg sync.WaitGroup
	stations := make([]string, 0)

	for city, tempreatures := range stationTemps {
		stations = append(stations, city)
		wg.Add(1)
		// running separate goroutines for separate cities for faster process
		go func(city string, tempreatures []float64) {
			defer wg.Done()

			min, max, sum := tempreatures[0], tempreatures[0], 0.0

			for i := range tempreatures {
				if tempreatures[i] < min {
					min = tempreatures[i]
				}

				if tempreatures[i] > max {
					max = tempreatures[i]
				}

				sum += tempreatures[i]
			}

			mean := sum / float64(len(tempreatures))

			mx.Lock()
			result[city] = stats{min, max, mean}
			mx.Unlock()

		}(city, tempreatures)
	}

	wg.Wait()

	sort.Strings(stations)

	w := bufio.NewWriter(outputFile)

	for i := range stations {
		station := result[stations[i]]
		_, err := w.WriteString(fmt.Sprintf("%s=%.1f/%.1f/%.1f\n",
			stations[i], station.min, station.mean, station.max))
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
