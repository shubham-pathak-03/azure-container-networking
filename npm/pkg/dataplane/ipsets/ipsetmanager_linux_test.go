package ipsets

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func getCIDRipset() *IPSet {
	return NewIPSet("test-cidr-set", CIDRBlocks)
}

func getPodKeyLabelSet() *IPSet {
	return NewIPSet("test-podkeylabel-set", KeyLabelOfPod)
}

func getNamespaceKeyLabelList() *IPSet {
	return NewIPSet("test-ns-list", KeyLabelOfNameSpace)
}

// TODO perform tests with fexec
// commenting out stuff for list type becaues the set type isn't supported on my local machine
func TestIpsetRestore(t *testing.T) {
	iMgr := NewIPSetManager()

	cidrIPSet := getCIDRipset()
	podKeyLabelSet := getPodKeyLabelSet()
	// namespaceList := getNamespaceKeyLabelList()

	require.NoError(t, iMgr.CreateIPSet(cidrIPSet))
	require.NoError(t, iMgr.CreateIPSet(podKeyLabelSet))
	// require.NoError(t, iMgr.CreateIPSet(namespaceList))
	require.NoError(t, iMgr.AddToSet([]*IPSet{cidrIPSet}, "10.0.0.0", "a"))
	require.NoError(t, iMgr.AddToSet([]*IPSet{cidrIPSet}, "10.0.0.1", "b"))
	require.NoError(t, iMgr.AddToSet([]*IPSet{podKeyLabelSet}, "10.0.0.5", "c"))
	// require.NoError(t, iMgr.AddToList(namespaceList.Name, []string{podKeyLabelSet.Name}))

	require.NoError(t, iMgr.ApplyIPSets(""))

	require.NoError(t, iMgr.RemoveFromSet([]string{podKeyLabelSet.Name}, "10.0.0.5", "c"))
	require.NoError(t, iMgr.DeleteSet(podKeyLabelSet.Name))
	require.NoError(t, iMgr.ApplyIPSets(""))
}
