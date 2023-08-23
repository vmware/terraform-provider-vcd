package vcd

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

type partitionInfo struct {
	index int
	node  int
}

var (
	// numberOfPartitions is how many partitions we want to create
	numberOfPartitions = 0

	// partitionNode is the number of the current test runner
	partitionNode = 0

	// partitionDryRun will show what the partition would do, but won't run any tests
	partitionDryRun = false

	// mapOfTests is the list of tests, each with a sequential number and the node it is assigned to
	mapOfTests = make(map[string]partitionInfo)

	// partitionMx is a mutex used to guarantee that the map of tests is not accessed simultaneously
	partitionMx sync.Mutex

	// testMapMx is a mutex that controls the mapOfTests access
	testMapMx sync.Mutex
)

// getListOfTests retrieves the list of tests from the current directory
func getListOfTests() []string {
	files, err := os.ReadDir(".")
	if err != nil {
		panic(fmt.Errorf("error reading files in current directory: %s", err))
	}
	var testList []string

	// This regular expression finds every Test function declaration in the file
	// (?m) means multi-line, i.e. the '^' symbol matches at the start of each line
	// not only at the start of the text.
	findTestName := regexp.MustCompile(`(?m)^func (Test\w+)`)
	for _, f := range files {
		// skips non-test files
		if !strings.HasSuffix(f.Name(), "_test.go") {
			continue
		}
		// skips unit test files
		if strings.Contains(f.Name(), "_unit_test") {
			continue
		}
		fileContent, err := os.ReadFile(f.Name())
		if err != nil {
			panic(fmt.Errorf("error reading file %s: %s", f.Name(), err))
		}
		testNames := findTestName.FindAll(fileContent, -1)
		for _, fn := range testNames {
			// keeps only the test name
			testName := strings.Replace(string(fn), "func ", "", 1)
			testList = append(testList, testName)
		}
	}
	// The list of tests is sorted, so it will be the same in any node
	sort.Strings(testList)
	return testList
}

// getTestInfo retrieves test information in a thread-safe way
func getTestInfo(name string) (partitionInfo, bool) {
	testMapMx.Lock()
	defer testMapMx.Unlock()
	info, found := mapOfTests[name]
	return info, found
}

// getMapOfTests collects the list of tests and assigns node info
func getMapOfTests() map[string]partitionInfo {
	partitionMx.Lock()
	defer partitionMx.Unlock()
	// If this was the second access from a parallel test, we don't need to repeat the reading
	if len(mapOfTests) > 0 {
		return mapOfTests
	}
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

		// Rotate the node number
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
	if len(mapOfTests) == 0 {
		fmt.Printf("no tests found in this directory")
		os.Exit(1)
	}
	partInfo, found := getTestInfo(testName)
	if !found {
		fmt.Printf("test '%s' not found in the list of tests\n", testName)
		os.Exit(1)
	}

	if partInfo.node == partitionNode {
		fmt.Printf("[partitioning] [%d %s]\n", partInfo.index, testName)
		if partitionDryRun {
			t.Skipf("[DRY-RUN] partition node %d: test number %d ", partitionNode, partInfo.index)
		}
		// no action: the test belongs to the current node and will run
		return
	}
	// The test belong to a different node: skipping
	t.Skipf("not in partition %d : test '%s' number %d for node %d ", partitionNode, testName, partInfo.index, partInfo.node)
}
