package ipsets

import (
	"fmt"

	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ioutil"
	"github.com/Azure/azure-container-networking/npm/util"
	// osexec "os/exec"
)

// TODO look at current errors in kusto
// TODO eventually, have multiple retries and start at spot based on line number in error?
const maxRetryCount = 1
const ipsetRestoreLineFailurePattern = "Error in line (\\d+):"

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
	fmt.Println(iMgr.toAddOrUpdateCache)
	fmt.Println("DELETE CACHE")
	fmt.Println(iMgr.toDeleteCache)

	fileCreator := ioutil.NewFileCreator(ipsetRestoreLineFailurePattern, maxRetryCount)
	// fileCreator.AddErrorToRetryOn(ioutil.NewErrorDefinition("something")) // TODO
	handleDeletions(iMgr, fileCreator)
	handleCreations(iMgr, fileCreator) // need to create all sets before possibly referencing them in lists
	handleMemberUpdates(iMgr, fileCreator)

	// MORE DEBUGGING
	fmt.Println("RESTORE FILE")
	fmt.Println(fileCreator.ToString())
	return fileCreator.RunCommandWithFile(util.Ipset, util.IpsetRestoreFlag)
}

func handleDeletions(iMgr *IPSetManager, fileCreator *ioutil.FileCreator) {
	// flush all first so we don't try to delete an ipset referenced by a list we're deleting too
	for setName := range iMgr.toDeleteCache {
		flushSet(fileCreator, util.GetHashedName(setName))
	}
	for setName := range iMgr.toDeleteCache {
		destroySet(fileCreator, util.GetHashedName(setName))
	}
}

func flushSet(fileCreator *ioutil.FileCreator, hashedSetName string) {
	fileCreator.AddLine(0, nil, util.IpsetFlushFlag, hashedSetName) // TODO specify section and error handler
}

func destroySet(fileCreator *ioutil.FileCreator, setName string) {
	fileCreator.AddLine(0, nil, util.IpsetDestroyFlag, setName) // TODO specify section and error handler
}

func handleCreations(iMgr *IPSetManager, fileCreator *ioutil.FileCreator) {
	for setName := range iMgr.toAddOrUpdateCache {
		set := iMgr.setMap[setName]
		createSet(fileCreator, set)
	}
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

	fileCreator.AddLine(0, nil, specs...) // TODO specify section and error handler
}

func handleMemberUpdates(iMgr *IPSetManager, fileCreator *ioutil.FileCreator) {
	for setName := range iMgr.toAddOrUpdateCache {
		set := iMgr.setMap[setName]
		updateMembers(fileCreator, set)
	}
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
		fileCreator.AddLine(0, nil, util.IpsetAppendFlag, set.HashedName, ip) // TODO specify section and error handler
	}
}

func addListMembers(fileCreator *ioutil.FileCreator, set *IPSet) {
	for _, member := range set.MemberIPSets {
		fileCreator.AddLine(0, nil, util.IpsetAppendFlag, set.HashedName, member.HashedName) // TODO specify section and error handler
	}
}
