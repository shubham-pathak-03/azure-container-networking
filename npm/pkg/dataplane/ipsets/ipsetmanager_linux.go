package ipsets

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Azure/azure-container-networking/npm/util"
	kexec "k8s.io/utils/exec"
)

// TODO look at current errors in kusto

// TODO eventually, have multiple retries and start at spot based on line number in error?
const maxRetryCount = 1

type ipsetRestoreFile struct {
	lines      []string
	retryCount int
}

var linuxExec kexec.Interface

func assertExecExists() {
	if linuxExec == nil {
		linuxExec = kexec.New()
	}
}

// don't need networkID
func (iMgr *IPSetManager) applyIPSets(networkID string) error {
	// DEBUGGING)
	fmt.Println("DIRTY CACHE")
	fmt.Println(iMgr.dirtyCaches)
	fmt.Println("DELETE CACHE")
	fmt.Println(iMgr.deleteCache)

	file := initIPSetRestoreFile()
	file.create(iMgr)

	// MORE DEBUGGING
	fmt.Println("RESTORE FILE")
	fmt.Println(file.toString(0))

	assertExecExists() // TODO remove if using osexec
	for {
		err := file.restore(0) // TODO could retry from line that fails
		if err == nil {
			return nil
		}
		file.retryCount += 1
		if file.retryCount >= maxRetryCount {
			return fmt.Errorf("failed to restore ipets after %d tries with final error: %w", maxRetryCount, err)
		}
	}
}

func initIPSetRestoreFile() *ipsetRestoreFile {
	return &ipsetRestoreFile{
		lines:      make([]string, 0),
		retryCount: 0,
	}
}

func (file *ipsetRestoreFile) create(iMgr *IPSetManager) {
	// need to create all sets before possibly referencing them in lists
	for setName := range iMgr.dirtyCaches {
		set := iMgr.setMap[setName]
		file.createSet(set)
		// TEMP NOTES: uses ipset create --exist
		//     alternatively, could maintain a create cache, but this complicates the dirty and delete cache
	}

	for setName := range iMgr.dirtyCaches {
		set := iMgr.setMap[setName]
		file.updateMembers(set)
	}

	for _, hashedName := range iMgr.deleteCache {
		file.delete(hashedName)
	}
}

func (file *ipsetRestoreFile) createSet(set *IPSet) {
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

	file.addLine(specs...)
}

func (file *ipsetRestoreFile) updateMembers(set *IPSet) {
	file.flush(set.HashedName)

	// DEBUGGING
	fmt.Printf("DEBUG-ME\nname: %s\nkind: %s\npodip: \n", set.Name, set.Kind)
	fmt.Println(set.IPPodKey)
	fmt.Println("members: ")
	fmt.Println(set.MemberIPSets)
	fmt.Println()

	if set.Kind == HashSet {
		file.addHashSetMembers(set)
	} else {
		file.addListMembers(set)
	}
}

func (file *ipsetRestoreFile) addHashSetMembers(set *IPSet) {
	for ip := range set.IPPodKey {
		file.addLine(util.IpsetAppendFlag, set.HashedName, ip)
	}
}

func (file *ipsetRestoreFile) addListMembers(set *IPSet) {
	for _, member := range set.MemberIPSets {
		file.addLine(util.IpsetAppendFlag, set.HashedName, member.HashedName)
	}
}

func (file *ipsetRestoreFile) flush(setName string) {
	file.addLine(util.IpsetFlushFlag, setName)
}

func (file *ipsetRestoreFile) delete(setName string) {
	file.flush(setName)
	file.destroy(setName)
}

func (file *ipsetRestoreFile) destroy(setName string) {
	file.addLine(util.IpsetDestroyFlag, setName)
}

func (file *ipsetRestoreFile) addLine(items ...string) {
	spaceSeparatedItems := strings.Join(items, " ")
	file.lines = append(file.lines, spaceSeparatedItems)
}

func (file *ipsetRestoreFile) toString(lineStart int) string {
	// TODO add check that lineStart is in range?
	return strings.Join(file.lines[lineStart:], "\n") + "\n"
}

func (file *ipsetRestoreFile) getBuffer(lineStart int) *bytes.Buffer {
	return bytes.NewBufferString(file.toString(lineStart))
}

func (file *ipsetRestoreFile) restore(lineStart int) error {
	lockFile, err := getIPSetLockFile()
	if err != nil {
		return fmt.Errorf("failed to get IPSet lock file with error: %w", err)
	}
	defer closeLockFile(lockFile)

	restoreCommand := linuxExec.Command(util.Ipset, util.IpsetRestoreFlag)
	restoreCommand.SetStdin(file.getBuffer(lineStart))
	// restoreCommand.Stdin = file.getBuffer(lineStart) // if we use os/exec

	// EXTRA (for debugging exit status 1)
	stdErrBuffer := bytes.NewBuffer(nil)
	restoreCommand.SetStderr(stdErrBuffer)
	// restoreCommand.Stderr = stdErrBuffer // if we use os/exec

	// TODO would this ever be useful?
	// outBuffer := bytes.NewBuffer(nil)
	// restoreCommand.Stdout = outBuffer

	if err := restoreCommand.Start(); err != nil {
		return fmt.Errorf("failed to start ipset restore with error: %w", err)
	}
	if err := restoreCommand.Wait(); err != nil {
		return fmt.Errorf("failed to restore ipsets with error: %w: %s", err, stdErrBuffer.String())
	}

	return nil
}

func getIPSetLockFile() (*os.File, error) {
	return nil, nil // no lock file for ipset (will need for iptables equivalent)
}

func closeLockFile(file *os.File) {
	// do nothing for ipset (will need for iptables equivalent)
}
