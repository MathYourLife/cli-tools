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
	"strconv"
)

type Histogram struct {
	input chan string
	stopInput chan bool
	inputComplete chan bool

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
