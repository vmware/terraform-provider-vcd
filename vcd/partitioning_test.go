// Do not add tags to this file

package vcd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
)

// partitionInfo contains the progressive number of the test (index) and the node to which it is assigned
type partitionInfo struct {
	index int
	node  int
}

var (
	// testListFileName is the name of a file containing the tests to run in this node, prepared by a third party
	testListFileName string

	// numberOfPartitions is how many partitions we want to create
	numberOfPartitions = 0

	// partitionNode is the number of the current test runner
	partitionNode = 0

	// mapOfTests is the list of tests, each with a sequential number and the node it is assigned to
	mapOfTests = make(map[string]partitionInfo)

	// partitionMutexes is a set of mutexes used to guarantee that partitioning elements are thread safe
	partitionMutexes = map[string]*sync.Mutex{
		"list_creation":  {},
		"map_access":     {},
		"file_access":    {},
		"progress_count": {},
	}

	// listOfTestForNode contains the tests for the current node
	listOfTestForNode []string

	// listOfProcessedTests contains the list of tests processed by the partitioned node
	listOfProcessedTests []string

	// notPartitionedTests are the tests that have side effects and should not be partitioned
	notPartitionedTests = []string{
		"TestMain",               // mother of all tests. Can't be split
		"TestTags",               // only runs if the suite is called without tags
		"TestProvider",           // used to create provider
		"TestProvider_impl",      // used to define provider
		"TestCustomTemplates",    // used to fill binary tests
		"TestAccClientUserAgent", // sets user agent
	}

	// progressiveCount marks the current test being processed by the node
	progressiveCount int = 0
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
			if contains(notPartitionedTests, testName) {
				continue
			}
			testList = append(testList, testName)
		}
	}
	// The list of tests is sorted, so it will be the same in any node
	sort.Strings(testList)
	return testList
}

// getTestInfo retrieves test information in a thread-safe way
func getTestInfo(name string) (partitionInfo, bool) {
	partitionMutexes["map_access"].Lock()
	defer partitionMutexes["map_access"].Unlock()
	info, found := mapOfTests[name]
	return info, found
}

// getPreCompiledMapOfTests retrieves the list of tests from a file created by a third party
func getPreCompiledMapOfTests() map[string]partitionInfo {
	readFile, err := os.Open(testListFileName)
	if err != nil {
		fmt.Printf("error opening '%s': %s\n", testListFileName, err)
		os.Exit(1)
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	index := 0
	for fileScanner.Scan() {
		testName := strings.TrimSpace(fileScanner.Text())
		// Skipping comments and empty lines
		if strings.HasPrefix(testName, "#") {
			continue
		}
		if testName == "" {
			continue
		}
		index++
		mapOfTests[testName] = partitionInfo{
			index: index,
			node:  partitionNode,
		}
	}
	if len(mapOfTests) == 0 {
		fmt.Printf("no test names found in file '%s'\n", testListFileName)
		os.Exit(1)
	}
	return mapOfTests
}

// getMapOfTests collects the list of tests and assigns node info
func getMapOfTests(vcdVersion, providerUrl string) map[string]partitionInfo {
	partitionMutexes["list_creation"].Lock()
	defer partitionMutexes["list_creation"].Unlock()
	// If this was the second access from a parallel test, we don't need to repeat the reading
	if len(mapOfTests) > 0 {
		return mapOfTests
	}
	if testListFileName != "" {
		return getPreCompiledMapOfTests()
	}
	listOfTests := getListOfTests()
	testNumber := 0

	nodeNumber := 0
	var testMap = make(map[string]partitionInfo)
	for _, tn := range listOfTests {
		// Every test gets assigned a number
		testNumber++

		// Rotate the node number
		nodeNumber++
		if nodeNumber > numberOfPartitions {
			nodeNumber = 1
		}
		if nodeNumber == partitionNode {
			listOfTestForNode = append(listOfTestForNode, tn)
		}
		testMap[tn] = partitionInfo{
			index: testNumber,
			node:  nodeNumber,
		}
	}
	retrievedTestsFileName := getTestFileName("retrieved", vcdVersion)
	_ = os.WriteFile(retrievedTestsFileName, []byte(strings.Join(listOfTests, "\n")), 0600)

	plannedTestsFileName := getTestFileName("planned", vcdVersion)
	var testList strings.Builder
	testList.WriteString(fmt.Sprintf("# VCD: %s \n", strings.Replace(providerUrl, "/api", "", 1)))
	testList.WriteString(fmt.Sprintf("# version: %s \n", vcdVersion))
	testList.WriteString(fmt.Sprintf("# node N.: %d \n", partitionNode))
	for _, tn := range listOfTestForNode {
		testList.WriteString(fmt.Sprintf("%s\n", tn))
	}
	_ = os.WriteFile(plannedTestsFileName, []byte(testList.String()), 0600)
	return testMap
}

// getTestFileName builds a file name for one of the testing artefacts
func getTestFileName(namePart, vcdVersion string) string {
	prefix := os.Getenv("BUILD_NUMBER")
	if prefix == "" {
		prefix = "LOCAL"
	}
	return fmt.Sprintf("%s-tests-%s-in-node-%d-%s.txt", prefix, namePart, partitionNode, vcdVersion)
}

// writeProcessedTests adds a line to an existing file.
// If the file does not exist, it gets created before adding the line
func writeProcessedTests(fileName, line string) {
	partitionMutexes["file_access"].Lock()
	defer partitionMutexes["file_access"].Unlock()
	var fileHandler *os.File
	var err error
	if fileExists(fileName) {
		fileHandler, err = os.OpenFile(filepath.Clean(fileName), os.O_WRONLY|os.O_APPEND, os.ModeAppend)
	} else {
		fileHandler, err = os.Create(filepath.Clean(fileName))
	}
	if err != nil {
		fmt.Printf("##### ERROR opening file %s : %s\n", fileName, err)
		os.Exit(1)
	}
	defer safeClose(fileHandler)
	w := bufio.NewWriter(fileHandler)
	_, err = fmt.Fprintln(w, line)
	if err != nil {
		fmt.Printf("error writing to file %s: %s\n", fileName, err)
		os.Exit(1)
	}
	_ = w.Flush()
}

// progressCountIncrease adds to the progressive number of tests in a safe mode
func progressCountIncrease() {
	partitionMutexes["progress_count"].Lock()
	defer partitionMutexes["progress_count"].Unlock()
	progressiveCount++
}

// handlePartitioning is called by every test in the suite
func handlePartitioning(vcdVersion, providerUrl string, t *testing.T) {
	// If partitioning is not enabled we bail out immediately
	if numberOfPartitions == 0 {
		return
	}
	// Number of partitions should be at least 2
	if numberOfPartitions == 1 {
		fmt.Printf("number of partitions (-vcd-partitions) must be greater than 1\n")
		os.Exit(1)
	}

	// When partitioning is enabled, we must have a node identified for the current run
	if partitionNode == 0 {
		fmt.Printf("number of partitions (-vcd-partitions) was set, but not the partition node (-vcd-partition-node)\n")
		os.Exit(1)
	}

	// The current node can't be higher than the number of partitions
	if partitionNode > numberOfPartitions {
		fmt.Printf("partition node (%d) is bigger than number of partitions (%d)\n", partitionNode, numberOfPartitions)
		os.Exit(1)
	}
	testName := t.Name()
	// If this is the first test being processed, we collect the list of tests
	if len(mapOfTests) == 0 {
		// The list of tests will be produced locally with a simple, repeatable algorithm (the default)
		// or be retrieved from a file created by a third party
		mapOfTests = getMapOfTests(vcdVersion, providerUrl)
	}
	if len(mapOfTests) == 0 {
		fmt.Printf("no tests found in this directory")
		os.Exit(1)
	}
	partInfo, found := getTestInfo(testName)
	if !found {
		if testListFileName == "" {
			// The test name was retrieved by the internal algorithm. If it is missing, one of the following could
			// be happening (in order of likelihood):
			// 1. lack of preTestCheck call in the test function
			// 2. not formatted Go files that could fool the test name detection
			// 3. a manipulation of the input files
			// 4. an error in the partitioning algorithm
			fmt.Printf("test '%s' not found in the list of tests\n", testName)
			os.Exit(1)
		}
		// The test was retrieved from a third-party file. There is no information about other nodes.
		t.Skipf("not scheduled to run in partition %d : test '%s' ", partitionNode, testName)
	}

	if partInfo.node == partitionNode {
		fileName := getTestFileName("processed", vcdVersion)
		if len(listOfProcessedTests) == 0 {
			// This is the first test: we want to start with an empty test list file
			if fileExists(fileName) {
				_ = os.Remove(fileName)
			}
			writeProcessedTests(fileName, "# LIST OF PROCESSED TESTS")
			writeProcessedTests(fileName, fmt.Sprintf("# VCD    : %s", strings.Replace(providerUrl, "/api", "", 1)))
			writeProcessedTests(fileName, fmt.Sprintf("# version : %s", vcdVersion))
		}
		// We may have a duplicate when a test has more than one calls to preTestCheck function
		if !contains(listOfProcessedTests, testName) {
			progressCountIncrease()
			writeProcessedTests(fileName, testName)
			listOfProcessedTests = append(listOfProcessedTests, testName)
		}
		fmt.Printf("[partitioning] [%s {%d of %d} (node %d)]\n", testName, progressiveCount, len(listOfTestForNode), partitionNode)
		// no action: the test belongs to the current node and will run
		return
	}
	// The test belong to a different node: skipping
	t.Skipf("not in partition %d : test '%s' number %d for node %d ", partitionNode, testName, partInfo.index, partInfo.node)
}
