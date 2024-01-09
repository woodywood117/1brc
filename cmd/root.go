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
	"unsafe"

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

	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buffer := bufio.NewReaderSize(file, 1024*1024*1024)

	// Read file line by line and calculate
	var line string
	var l []byte
	for {
		l, _, err = buffer.ReadLine()
		if err != nil {
			break
		}
		line = unsafe.String(unsafe.SliceData(l), len(l))
		//line = string(l)

		index := strings.Index(line, ";")
		station := line[:index]
		temp := line[index+1:]
		value, err := strconv.ParseFloat(temp, 64)
		if err != nil {
			err = errors.New("invalid line")
			break
		}

		measurement, ok := measurements[station]
		if !ok {
			measurement = &Measurement{
				Min:   100,
				Max:   0,
				Sum:   0,
				Count: 0,
			}
			measurements[fmt.Sprintf("%s", station)] = measurement
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
