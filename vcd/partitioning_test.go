package vcd

import (
	"fmt"
	"testing"
)

var partitioningCount = 0

func handlePartitioning(t *testing.T) {

	partitioningCount++
	fmt.Printf("%d %s\n", partitioningCount, t.Name())
	t.Skip("partitioning")
}
