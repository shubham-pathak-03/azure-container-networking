package ipsets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testCIDRName              = "test-cidr-set"
	testPodKeyLabelName       = "test-podkeylabel-set"
	testNamespaceKeyLabelName = "test-ns-list"
)

// TODO perform tests with fexec
// commenting out stuff for list type becaues the set type isn't supported on my local machine
func TestIpsetRestore(t *testing.T) {
	iMgr := NewIPSetManager()

	require.NoError(t, iMgr.CreateIPSet(testCIDRName, CIDRBlocks))
	require.NoError(t, iMgr.CreateIPSet(testPodKeyLabelName, KeyLabelOfPod))
	// require.NoError(t, iMgr.CreateIPSet(testNamespaceKeyLabelName, KeyLabelOfNamespace))
	require.NoError(t, iMgr.AddToSet([]string{testCIDRName}, "10.0.0.0", "a"))
	require.NoError(t, iMgr.AddToSet([]string{testCIDRName}, "10.0.0.1", "b"))
	require.NoError(t, iMgr.AddToSet([]string{testPodKeyLabelName}, "10.0.0.5", "c"))
	// require.NoError(t, iMgr.AddToList(testNamespaceKeyLabelName, []string{testPodKeyLabelName}))

	require.NoError(t, iMgr.ApplyIPSets(""))

	require.NoError(t, iMgr.RemoveFromSet([]string{testPodKeyLabelName}, "10.0.0.5", "c"))
	require.NoError(t, iMgr.DeleteSet(testPodKeyLabelName))
	require.NoError(t, iMgr.ApplyIPSets(""))
}
