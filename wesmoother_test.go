package smoother

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func loadFile(filename string) ([]float64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var numbers []float64
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		number, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
		if err != nil {
			// ignore any non-numeric lines
			continue
		}
		numbers = append(numbers, number)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return numbers, nil
}

func TestWESmoother(t *testing.T) {
	filenames := []string{"docs/nmr.dat", "docs/wood.txt"}
	lambda := 10.0
	d := 2

	for _, filename := range filenames {
		data, err := loadFile(filename)
		if err != nil {
			t.Fatalf("Failed to load file: %v", err)
		}

		_, err = WESmoother(data, lambda, d)
		if err != nil {
			t.Fatalf("Failed to apply WESmoother: %v", err)
		}
	}
}

func BenchmarkWESmoother(b *testing.B) {
	filenames := []string{"docs/nmr.dat", "docs/wood.txt"}
	lambda := 10.0
	d := 2

	for _, filename := range filenames {
		data, err := loadFile(filename)
		if err != nil {
			b.Fatalf("Failed to load file: %v", err)
		}

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err = WESmoother(data, lambda, d)
			if err != nil {
				b.Fatalf("Failed to apply WESmoother: %v", err)
			}
		}
	}
}
