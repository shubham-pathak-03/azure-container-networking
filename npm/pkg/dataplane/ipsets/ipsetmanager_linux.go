package ipsets

import (
	"fmt"

	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ioutil"
	"github.com/Azure/azure-container-networking/npm/util"
	kexec "k8s.io/utils/exec"
	// osexec "os/exec"
)

// TODO look at current errors in kusto
// TODO eventually, have multiple retries and start at spot based on line number in error?
const maxRetryCount = 1

var linuxExec kexec.Interface

func assertExecExists() {
	if linuxExec == nil {
		linuxExec = kexec.New()
	}
}

// TODO make corresponding function in generic ipsetmanager
func destroyNPMIPSets() error {
	// called on failure or when NPM is created
	// so no ipset cache. need to use ipset list like in ipsm.go

	// create restore file that flushes all sets, then deletes all sets
	// technically don't need to flush a hashset

	return nil
}

// don't need networkID
func (iMgr *IPSetManager) applyIPSets(networkID string) error {
	// DEBUGGING)
	fmt.Println("DIRTY CACHE")
	fmt.Println(iMgr.dirtyCaches)
	fmt.Println("DELETE CACHE")
	fmt.Println(iMgr.deleteCache)

	fileCreator := createSaveFile(iMgr)

	// MORE DEBUGGING
	fmt.Println("RESTORE FILE")
	fmt.Println(fileCreator.ToString())

	assertExecExists() // TODO remove if using osexec

	// restore (retry until success)
	for {
		err := fileCreator.RunWithFile(linuxExec.Command(util.Ipset, util.IpsetRestoreFlag)) // TODO could retry from line that fails
		if err == nil {
			return nil
		}
		fileCreator.IncRetryCount()
		if fileCreator.GetRetryCount() >= maxRetryCount {
			return fmt.Errorf("failed to restore ipets after %d tries with final error: %w", maxRetryCount, err)
		}
	}
}

func createSaveFile(iMgr *IPSetManager) *ioutil.FileCreator {
	fileCreator := ioutil.InitFileCreator()

	// need to create all sets before possibly referencing them in lists
	for setName := range iMgr.dirtyCaches {
		set := iMgr.setMap[setName]
		createSet(fileCreator, set)
		// TEMP NOTES: uses ipset create --exist
		//     alternatively, could maintain a create cache, but this complicates the dirty and delete cache
	}

	for setName := range iMgr.dirtyCaches {
		set := iMgr.setMap[setName]
		updateMembers(fileCreator, set)
	}

	for _, hashedName := range iMgr.deleteCache {
		deleteSet(fileCreator, hashedName)
	}
	return fileCreator
}

func createSet(fileCreator *ioutil.FileCreator, set *IPSet) {
	methodFlag := util.IpsetNetHashFlag
	if set.Kind == ListSet {
		methodFlag = util.IpsetSetListFlag
	} else if set.Type == NamedPorts {
		methodFlag = util.IpsetIPPortHashFlag
	}

	specs := []string{util.IpsetCreationFlag, set.HashedName, util.IpsetExistFlag, methodFlag}
	if set.Type == CIDRBlocks {
		specs = append(specs, util.IpsetMaxelemName, util.IpsetMaxelemNum)
	}

	fileCreator.AddLine(specs...)
}

func updateMembers(fileCreator *ioutil.FileCreator, set *IPSet) {
	flushSet(fileCreator, set.HashedName)

	// DEBUGGING
	fmt.Printf("DEBUG-ME\nname: %s\nkind: %s\npodip: \n", set.Name, set.Kind)
	fmt.Println(set.IPPodKey)
	fmt.Println("members: ")
	fmt.Println(set.MemberIPSets)
	fmt.Println()

	if set.Kind == HashSet {
		addHashSetMembers(fileCreator, set)
	} else {
		addListMembers(fileCreator, set)
	}
}

func addHashSetMembers(fileCreator *ioutil.FileCreator, set *IPSet) {
	for ip := range set.IPPodKey {
		fileCreator.AddLine(util.IpsetAppendFlag, set.HashedName, ip)
	}
}

func addListMembers(fileCreator *ioutil.FileCreator, set *IPSet) {
	for _, member := range set.MemberIPSets {
		fileCreator.AddLine(util.IpsetAppendFlag, set.HashedName, member.HashedName)
	}
}

func flushSet(fileCreator *ioutil.FileCreator, setName string) {
	fileCreator.AddLine(util.IpsetFlushFlag, setName)
}

func deleteSet(fileCreator *ioutil.FileCreator, setName string) {
	// NOTE assume that the set is empty and isn't referenced in any list sets (should make checks in generic ipsetmanager)
	destroySet(fileCreator, setName)
}

func destroySet(fileCreator *ioutil.FileCreator, setName string) {
	fileCreator.AddLine(util.IpsetDestroyFlag, setName)
}
