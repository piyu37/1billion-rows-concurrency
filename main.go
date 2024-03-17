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

type stats struct {
	min, max, mean float64
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

	stationTemps := make(map[string][]float64)

	scanner := bufio.NewScanner(inputFile)

	for scanner.Scan() {
		text := scanner.Text()

		// refactoring: instead of using strings.Split, use strings.Index as
		// split creates a new slice for each split part, leading to more memory usage
		// and subsequent GC overhead
		index := strings.Index(text, ";")
		city := text[:index]
		tempString := text[index+1:]
		temp, _ := strconv.ParseFloat(tempString, 64)
		stationTemps[city] = append(stationTemps[city], temp)
	}

	result := make(map[string]stats)

	for city, tempreatures := range stationTemps {
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

		result[city] = stats{min, max, mean}
	}

	stations := make([]string, 0)

	for i := range result {
		stations = append(stations, i)
	}

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
