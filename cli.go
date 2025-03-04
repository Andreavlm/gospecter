package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

func main() {
	inputFile := flag.String("input", "", "Path to the input WAV file")
	outputFile := flag.String("output", "output.png", "Path to the output PNG file")
	width := flag.Int("width", 1600, "Width of the spectrogram")
	height := flag.Int("height", 1200, "Height of the spectrogram")
	windowSize := flag.Int("window", 2048, "Window size for the FFT")
	help := flag.Bool("help", false, "Display help")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *inputFile == "" {
		fmt.Println("Error: input file is required")
		flag.Usage()
		os.Exit(1)
	}

	err := createSpectrogram(*inputFile, *outputFile, *width, *height, *windowSize)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	} else {
		fmt.Println("Spectrogram created successfully")
	}
}

func parseInt(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		fmt.Println("Error: invalid integer value", s)
		os.Exit(1)
	}
	return value
}
