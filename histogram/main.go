/*
histogram - generate a histogram from stdin values

Based on bitly's data_hacks
https://github.com/bitly/data_hacks/blob/master/data_hacks/histogram.py
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Data []float64

func (d Data) Len() int           { return len(d) }
func (d Data) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d Data) Less(i, j int) bool { return d[i] < d[j] }

type Histogram struct {
	input chan string
	stopInput chan bool
	inputComplete chan bool

	data Data

	max float64
	min float64
	mean float64
	count float64
	ss float64
}

func NewHistogram() (*Histogram, error) {
	h := &Histogram{
		min: math.MaxFloat64,
		max: -math.MaxFloat64,

		data: Data{},

		input: make(chan string),
		stopInput: make(chan bool),
		inputComplete: make(chan bool),
	}
	go h.parseInput()
	return h, nil
}

func (h *Histogram) parseInput() {

PARSE_LOOP:
	for {
		select {
		case val, ok := <- h.input:
			if !ok {
				break PARSE_LOOP
			}
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Println(err)
			}
			h.addValue(f)
		case <-h.stopInput:
			close(h.input)
		}
	}
	h.inputComplete <- true
}

func (h *Histogram) addValue(v float64) {
	h.data = append(h.data, v)
	if h.count == 0 {
		h.ss = 0
	} else {
		h.ss += (h.count * (v - h.mean) * (v - h.mean)) / (h.count + 1)
	}

	h.count++
	h.mean += (v - h.mean) / h.count
	if v < h.min {
		h.min = v
	}
	if v > h.max {
		h.max = v
	}
}

func (h *Histogram) Input() chan<- string {
	return h.input
}

func (h *Histogram) Close() {

	h.stopInput <- true
	<- h.inputComplete

	fmt.Printf("count:\t%0.f\n", h.count)
	fmt.Printf("mean:\t%f\n", h.mean)
	fmt.Printf("max:\t%f\n", h.max)
	fmt.Printf("min:\t%f\n", h.min)
	fmt.Printf("ss:\t%f\n", h.ss)
	fmt.Printf("var:\t%f\n", h.ss / h.count)
	fmt.Printf("sd:\t%f\n", math.Sqrt(h.ss / h.count))

	sort.Sort(h.data)
	bins := float64(10)
	w := (h.max - h.min) / bins
	idx := 0
	boundaries := []int{0}
	for bin := 0.0; bin < bins; bin++ {
		for {
			if h.data[idx] >= w * (bin + 1) + h.min {
				boundaries = append(boundaries, idx)
				break
			}
			idx++
		}
	}
	y := 0
	labels := []string{}
	repeats := []int{}
	maxY := 0
	for bin := 1; bin < int(bins) + 1; bin++ {
		if bin == int(bins) {
			y = boundaries[bin] - boundaries[bin - 1] + 1
		} else {
			y = boundaries[bin] - boundaries[bin - 1]
		}
		labels = append(labels, fmt.Sprintf("%10.4f - %10.4f [%6d]: ", h.min + (w * float64(bin - 1)), h.min + (w * float64(bin)), y))
		repeats = append(repeats, y)
		if y > maxY {
			maxY = y
		}
	}

	scale := maxY / 50
	fmt.Printf("∎:\t%d\n", scale)
	for i := 0; i < len(labels); i++ {
		fmt.Printf(labels[i])
		fmt.Println(strings.Repeat("∎", repeats[i] / scale))
	}
}

func main() {
	h, _ := NewHistogram()

	bio := bufio.NewReader(os.Stdin)

	for {
		fullLine := []byte{}
		line, hasMoreInLine, err := bio.ReadLine()
		fullLine = append(fullLine, line...)
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		for {
			if !hasMoreInLine {
				break
			}
			line, hasMoreInLine, err = bio.ReadLine()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			fullLine = append(fullLine, line...)
		}
		h.Input() <- string(fullLine)
	}

	h.Close()

}
