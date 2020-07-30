package main

import (
	"bufio"
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

// Use bash to run this with an increasing number of iterations:
// ```
// for n in $(seq 0 100)
// do
//   go run main.go -n $n; sleep 1
// done
// ```

var iterations int

func main() {
	flag.IntVar(&iterations, "n", 1000, "number of iterations")
	flag.Parse()

	xys, err := readData("data.txt")
	if err != nil {
		log.Fatalf("could not read data.txt: %v", err)
	}
	_ = xys

	err = plotData("out.png", xys)
	if err != nil {
		log.Fatalf("could not plot data: %v", err)
	}
}

type xy struct{ x, y float64 }

func readData(path string) (plotter.XYs, error) {
	// read line by line, not all at once (like ioutil.ReadFile)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var xys plotter.XYs
	s := bufio.NewScanner(f)
	for s.Scan() {
		var x, y float64
		_, err := fmt.Sscanf(s.Text(), "%f,%f", &x, &y)
		if err != nil {
			log.Printf("discarding bad data point: %q: %v", s.Text(), err)
		}
		xys = append(xys, plotter.XY{X: x, Y: y})
	}
	if err := s.Err(); err != nil {
		return nil, fmt.Errorf("could not scan: %v", err)
	}

	return xys, nil
}

func plotData(path string, xys plotter.XYs) error {
	// Create a file to write the plot to
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create %s: %v", path, err)
	}

	// Plot the data
	p, err := plot.New()
	if err != nil {
		return fmt.Errorf("could not create plot: %v", err)
	}

	// create scatter with all data points
	s, err := plotter.NewScatter(xys)
	s.GlyphStyle.Shape = draw.CrossGlyph{}
	s.Color = color.RGBA{R: 255, A: 255}
	if err != nil {
		return fmt.Errorf("could not create scatter: %v", err)
	}
	p.Add(s)

	// create linear regression result
	m, c := linearRegression(xys, 0.01)
	l, err := plotter.NewLine(plotter.XYs{
		{X: 3, Y: (3 * m) + c}, {X: 20, Y: (20 * m) + c},
	})
	if err != nil {
		return fmt.Errorf("could not create line: %v", err)
	}
	p.Add(l)

	wt, err := p.WriterTo(512, 512, "png")
	if err != nil {
		return fmt.Errorf("could not create writer: %v", err)
	}
	_, err = wt.WriteTo(f)
	if err != nil {
		return fmt.Errorf("could not write to %s: %v", path, err)
	}

	// Make sure the file closes properly,
	// otherwise the data might not have plotted correctly
	if err = f.Close(); err != nil {
		return fmt.Errorf("could not close %s: %v", path, err)
	}
	return nil
}

func linearRegression(xys plotter.XYs, alpha float64) (m, c float64) {
	// NOTE: This is a simple way of finding the line of best fit
	// ...BUT, it takes 4004001 iterations
	// const (
	// 	min   = -100.0
	// 	max   = 100.0
	// 	delta = 0.1
	// )
	// count := 0
	// minCost := math.MaxFloat64
	// for im := min; im < max; im += delta {
	// 	for ic := min; ic < max; ic += delta {
	// 		count++
	// 		cost := computeCost(xys, im, ic)
	// 		if cost < minCost {
	// 			minCost = cost
	// 			m, c = im, ic
	// 			dm, dc := computeGradient(xys, m, c)
	// 			fmt.Printf("grad(%.2f, %.2f) = (%.2f, %.2f)\n", m, c, dm, dc)
	// 		}
	// 	}
	// }
	// fmt.Printf("cost(%.2f, %.2f) = %.2f\n", m, c, computeCost(xys, m, c))
	// fmt.Printf("tried %d times\n", count)

	// NOTE: Instead, lets use gradient descent:
	for i := 0; i < iterations; i++ {
		dm, dc := computeGradient(xys, m, c)
		m += -dm * alpha
		c += -dc * alpha
		// fmt.Printf("grad(%.2f, %.2f) = (%.2f, %.2f)\n", m, c, dm, dc)
	}
	fmt.Printf("cost(%.2f, %.2f) = %.2f\n", m, c, computeCost(xys, m, c))

	return m, c
}

func computeCost(xys plotter.XYs, m, c float64) float64 {
	// cost = 1/N * sum((y - (m*x+c))^2)
	s := 0.0
	for _, xy := range xys {
		d := xy.Y - (xy.X*m + c)
		s += d * d
	}
	return s / float64(len(xys))
}

func computeGradient(xys plotter.XYs, m, c float64) (dm, dc float64) {
	// cost = 1/N sum((y - (m*x+c))^2)
	// d = (y - (m*x+c))
	// cost/dm = 2/N * sum(-x * d)
	// cost/dc = 2/N * sum(-d)
	for _, xy := range xys {
		d := xy.Y - (xy.X*m + c)
		dm += -xy.X * d
		dc += -d
	}
	n := float64(len(xys))
	return 2 / n * dm, 2 / n * dc
}
