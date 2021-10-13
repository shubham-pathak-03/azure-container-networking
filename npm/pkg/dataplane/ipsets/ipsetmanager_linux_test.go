package ipsets

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ioutil"
	"github.com/Azure/azure-container-networking/npm/util"
	testutils "github.com/Azure/azure-container-networking/test/utils"

	"github.com/stretchr/testify/require"
)

type testSet struct {
	metadata   *IPSetMetadata
	hashedName string
}

func createTestSet(name string, setType SetType) *testSet {
	set := &testSet{
		metadata: &IPSetMetadata{name, setType},
	}
	set.hashedName = util.GetHashedName(set.metadata.GetPrefixName())
	return set
}

var (
	fakeSuccessCommand = testutils.TestCmd{
		Cmd:      []string{util.Ipset, util.IpsetRestoreFlag},
		Stdout:   "success",
		ExitCode: 0,
	}

	testNSSet           = createTestSet("test-ns-set", NameSpace)
	testKeyPodSet       = createTestSet("test-keyPod-set", KeyLabelOfPod)
	testKVPodSet        = createTestSet("test-kvPod-set", KeyValueLabelOfPod)
	testNamedportSet    = createTestSet("test-namedport-set", NamedPorts)
	testCIDRSet         = createTestSet("test-cidr-set", CIDRBlocks)
	testKeyNSList       = createTestSet("test-keyNS-list", KeyLabelOfNameSpace)
	testKVNSList        = createTestSet("test-kvNS-list", KeyValueLabelOfNameSpace)
	testNestedLabelList = createTestSet("test-nestedlabel-list", NestedLabelOfPod)

	toAddOrUpdateSetNames = []string{
		testNSSet.metadata.GetPrefixName(),
		testKeyPodSet.metadata.GetPrefixName(),
		testKVPodSet.metadata.GetPrefixName(),
		testNamedportSet.metadata.GetPrefixName(),
		testCIDRSet.metadata.GetPrefixName(),
		testKeyNSList.metadata.GetPrefixName(),
		testKVNSList.metadata.GetPrefixName(),
		testNestedLabelList.metadata.GetPrefixName(),
	}
)

func TestDestroyNPMIPSets(t *testing.T) {
	// TODO
}

// TODO perform tests with fexec
// commenting out stuff for list type becaues the set type isn't supported on my local machine
func TestApplyCreationsAndAdds(t *testing.T) {
	calls := []testutils.TestCmd{fakeSuccessCommand}
	iMgr := NewIPSetManager("test-node", common.NewMockIOShim(calls))

	createSetsTestHelper(t, iMgr)

	creator := iMgr.getFileCreator(1, nil, toAddOrUpdateSetNames)
	actualFileString := getSortedFileString(creator)

	lines := []string{
		fmt.Sprintf("-N %s -exist nethash", testNSSet.hashedName),
		fmt.Sprintf("-N %s -exist nethash", testKeyPodSet.hashedName),
		fmt.Sprintf("-N %s -exist nethash", testKVPodSet.hashedName),
		fmt.Sprintf("-N %s -exist hash:ip,port", testNamedportSet.hashedName),
		fmt.Sprintf("-N %s -exist nethash maxelem 4294967295", testCIDRSet.hashedName),
		fmt.Sprintf("-N %s -exist setlist", testKeyNSList.hashedName),
		fmt.Sprintf("-N %s -exist setlist", testKVNSList.hashedName),
		fmt.Sprintf("-N %s -exist setlist", testNestedLabelList.hashedName),
	}
	lines = append(lines, getSortedLines(testNSSet, "10.0.0.0", "10.0.0.1")...)
	lines = append(lines, getSortedLines(testKeyPodSet, "10.0.0.5")...)
	lines = append(lines, getSortedLines(testKVPodSet)...)
	lines = append(lines, getSortedLines(testNamedportSet)...)
	lines = append(lines, getSortedLines(testCIDRSet)...)
	lines = append(lines, getSortedLines(testKeyNSList, testNSSet.hashedName, testKeyPodSet.hashedName)...)
	lines = append(lines, getSortedLines(testKVNSList, testKVPodSet.hashedName)...)
	lines = append(lines, getSortedLines(testNestedLabelList)...)
	expectedFileString := strings.Join(lines, "\n") + "\n"

	assertEqualFileStrings(t, expectedFileString, actualFileString)
	wasFileAltered, err := creator.RunCommandOnceWithFile(util.Ipset, util.IpsetRestoreFlag)
	require.False(t, wasFileAltered)
	require.NoError(t, err)
}

func createSetsTestHelper(t *testing.T, iMgr *IPSetManager) {
	// TODO remove addReference() calls and use IPSetMode ApplyAllIPSets
	iMgr.CreateIPSet(testNSSet.metadata)
	require.NoError(t, iMgr.AddToSet([]*IPSetMetadata{testNSSet.metadata}, "10.0.0.0", "a"))
	require.NoError(t, iMgr.AddToSet([]*IPSetMetadata{testNSSet.metadata}, "10.0.0.1", "b"))
	require.NoError(t, iMgr.AddReference(testNSSet.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testKeyPodSet.metadata)
	require.NoError(t, iMgr.AddToSet([]*IPSetMetadata{testKeyPodSet.metadata}, "10.0.0.5", "c"))
	require.NoError(t, iMgr.AddReference(testKeyPodSet.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testKVPodSet.metadata)
	require.NoError(t, iMgr.AddReference(testKVPodSet.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testNamedportSet.metadata)
	require.NoError(t, iMgr.AddReference(testNamedportSet.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testCIDRSet.metadata)
	require.NoError(t, iMgr.AddReference(testCIDRSet.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testKeyNSList.metadata)
	require.NoError(t, iMgr.AddToList(testKeyNSList.metadata, []*IPSetMetadata{testNSSet.metadata, testKeyPodSet.metadata}))
	require.NoError(t, iMgr.AddReference(testKeyNSList.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testKVNSList.metadata)
	require.NoError(t, iMgr.AddToList(testKVNSList.metadata, []*IPSetMetadata{testKVPodSet.metadata}))
	require.NoError(t, iMgr.AddReference(testKVNSList.metadata.GetPrefixName(), "reference1", NetPolType))

	iMgr.CreateIPSet(testNestedLabelList.metadata)
	require.NoError(t, iMgr.AddReference(testNestedLabelList.metadata.GetPrefixName(), "reference1", NetPolType))

	assertEqualContentsTestHelper(t, toAddOrUpdateSetNames, iMgr.toAddOrUpdateCache)
}

func assertEqualContentsTestHelper(t *testing.T, setNames []string, cache map[string]struct{}) {
	require.Equal(t, len(setNames), len(cache), "cache is different than list of set names")
	for _, setName := range setNames {
		_, exists := cache[setName]
		require.True(t, exists, "cache is different than list of set names")
	}
}

// the order of adds is nondeterministic, so we're sorting them
func getSortedLines(set *testSet, members ...string) []string {
	result := []string{fmt.Sprintf("-F %s", set.hashedName)}
	adds := make([]string, len(members))
	for k, member := range members {
		adds[k] = fmt.Sprintf("-A %s %s", set.hashedName, member)
	}
	sort.Strings(adds)
	return append(result, adds...)
}

// the order of adds is nondeterministic, so we're sorting all neighboring adds
func getSortedFileString(creator *ioutil.FileCreator) string {
	lines := strings.Split(creator.ToString(), "\n")

	sortedLines := make([]string, 0)
	k := 0
	for k < len(lines) {
		line := lines[k]
		if !isAddLine(line) {
			sortedLines = append(sortedLines, line)
			k++
			continue
		}
		addLines := make([]string, 0)
		for k < len(lines) {
			line := lines[k]
			if !isAddLine(line) {
				break
			}
			addLines = append(addLines, line)
			k++
		}
		sort.Strings(addLines)
		sortedLines = append(sortedLines, addLines...)
	}
	return strings.Join(sortedLines, "\n")
}

func isAddLine(line string) bool {
	return len(line) >= 2 && line[:2] == "-A"
}

func assertEqualFileStrings(t *testing.T, expectedFileString, actualFileString string) {
	if expectedFileString == actualFileString {
		return
	}
	fmt.Println("EXPECTED FILE STRING:")
	for _, line := range strings.Split(expectedFileString, "\n") {
		fmt.Println(line)
	}
	fmt.Println("ACTUAL FILE STRING")
	for _, line := range strings.Split(actualFileString, "\n") {
		fmt.Println(line)
	}
	require.FailNow(t, "got unexpected file string (see print contents above")
}

func TestApplyDeletions(t *testing.T) {
	calls := []testutils.TestCmd{fakeSuccessCommand, fakeSuccessCommand}
	iMgr := NewIPSetManager("test-node", common.NewMockIOShim(calls))

	createSetsTestHelper(t, iMgr)
	require.NoError(t, iMgr.ApplyIPSets(""))

	// Remove members and delete others
	// TODO remove deleteReference() calls and use IPSetMode ApplyAllIPSets
	require.NoError(t, iMgr.RemoveFromSet([]*IPSetMetadata{testNSSet.metadata}, "10.0.0.1", "b"))
	require.NoError(t, iMgr.RemoveFromList(testKeyNSList.metadata, []*IPSetMetadata{testKeyPodSet.metadata}))

	require.NoError(t, iMgr.DeleteReference(testCIDRSet.metadata.GetPrefixName(), "reference1", NetPolType))
	iMgr.DeleteIPSet(testCIDRSet.metadata.GetPrefixName())

	require.NoError(t, iMgr.DeleteReference(testNestedLabelList.metadata.GetPrefixName(), "reference1", NetPolType))
	iMgr.DeleteIPSet(testNestedLabelList.metadata.GetPrefixName())

	toDeleteSetNames := []string{testCIDRSet.metadata.GetPrefixName(), testNestedLabelList.metadata.GetPrefixName()}
	assertEqualContentsTestHelper(t, toDeleteSetNames, iMgr.toDeleteCache)
	toAddOrUpdateSetNames := []string{testNSSet.metadata.GetPrefixName(), testKeyNSList.metadata.GetPrefixName()}
	assertEqualContentsTestHelper(t, toAddOrUpdateSetNames, iMgr.toAddOrUpdateCache)
	creator := iMgr.getFileCreator(1, toDeleteSetNames, toAddOrUpdateSetNames)
	actualFileString := getSortedFileString(creator)

	lines := []string{
		fmt.Sprintf("-F %s", testCIDRSet.hashedName),
		fmt.Sprintf("-F %s", testNestedLabelList.hashedName),
		fmt.Sprintf("-X %s", testCIDRSet.hashedName),
		fmt.Sprintf("-X %s", testNestedLabelList.hashedName),
		fmt.Sprintf("-N %s -exist nethash", testNSSet.hashedName),
		fmt.Sprintf("-N %s -exist setlist", testKeyNSList.hashedName),
	}
	lines = append(lines, getSortedLines(testNSSet, "10.0.0.0")...)
	lines = append(lines, getSortedLines(testKeyNSList, testNSSet.hashedName)...)
	expectedFileString := strings.Join(lines, "\n") + "\n"

	assertEqualFileStrings(t, expectedFileString, actualFileString)
	wasFileAltered, err := creator.RunCommandOnceWithFile(util.Ipset, util.IpsetRestoreFlag)
	require.False(t, wasFileAltered)
	require.NoError(t, err)
}

/*
	want to test:
	file level error handlers (if there are any)
	all line level handlers (assert the file changes as we want it to) <-- need to rework the file-creator to get the file after retry?
*/
