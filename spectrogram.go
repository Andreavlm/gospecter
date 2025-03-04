package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
)

func createSpectrogram(inputFile, outputFile string, width, height, windowSize int) error {
	if filepath.Ext(inputFile) != ".wav" {
		return fmt.Errorf("input file is not a wav file")
	}

	file, err := os.Open(inputFile)
	if err != nil {
		return fmt.Errorf("failed to open input file: %v", err)
	}
	defer file.Close()

	var header [44]byte
	_, err = file.Read(header[:])
	if err != nil {
		return fmt.Errorf("failed to read wav file header: %v", err)
	}

	sampleRate := int(header[24]) | int(header[25])<<8 | int(header[26])<<16 | int(header[27])<<24
	dataSize := int(header[40]) | int(header[41])<<8 | int(header[42])<<16 | int(header[43])<<24

	data := make([]byte, dataSize)
	_, err = file.Read(data)
	if err != nil {
		return fmt.Errorf("failed to read wav file data: %v", err)
	}

	// Convert byte data to float64 samples
	samples := make([]float64, dataSize/2)
	for i := 0; i < len(samples); i++ {
		sample := int16(data[2*i]) | int16(data[2*i+1])<<8
		samples[i] = float64(sample) / 32768.0
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		start := x * (len(samples) - windowSize) / width
		end := start + windowSize
		if end > len(samples) {
			end = len(samples)
		}
		window := samples[start:end]
		spectrum := fft(window)
		for y := 0; y < height; y++ {
			frequency := float64(y) / float64(height) * float64(sampleRate) / 2
			index := int(frequency * float64(windowSize) / float64(sampleRate))
			if index < len(spectrum) {
				amplitude := math.Sqrt(spectrum[index].Real*spectrum[index].Real + spectrum[index].Imag*spectrum[index].Imag)
				colorValue := uint8(math.Log10(amplitude+1) * 255 / math.Log10(32768))
				img.Set(x, height-y-1, colormap(colorValue))
			}
		}
	}

	outFile, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if err != nil {
		return fmt.Errorf("failed to encode spectrogram to png: %v", err)
	}

	return nil
}

func colormap(value uint8) color.Color {
	r := uint8(math.Min(255, float64(value)*2))
	g := uint8(0)
	b := uint8(0)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

type complex struct {
	Real, Imag float64
}

func fft(x []float64) []complex {
	N := len(x)
	if N <= 1 {
		return []complex{{Real: x[0], Imag: 0}}
	}

	even := make([]float64, N/2)
	odd := make([]float64, N/2)
	for i := 0; i < N/2; i++ {
		even[i] = x[i*2]
		odd[i] = x[i*2+1]
	}

	evenFFT := fft(even)
	oddFFT := fft(odd)

	T := make([]complex, N)
	for k := 0; k < N/2; k++ {
		t := complex{
			Real: math.Cos(2*math.Pi*float64(k)/float64(N))*oddFFT[k].Real - math.Sin(2*math.Pi*float64(k)/float64(N))*oddFFT[k].Imag,
			Imag: math.Sin(2*math.Pi*float64(k)/float64(N))*oddFFT[k].Real + math.Cos(2*math.Pi*float64(k)/float64(N))*oddFFT[k].Imag,
		}
		T[k] = complex{
			Real: evenFFT[k].Real + t.Real,
			Imag: evenFFT[k].Imag + t.Imag,
		}
		T[k+N/2] = complex{
			Real: evenFFT[k].Real - t.Real,
			Imag: evenFFT[k].Imag - t.Imag,
		}
	}

	return T
}
