package cmd

import (
	"bufio"
	"errors"
	"github.com/spf13/cobra"
	"io"
	"math/bits"
	"os"
	"unsafe"
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

const Gigabyte = 1024 * 1024 * 1024

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

	buffer := bufio.NewReaderSize(file, Gigabyte)

	// Read file line by line and calculate
	var l []byte
	var index int
	var position int
	var station string
	var temp float64
	for {
		l, _, err = buffer.ReadLine()
		if err != nil {
			break
		}

		index = 0
		position = 0
		for {
			word := *(*uint64)(unsafe.Pointer(&l[position]))
			index = FindDelimiter(word)
			if index != 8 {
				position += index
				break
			}
			position += 8
		}
		index = position

		station = unsafe.String(unsafe.SliceData(l[:index]), index)
		temp = FastParseFloat(l[index+1:])

		measurement, ok := measurements[station]
		if !ok {
			measurement = &Measurement{
				Min:   100,
				Max:   0,
				Sum:   0,
				Count: 0,
			}
			measurements[string(l[:index])] = measurement
		}
		measurement.Min = min(measurement.Min, temp)
		measurement.Max = max(measurement.Max, temp)
		measurement.Sum += temp
		measurement.Count++
	}
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}

	// Sort keys alphabetically
	//var keys []string
	//for key := range measurements {
	//	keys = append(keys, key)
	//}
	//slices.Sort(keys)

	// Print results
	//for _, key := range keys {
	//	measurement := measurements[key]
	//	avg := measurement.Sum / float64(measurement.Count)
	//	fmt.Printf("%s: %.1f %.1f %.1f\n", key, measurement.Min, avg, measurement.Max)
	//}
}

// FastParseFloat parses floats in the format of "12.3" or "-12.3".
func FastParseFloat(s []byte) float64 {
	i := uint(0)
	minus := s[0] == '-'
	if minus {
		i++
	}

	d := uint64(0)
	for i < uint(len(s)) {
		if s[i] >= '0' && s[i] <= '9' {
			d = d*10 + uint64(s[i]-'0')
			i++
			continue
		}
		break
	}
	i++

	// Fast path - just integer.
	if s[i] == '0' {
		f := float64(d)
		if minus {
			f = -f
		}
		return f
	}

	d = d*10 + uint64(s[i]-'0')
	i++
	f := float64(d) / 1e1
	if minus {
		f = -f
	}
	return f
}

// FindDelimiter returns the index of the first byte that is a semicolon.
// It's magic and uses some bit twiddling.
// It's faster because it checks 8 bytes at once.
func FindDelimiter(word uint64) int {
	input := word ^ 0x3B3B3B3B3B3B3B3B
	tmp := input - 0x0101010101010101
	tmp = tmp & ^input
	tmp = tmp & 0x8080808080808080
	return bits.TrailingZeros64(tmp) >> 3
}
