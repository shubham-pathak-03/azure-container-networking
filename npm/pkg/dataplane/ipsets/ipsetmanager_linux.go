package ipsets

import (
	"fmt"

	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ioutil"
	"github.com/Azure/azure-container-networking/npm/util"
	// osexec "os/exec"
)

const (
	maxRetryCount                  = 1
	deletionPrefix                 = "delete"
	creationPrefix                 = "create"
	ipsetRestoreLineFailurePattern = "Error in line (\\d+):"
	setExistsPattern               = "Set cannot be created: set with the same name already exists"
	setDoesntExistPattern          = "The set with the given name does not exist"
	setInUseByKernelPattern        = "Set cannot be destroyed: it is in use by a kernel component"
	memberSetDoesntExist           = "Set to be added/deleted/tested as element does not exist"
)

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
	iMgr.debugPrintCaches() // FIXME remove

	creator := ioutil.NewFileCreator(ipsetRestoreLineFailurePattern, maxRetryCount)
	// creator.AddErrorToRetryOn(ioutil.NewErrorDefinition("something")) // TODO
	iMgr.handleDeletions(creator)
	iMgr.handleAddOrUpdates(creator)

	debugPrintRestoreFile(creator) // FIXME remove

	return creator.RunCommandWithFile(util.Ipset, util.IpsetRestoreFlag)
}

func (iMgr *IPSetManager) handleDeletions(creator *ioutil.FileCreator) {
	// flush all first so we don't try to delete an ipset referenced by a list we're deleting too
	// error handling:
	// - abort the flush and delete call for a set if the set doesn't exist
	// - if the set is in use by a kernel component, then skip the delete and mark it as a failure
	for setName := range iMgr.toDeleteCache {
		errorHandlers := []*ioutil.LineErrorHandler{
			{
				Definition: ioutil.NewErrorDefinition(setDoesntExistPattern),
				Method:     ioutil.AbortSection,
				Callback:   func() { log.Logf("was going to delete set %s but it doesn't exist", setName) },
			},
		}
		sectionID := getSectionID(deletionPrefix, setName)
		hashedSetName := util.GetHashedName(setName)
		creator.AddLine(sectionID, errorHandlers, util.IpsetFlushFlag, hashedSetName) // flush set
	}

	for setName := range iMgr.toDeleteCache {
		errorHandlers := []*ioutil.LineErrorHandler{
			{
				Definition: ioutil.NewErrorDefinition(setInUseByKernelPattern),
				Method:     ioutil.SkipLine,
				Callback: func() {
					log.Errorf("was going to delete set %s but it is in use by a kernel component", setName)
					// TODO mark the set as a failure and reconcile what iptables rule or ipset is referring to it
				},
			},
		}
		sectionID := getSectionID(deletionPrefix, setName)
		hashedSetName := util.GetHashedName(setName)
		creator.AddLine(sectionID, errorHandlers, util.IpsetDestroyFlag, hashedSetName) // destroy set
	}
}

func (iMgr *IPSetManager) handleAddOrUpdates(creator *ioutil.FileCreator) {
	// create all sets first
	// error handling:
	// - abort the create, flush, and add calls if create doesn't work
	for setName := range iMgr.toAddOrUpdateCache {
		set := iMgr.setMap[setName]

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

		errorHandlers := []*ioutil.LineErrorHandler{
			{
				Definition: ioutil.NewErrorDefinition(setExistsPattern),
				Method:     ioutil.AbortSection,
				Callback: func() {
					log.Errorf("was going to add/update set %s but couldn't create the set", setName)
					// TODO mark the set as a failure and handle this
				},
			},
		}
		sectionID := getSectionID(creationPrefix, setName)
		creator.AddLine(sectionID, errorHandlers, specs...) // create set
	}

	// flush and add all IPs/members for each set
	// error handling:
	// - if a member set can't be added to a list because it doesn't exist, then skip the add and mark it as a failure
	for setName := range iMgr.toAddOrUpdateCache {
		set := iMgr.setMap[setName]
		sectionID := getSectionID(creationPrefix, setName)
		creator.AddLine(sectionID, nil, util.IpsetFlushFlag, set.HashedName) // flush set (no error handler needed)

		debugPrintSetContents(set) // FIXME remove

		if set.Kind == HashSet {
			for ip := range set.IPPodKey {
				// TODO add error handler?
				creator.AddLine(sectionID, nil, util.IpsetAppendFlag, set.HashedName, ip) // add IP
			}
		} else {
			for _, member := range set.MemberIPSets {
				errorHandlers := []*ioutil.LineErrorHandler{
					{
						Definition: ioutil.NewErrorDefinition(memberSetDoesntExist),
						Method:     ioutil.SkipLine,
						Callback: func() {
							log.Errorf("was going to add member set %s to list %s, but the member doesn't exist", member.Name, setName)
							// TODO handle error
						},
					},
				}
				creator.AddLine(sectionID, errorHandlers, util.IpsetAppendFlag, set.HashedName, member.HashedName) // add member
			}
		}
	}
}

func getSectionID(prefix, setName string) string {
	return fmt.Sprintf("%s-%s", prefix, setName)
}

func (iMgr *IPSetManager) debugPrintCaches() {
	fmt.Println("DIRTY CACHE")
	fmt.Println(iMgr.toAddOrUpdateCache)
	fmt.Println("DELETE CACHE")
	fmt.Println(iMgr.toDeleteCache)
}

func debugPrintRestoreFile(creator *ioutil.FileCreator) {
	fmt.Println("RESTORE FILE")
	fmt.Println(creator.ToString())
}

func debugPrintSetContents(set *IPSet) {
	fmt.Printf("DEBUG-ME\nname: %s\nkind: %s\npodip: \n", set.Name, set.Kind)
	fmt.Println(set.IPPodKey)
	fmt.Println("members: ")
	fmt.Println(set.MemberIPSets)
	fmt.Println()
}
