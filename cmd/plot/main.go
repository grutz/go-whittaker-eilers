package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	smoother "github.com/grutz/go-whittaker-eilers"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func floatsFromFile(filename string) ([]float64, error) {
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
			// skip non-float lines
			continue
		}
		numbers = append(numbers, number)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return numbers, nil
}

func makePoints(y []float64) plotter.XYs {
	pts := make(plotter.XYs, len(y))
	for i := range pts {
		pts[i].X = float64(i)
		pts[i].Y = y[i]
	}
	return pts
}

func do(filename string) {
	data, err := floatsFromFile(filename)
	if err != nil {
		panic(err)
	}
	basename := filepath.Base(filename)
	fmt.Printf("Working on %s\n", basename)

	// Do 5 iterations of smoothing
	var cleans [][]float64
	for _, lambda := range []float64{5, 10, 50, 100, 500} {
		clean, err := smoother.WESmoother(data, lambda, 2)
		if err != nil {
			panic(err)
		}
		cleans = append(cleans, clean)
		p := plot.New()
		p.Title.Text = fmt.Sprintf("%s: Orig vs. %f Lambda", basename, lambda)
		p.X.Label.Text = "X"
		p.Y.Label.Text = "Y"
		err = plotutil.AddLines(
			p,
			fmt.Sprintf("Lambda %f", lambda), makePoints(clean),
			basename, makePoints(data),
		)
		if err != nil {
			panic(err)
		}

		err = p.Save(20*vg.Inch, 10*vg.Inch, fmt.Sprintf("%s-lambda-%d.png", basename, int(lambda)))
		if err != nil {
			panic(err)
		}
	}

	// Make the combined plot file
	p := plot.New()
	p.Title.Text = fmt.Sprintf("%s: Orig vs Clean", basename)
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	err = plotutil.AddLines(
		p,
		"Clean 5", makePoints(cleans[0]),
		"Clean 10", makePoints(cleans[1]),
		"Clean 50", makePoints(cleans[2]),
		"Clean 100", makePoints(cleans[3]),
		"Clean 500", makePoints(cleans[4]),
		basename, makePoints(data),
	)
	if err != nil {
		panic(err)
	}

	err = p.Save(20*vg.Inch, 10*vg.Inch, fmt.Sprintf("%s-combined.png", basename))
	if err != nil {
		panic(err)
	}

}

func main() {
	do("../../docs/nmr.dat")
	do("../../docs/wood.txt")
}
