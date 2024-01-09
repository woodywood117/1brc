package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "1brc",
	Run: run,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	type Measurement struct {
		Min   float64
		Max   float64
		Sum   float64
		Count int
	}
	var measurements = make(map[string]*Measurement)

	file, err := os.Open("measurements_1M.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buffer := bufio.NewReaderSize(file, 1024*1024*1024)
	//buffer := bufio.NewReader(file)

	// Read file line by line and calculate
	var line string
	for {
		line, err = buffer.ReadString('\n')
		if err != nil {
			break
		}
		line = line[:len(line)-1]

		parts := strings.Split(line, ";")
		if len(parts) != 2 {
			err = errors.New("invalid line")
			break
		}

		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			err = errors.New("invalid line")
			break
		}

		measurement, ok := measurements[parts[0]]
		if !ok {
			measurement = &Measurement{
				Min:   100,
				Max:   0,
				Sum:   0,
				Count: 0,
			}
			measurements[parts[0]] = measurement
		}
		measurement.Min = min(measurement.Min, value)
		measurement.Max = max(measurement.Max, value)
		measurement.Sum += value
		measurement.Count++
	}
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}

	// Sort keys alphabetically
	var keys []string
	for key := range measurements {
		keys = append(keys, key)
	}
	slices.Sort(keys)

	// Print results
	for _, key := range keys {
		measurement := measurements[key]
		avg := measurement.Sum / float64(measurement.Count)
		fmt.Printf("%s: %.1f %.1f %.1f\n", key, measurement.Min, avg, measurement.Max)
	}
}
