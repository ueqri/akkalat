package main

import (
	"flag"
	"fmt"

	"github.com/sarchlab/akkalat/runner"
	"gitlab.com/akita/mgpusim/v3/benchmarks/heteromark/pagerank"
)

var numNode = flag.Int("node", 16, "The number of nodes")
var sparsity = flag.Float64("sparsity", 0.001, "The sparsity of the graph")
var maxIterations = flag.Int("iterations", 16, "The number of iterations")

func main() {
	flag.Parse()

	runner := new(runner.Runner).ParseFlag().Init()

	benchmark := pagerank.NewBenchmark(runner.Driver())
	benchmark.NumNodes = uint32(*numNode)

	if *sparsity > 1 {
		*sparsity = 1
	}
	numConn := int(float64(*numNode**numNode) * *sparsity)
	if numConn < *numNode {
		numConn = *numNode
	}
	fmt.Printf("Number node %d, number connection %d\n", *numNode, numConn)

	benchmark.NumConnections = uint32(numConn)
	benchmark.MaxIterations = uint32(*maxIterations)

	runner.AddBenchmark(benchmark)

	runner.Run()
}
