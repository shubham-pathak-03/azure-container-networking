package ipsets

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-container-networking/npm/util"
	testutils "github.com/Azure/azure-container-networking/test/utils"

	"github.com/stretchr/testify/require"
)

const (
	testNSSetName           = "test-ns-set"
	testKeyPodSetName       = "test-keyPod-set"
	testKVPodSetName        = "test-kvPod-set"
	testNamedportSetName    = "test-namedport-set"
	testCIDRSetName         = "test-cidr-set"
	testKeyNSListName       = "test-keyNS-list"
	testKVNSListName        = "test-kvNS-list"
	testNestedLabelListName = "test-nestedlabel-list"
)

var (
	testNSSetHashedName           = util.GetHashedName(testNSSetName)
	testKeyPodSetHashedName       = util.GetHashedName(testKeyPodSetName)
	testKVPodSetHashedName        = util.GetHashedName(testKVPodSetName)
	testNamedportSetHashedName    = util.GetHashedName(testNamedportSetName)
	testCIDRSetHashedName         = util.GetHashedName(testCIDRSetName)
	testKeyNSListHashedName       = util.GetHashedName(testKeyNSListName)
	testKVNSListHashedName        = util.GetHashedName(testKVNSListName)
	testNestedLabelListHashedName = util.GetHashedName(testNestedLabelListName)

	fakeSuccessCommand = testutils.TestCmd{
		Cmd:      []string{util.Ipset, util.IpsetRestoreFlag},
		Stdout:   "success",
		ExitCode: 0,
	}

	toAddOrUpdateSetNames = []string{
		testNSSetName,
		testKeyPodSetName,
		testKVPodSetName,
		testNamedportSetName,
		testCIDRSetName,
		testKeyNSListName,
		testKVNSListName,
		testNestedLabelListName,
	}
)

func TestDestroyNPMIPSets(t *testing.T) {
	// TODO
}

// TODO perform tests with fexec
// commenting out stuff for list type becaues the set type isn't supported on my local machine
func TestApplyCreationsAndAdds(t *testing.T) {
	calls := []testutils.TestCmd{fakeSuccessCommand}
	fexec := testutils.GetFakeExecWithScripts(calls)

	iMgr := NewIPSetManager("test-node")

	createSetsTestHelper(t, iMgr)

	// FIXME ordering of appends is random
	creator := iMgr.getFileCreator(nil, toAddOrUpdateSetNames, fexec)
	lines := []string{
		fmt.Sprintf("-N %s -exist nethash", testNSSetHashedName),
		fmt.Sprintf("-N %s -exist nethash", testKeyPodSetHashedName),
		fmt.Sprintf("-N %s -exist nethash", testKVPodSetHashedName),
		fmt.Sprintf("-N %s -exist hash:ip,port", testNamedportSetHashedName),
		fmt.Sprintf("-N %s -exist nethash maxelem 4294967295", testCIDRSetHashedName),
		fmt.Sprintf("-N %s -exist setlist", testKeyNSListHashedName),
		fmt.Sprintf("-N %s -exist setlist", testKVNSListHashedName),
		fmt.Sprintf("-N %s -exist setlist", testNestedLabelListHashedName),
		fmt.Sprintf("-F %s", testNSSetHashedName),
		fmt.Sprintf("-A %s 10.0.0.0", testNSSetHashedName),
		fmt.Sprintf("-A %s 10.0.0.1", testNSSetHashedName),
		fmt.Sprintf("-F %s", testKeyPodSetHashedName),
		fmt.Sprintf("-A %s 10.0.0.5", testKeyPodSetHashedName),
		fmt.Sprintf("-F %s", testKVPodSetHashedName),
		fmt.Sprintf("-F %s", testNamedportSetHashedName),
		fmt.Sprintf("-F %s", testCIDRSetHashedName),
		fmt.Sprintf("-F %s", testKeyNSListHashedName),
		fmt.Sprintf("-A %s %s", testKeyNSListHashedName, testNSSetHashedName),
		fmt.Sprintf("-A %s %s", testKeyNSListHashedName, testKeyPodSetHashedName),
		fmt.Sprintf("-F %s", testKVNSListHashedName),
		fmt.Sprintf("-A %s %s", testKVNSListHashedName, testKVPodSetHashedName),
		fmt.Sprintf("-F %s", testNestedLabelListHashedName),
	}
	expectedFileString := strings.Join(lines, "\n") + "\n"

	require.Equal(t, expectedFileString, creator.ToString())
	require.NoError(t, creator.RunCommandWithFile(util.Ipset, util.IpsetRestoreFlag))
	// require.NoError(t, iMgr.ApplyIPSets("")) FIXME use this instead of above

	// Test delete and destroy
	require.NoError(t, iMgr.RemoveFromSet([]string{"test-podkeylabel-set"}, "10.0.0.5", "c"))
	iMgr.DeleteIPSet("test-podkeylabel-set")
	require.NoError(t, iMgr.ApplyIPSets(""))
}

func createSetsTestHelper(t *testing.T, iMgr *IPSetManager) {
	// TODO remove addReferences and use IPSetMode ApplyAllIPSets
	iMgr.CreateIPSet(testNSSetName, NameSpace)
	require.NoError(t, iMgr.AddToSet([]string{testNSSetName}, "10.0.0.0", "a"))
	require.NoError(t, iMgr.AddToSet([]string{testNSSetName}, "10.0.0.1", "b"))
	require.NoError(t, iMgr.AddReference(testNSSetName, "reference1", NetPolType)) // FIXME remove this and all other AddReferences

	iMgr.CreateIPSet(testKeyPodSetName, KeyLabelOfPod)
	require.NoError(t, iMgr.AddToSet([]string{testKeyPodSetName}, "10.0.0.5", "c"))
	require.NoError(t, iMgr.AddReference(testKeyPodSetName, "reference1", NetPolType))

	iMgr.CreateIPSet(testKVPodSetName, KeyValueLabelOfPod)
	require.NoError(t, iMgr.AddReference(testKVPodSetName, "reference1", NetPolType))

	iMgr.CreateIPSet(testNamedportSetName, NamedPorts)
	require.NoError(t, iMgr.AddReference(testNamedportSetName, "reference1", NetPolType))

	iMgr.CreateIPSet(testCIDRSetName, CIDRBlocks)
	require.NoError(t, iMgr.AddReference(testCIDRSetName, "reference1", NetPolType))

	iMgr.CreateIPSet(testKeyNSListName, KeyLabelOfNameSpace)
	require.NoError(t, iMgr.AddToList(testKeyNSListName, []string{testNSSetName, testKeyPodSetName}))
	require.NoError(t, iMgr.AddReference(testKeyNSListName, "reference1", NetPolType))

	iMgr.CreateIPSet(testKVNSListName, KeyValueLabelOfNameSpace)
	require.NoError(t, iMgr.AddToList(testKVNSListName, []string{testKVPodSetName}))
	require.NoError(t, iMgr.AddReference(testKVNSListName, "reference1", NetPolType))

	iMgr.CreateIPSet(testNestedLabelListName, NestedLabelOfPod)
	require.NoError(t, iMgr.AddReference(testNestedLabelListName, "reference1", NetPolType))

	assertEqualContentsTestHelper(t, toAddOrUpdateSetNames, iMgr.toAddOrUpdateCache)

	iMgr.debugPrintCaches() // FIXME remove
}

func assertEqualContentsTestHelper(t *testing.T, setNames []string, cache map[string]struct{}) {
	require.Equal(t, len(setNames), len(cache), "cache is different than list of set names")
	for _, setName := range setNames {
		_, exists := cache[setName]
		require.True(t, exists, "cache is different than list of set names")
	}
}

func TestApplyDeletions(t *testing.T) {
	// calls := []testutils.TestCmd{fakeSuccessCommand}
	// fexec := testutils.GetFakeExecWithScripts(calls)

	iMgr := NewIPSetManager("test-node")

	createSetsTestHelper(t, iMgr)
	require.NoError(t, iMgr.ApplyIPSets(""))

	// delete a
}

/*
	want to test:
	file level error handlers (if there are any)
	all line level handlers (assert the file changes as we want it to) <-- need to rework the file-creator to get the file after retry?
*/
