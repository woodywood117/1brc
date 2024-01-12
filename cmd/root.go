package cmd

import (
	"1brc/fastmap"
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

type Measurement struct {
	Min   int64
	Max   int64
	Sum   int64
	Count int
}

const Gigabyte = 1024 * 1024 * 1024
const ChunkSize = Gigabyte

func run(_ *cobra.Command, _ []string) {
	var measurements = fastmap.New[*Measurement]()

	file, err := os.Open("measurements.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var data = make([]byte, ChunkSize, ChunkSize)
	var leftover []byte
	var eof bool
	for {
		if len(leftover) > 0 {
			copy(data, leftover)
		}
		written, err := FillChunk(data[len(leftover):], file)
		if err != nil {
			if errors.Is(err, io.EOF) {
				eof = true
			} else {
				panic(err)
			}
		}

		if written+len(leftover) < len(data) {
			if !eof {
				panic("Should not happen")
			}
			data = data[:written+len(leftover)]
		}

		leftover = ParseChunk(data, measurements)

		if eof {
			break
		}
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
	//	avg := (float64(measurement.Sum) / 1e1) / float64(measurement.Count)
	//	fmt.Printf("%s: %.1f %.1f %.1f\n", key, float64(measurement.Min)/1e1, avg, float64(measurement.Max)/1e1)
	//}
}

// FastParseFloat parses floats in the format of "12.3" or "-12.3".
func FastParseFloat(s []byte) int64 {
	i := uint(0)
	minus := s[0] == '-'
	if minus {
		i++
	}

	d := int64(0)
	for i < uint(len(s)) {
		if s[i] >= '0' && s[i] <= '9' {
			d = d*10 + int64(s[i]-'0')
			i++
			continue
		}
		break
	}
	i++

	// Fast path - just integer.
	if s[i] == '0' {
		if minus {
			d = -d
		}
		return d
	}

	d = d*10 + int64(s[i]-'0')
	i++
	if minus {
		d = -d
	}
	return d
}

// FindSemicolon returns the index of the first byte that is a semicolon.
// It's magic and uses some bit twiddling.
// It's faster because it checks 8 bytes at once.
func FindSemicolon(word uint64) int {
	input := word ^ 0x3B3B3B3B3B3B3B3B
	tmp := input - 0x0101010101010101
	tmp = tmp & ^input
	tmp = tmp & 0x8080808080808080
	return bits.TrailingZeros64(tmp) >> 3
}
func FindNewLine(word uint64) int {
	input := word ^ 0x0A0A0A0A0A0A0A0A
	tmp := input - 0x0101010101010101
	tmp = tmp & ^input
	tmp = tmp & 0x8080808080808080
	return bits.TrailingZeros64(tmp) >> 3
}

func ParseChunk(chunk []byte, measurements *fastmap.Map[*Measurement]) (leftover []byte) {
	var l []byte
	var start int
	var word uint64
	var newline int
	var semi int
	var position int
	var end int
	var station string
	var temp int64
	for {
		if start >= len(chunk) {
			break
		}
		end = start
		position = 0
		for end+position < len(chunk) {
			word = *(*uint64)(unsafe.Pointer(&chunk[end+position]))
			newline = FindNewLine(word)
			if newline != 8 {
				position += newline
				break
			}
			position += 8
		}
		end += position
		if end >= len(chunk) {
			leftover = chunk[start:]
			return
		}
		l = chunk[start:end]
		start = end + 1

		semi = 0
		position = 0
		for {
			word = *(*uint64)(unsafe.Pointer(&l[position]))
			semi = FindSemicolon(word)
			if semi != 8 {
				position += semi
				break
			}
			position += 8
		}
		semi = position

		station = unsafe.String(unsafe.SliceData(l[:semi]), semi)
		temp = FastParseFloat(l[semi+1:])

		measurement, ok := measurements.Get(station)
		if !ok {
			measurement = &Measurement{
				Min:   100,
				Max:   0,
				Sum:   0,
				Count: 0,
			}
			measurements.Set(string(l[:semi]), measurement)
		}
		measurement.Min = min(measurement.Min, temp)
		measurement.Max = max(measurement.Max, temp)
		measurement.Sum += temp
		measurement.Count++
	}
	return
}

func FillChunk(chunk []byte, reader io.Reader) (int, error) {
	var total int
	for {
		n, err := reader.Read(chunk[total:])
		total += n
		if n == 0 || err != nil {
			return total, err
		}
	}
}
