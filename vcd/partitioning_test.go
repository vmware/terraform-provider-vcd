package vcd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type partitionInfo struct {
	index int
	node  int
}

var (
	numberOfPartitions = 0
	partitionNode      = 0
	partitionDryRun    = false
	//listofTests        = getListOfTests()
	mapOfTests = make(map[string]partitionInfo)
)

func getListOfTests() []string {
	files, err := os.ReadDir(".")
	if err != nil {
		panic(fmt.Errorf("error reading files in current directory: %s", err))
	}
	var testList []string

	findTestName := regexp.MustCompile(`(?m)^func (Test\w+)`)
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), "_test.go") {
			continue
		}
		if strings.Contains(f.Name(), "_unit_test") {
			continue
		}
		fileContent, err := os.ReadFile(f.Name())
		if err != nil {
			panic(fmt.Errorf("error reading file %s: %s", f.Name(), err))
		}
		testNames := findTestName.FindAll(fileContent, -1)
		for _, fn := range testNames {
			testName := strings.Replace(string(fn), "func ", "", 1)
			testList = append(testList, testName)
		}
	}
	// The list of tests is sorted, so it will be the same in any node
	sort.Strings(testList)
	return testList
}

func getMapOfTests() map[string]partitionInfo {
	listOfTests := getListOfTests()
	testNumber := 0

	nodeNumber := 0
	var testMap = make(map[string]partitionInfo)
	for _, tn := range listOfTests {
		if tn == "TestMain" {
			continue
		}
		// Every test gets assigned a number
		testNumber++
		nodeNumber++
		if nodeNumber > numberOfPartitions {
			nodeNumber = 1
		}
		testMap[tn] = partitionInfo{
			index: testNumber,
			node:  nodeNumber,
		}
	}
	return testMap
}

func handlePartitioning(t *testing.T) {
	if numberOfPartitions == 0 {
		return
	}
	if numberOfPartitions == 1 {
		fmt.Printf("number of partitions (-vcd-partitions) must be greater than 1\n")
		os.Exit(1)
	}
	if partitionNode == 0 {
		fmt.Printf("number of partitions (-vcd-partitions) was set, but not the partition node (-vcd-partition-node)\n")
		os.Exit(1)
	}
	if partitionNode > numberOfPartitions {
		fmt.Printf("partition node (%d) is bigger than number of partitions (%d)\n", partitionNode, numberOfPartitions)
		os.Exit(1)
	}
	testName := t.Name()
	if len(mapOfTests) == 0 {
		mapOfTests = getMapOfTests()
	}
	partInfo, found := mapOfTests[testName]
	if !found {
		fmt.Printf("test '%s' not found in the list of tests\n", testName)
		os.Exit(1)
	}

	if partInfo.node == partitionNode {
		fmt.Printf("[partitioning] [%d %s]\n", partInfo.index, testName)
		if partitionDryRun {
			t.Skipf("[DRY-RUN] partition node %d: test number %d ", partitionNode, partInfo.index)
		}
		// no action: the test will run
		return
	}
	t.Skipf("not in partition %d : test '%s' number %d for node %d ", partitionNode, testName, partInfo.index, partInfo.node)
}
