package policies

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/npm/metrics"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ioutil"
	"github.com/Azure/azure-container-networking/npm/util"
	utilexec "k8s.io/utils/exec"
)

const (
	defaultlockWaitTimeInSeconds string = "60"
	iptablesErrDoesNotExist      int    = 1
	reconcileChainTimeInMinutes         = 5
)

var (
	iptablesAzureChainList = []string{
		util.IptablesAzureChain,
		util.IptablesAzureIngressChain,
		util.IptablesAzureEgressChain,
		util.IptablesAzureAcceptChain,
	}
	iptablesAzureDeprecatedChainList = []string{
		// NPM v1
		util.IptablesAzureIngressFromChain,
		util.IptablesAzureIngressPortChain,
		util.IptablesAzureIngressDropsChain,
		util.IptablesAzureEgressToChain,
		util.IptablesAzureEgressPortChain,
		util.IptablesAzureEgressDropsChain,
		// older
		util.IptablesAzureTargetSetsChain,
		util.IptablesAzureIngressWrongDropsChain,
	}
	iptablesOldAndNewChainList = append(iptablesAzureChainList, iptablesAzureDeprecatedChainList...)

	jumpToAzureChainArgs = []string{util.IptablesForwardChain, util.IptablesJumpFlag, util.IptablesAzureChain, util.IptablesCtstateFlag, util.IptablesNewState}

	ingressOrEgressPolicyChainPattern = fmt.Sprintf("'Chain %s-\\|Chain %s-'", util.IptablesAzureIngressPolicyChainPrefix, util.IptablesAzureEgressPolicyChainPrefix)
)

// InitializeNPMChains creates all chains/rules and makes sure the jump from FORWARD chain to
// AZURE-NPM chain is after the jumps to KUBE-FORWARD & KUBE-SERVICES chains (if they exist).
func (pMgr *PolicyManager) InitializeNPMChains() error {
	log.Logf("Initializing AZURE-NPM chains.")
	creator := pMgr.getCreatorForInitChains()
	restoreError := restore(creator)
	if restoreError != nil {
		return restoreError
	}

	// add the jump rule from FORWARD chain to AZURE-NPM chain
	if err := pMgr.positionAzureChainJumpRule(); err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to position AZURE-NPM in FORWARD chain. %s", err.Error())
	}
	return nil
}

// RemoveNPMChains removes the jump rule from FORWARD chain to AZURE-NPM chain
// and flushes and deletes all NPM Chains.
func (pMgr *PolicyManager) RemoveNPMChains() error {
	errCode, err := pMgr.runIPTablesCommand(util.IptablesDeletionFlag, jumpToAzureChainArgs...)
	if errCode != iptablesErrDoesNotExist && err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to delete AZURE-NPM from FORWARD chain")
		// FIXME update ID
		return err
	}

	// flush all chains (will create any chain, including deprecated ones, if they don't exist)
	creator, chainsToFlush := pMgr.getCreatorAndChainsForReset()
	restoreError := restore(creator)
	if restoreError != nil {
		return restoreError
	}

	err = nil
	for _, chainName := range chainsToFlush {
		errCode, err = pMgr.runIPTablesCommand(util.IptablesDestroyFlag, chainName)
		if err != nil {
			log.Logf("couldn't delete chain %s with error [%w] and exit code [%d]", chainName, err, errCode)
		}
	}

	if err != nil {
		return fmt.Errorf("couldn't delete all chains")
	}
	return nil
}

// ReconcileChains periodically creates the jump rule from FORWARD chain to AZURE-NPM chain (if it d.n.e)
// and makes sure it's after the jumps to KUBE-FORWARD & KUBE-SERVICES chains (if they exist).
func (pMgr *PolicyManager) ReconcileChains(stopChannel <-chan struct{}) {
	go pMgr.reconcileChains(stopChannel)
}

func (pMgr *PolicyManager) reconcileChains(stopChannel <-chan struct{}) {
	ticker := time.NewTicker(time.Minute * time.Duration(reconcileChainTimeInMinutes))
	defer ticker.Stop()

	for {
		select {
		case <-stopChannel:
			return
		case <-ticker.C:
			if err := pMgr.positionAzureChainJumpRule(); err != nil {
				metrics.SendErrorLogAndMetric(util.NpmID, "Error: failed to reconcile jump rule to Azure-NPM due to %s", err.Error())
			}
		}
	}
}

// this function has a direct comparison in NPM v1 iptables manager (iptm.go)
func (pMgr *PolicyManager) runIPTablesCommand(operationFlag string, args ...string) (int, error) {
	allArgs := []string{util.IptablesWaitFlag, defaultlockWaitTimeInSeconds, operationFlag}
	allArgs = append(allArgs, args...)

	if operationFlag != util.IptablesCheckFlag {
		log.Logf("Executing iptables command with args %v", allArgs)
	}

	command := pMgr.ioShim.Exec.Command(util.Iptables, allArgs...)
	output, err := command.CombinedOutput()
	if msg, failed := err.(utilexec.ExitError); failed {
		errCode := msg.ExitStatus()
		if errCode > 0 && operationFlag != util.IptablesCheckFlag {
			msgStr := strings.TrimSuffix(string(output), "\n")
			if strings.Contains(msgStr, "Chain already exists") && operationFlag == util.IptablesChainCreationFlag {
				return 0, nil
			}
			metrics.SendErrorLogAndMetric(util.IptmID, "Error: There was an error running command: [%s %v] Stderr: [%v, %s]", util.Iptables, strings.Join(args, " "), err, msgStr)
		}
		return errCode, err
	}
	return 0, nil
}

func (pMgr *PolicyManager) getCreatorForInitChains() *ioutil.FileCreator {
	creator := pMgr.getNewCreatorWithChains(iptablesAzureChainList)

	// add AZURE-NPM chain rules
	creator.AddLine("", nil, util.IptablesAppendFlag, util.IptablesAzureChain, util.IptablesJumpFlag, util.IptablesAzureIngressChain)

	ingressDropSpecs := getDropOnMatchSpecsForAzureChain(util.IptablesAzureIngressDropMarkHex)
	ingressDropSpecs = append(ingressDropSpecs, getCommentSpecs(fmt.Sprintf("DROP-ON-INGRESS-DROP-MARK-%s", util.IptablesAzureIngressDropMarkHex))...)
	creator.AddLine("", nil, ingressDropSpecs...)

	creator.AddLine("", nil, util.IptablesAppendFlag, util.IptablesAzureChain, util.IptablesJumpFlag, util.IptablesAzureEgressChain)

	egressDropSpecs := getDropOnMatchSpecsForAzureChain(util.IptablesAzureEgressDropMarkHex)
	egressDropSpecs = append(egressDropSpecs, getCommentSpecs(fmt.Sprintf("DROP-ON-EGRESS-DROP-MARK-%s", util.IptablesAzureEgressDropMarkHex))...)
	creator.AddLine("", nil, egressDropSpecs...)

	creator.AddLine("", nil, util.IptablesAppendFlag, util.IptablesAzureChain, util.IptablesJumpFlag, util.IptablesAzureAcceptChain)

	// add AZURE-NPM-ACCEPT chain rules
	clearSpecs := []string{util.IptablesAppendFlag, util.IptablesAzureChain}
	clearSpecs = append(clearSpecs, getSetMarkSpecs(util.IptablesAzureClearMarkHex)...)
	clearSpecs = append(clearSpecs, getCommentSpecs("Clear-AZURE-NPM-MARKS")...)
	creator.AddLine("", nil, clearSpecs...)

	creator.AddLine("", nil, util.IptablesAppendFlag, util.IptablesAzureAcceptChain, util.IptablesJumpFlag, util.IptablesAccept)

	creator.AddLine("", nil, util.IptablesRestoreCommit)
	return creator
}

// position AZURE-NPM chain after KUBE-FORWARD and KUBE-SERVICE chains if they exist
// this function has a direct comparison in NPM v1 iptables manager (iptm.go)
func (pMgr *PolicyManager) positionAzureChainJumpRule() error {
	kubeServicesLine, err := pMgr.getChainLineNumber(util.IptablesKubeServicesChain, util.IptablesForwardChain)
	if err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to get index of KUBE-SERVICES in FORWARD chain with error: %s", err.Error())
		return err
	}

	index := kubeServicesLine + 1

	jumpRuleReturnCode, err := pMgr.runIPTablesCommand(util.IptablesCheckFlag, jumpToAzureChainArgs...)
	if jumpRuleReturnCode != iptablesErrDoesNotExist && err != nil {
		return fmt.Errorf("couldn't check if jump to AZURE-NPM exists: %w", err)
	}
	jumpRuleExists := err != nil
	jumpRuleInsertionArgs := append([]string{strconv.Itoa(index)}, jumpToAzureChainArgs...)

	if !jumpRuleExists {
		if errCode, err := pMgr.runIPTablesCommand(util.IptablesInsertionFlag, jumpRuleInsertionArgs...); err != nil {
			metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to insert AZURE-NPM chain in FORWARD chain with error code %d.", errCode)
			// FIXME update ID
			return err
		}
		return nil
	}

	npmChainLine, err := pMgr.getChainLineNumber(util.IptablesAzureChain, util.IptablesForwardChain)
	if err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to get index of AZURE-NPM in FORWARD chain with error: %s", err.Error())
		return err
	}

	// Kube-services line number is less than npm chain line number then all good
	if kubeServicesLine < npmChainLine {
		return nil
	} else if kubeServicesLine <= 0 {
		return nil
	}

	// AZURE-NPM chain is before KUBE-SERVICES then
	// delete existing jump rule and adding it in the right order
	metrics.SendErrorLogAndMetric(util.IptmID, "Info: Reconciler deleting and re-adding AZURE-NPM in FORWARD table.")
	if errCode, err := pMgr.runIPTablesCommand(util.IptablesDeletionFlag, jumpToAzureChainArgs...); err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to delete AZURE-NPM chain from FORWARD chain with error code %d.", errCode)
		return err
	}

	// Reduce index for deleted AZURE-NPM chain
	if index > 1 {
		index--
	}
	if errCode, err := pMgr.runIPTablesCommand(util.IptablesInsertionFlag, jumpRuleInsertionArgs...); err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: after deleting, failed to insert AZURE-NPM chain in FORWARD chain with error code %d.", errCode)
		return err
	}

	return nil
}

// returns 0 if the chain d.n.e.
// this function has a direct comparison in NPM v1 iptables manager (iptm.go)
func (pMgr *PolicyManager) getChainLineNumber(chain string, parentChain string) (int, error) {
	listForwardEntriesCommand := pMgr.ioShim.Exec.Command(util.Iptables, util.IptablesTableFlag, util.IptablesFilterTable, util.IptablesNumericFlag, util.IptablesListFlag, parentChain, util.IptablesLineNumbersFlag)
	grepCommand := pMgr.ioShim.Exec.Command("grep", chain)
	pipe, err := listForwardEntriesCommand.StdoutPipe()
	if err != nil {
		return 0, err
	}
	defer pipe.Close()
	grepCommand.SetStdin(pipe)

	if err = listForwardEntriesCommand.Start(); err != nil {
		return 0, err
	}
	// Without this wait, defunct iptable child process are created
	defer listForwardEntriesCommand.Wait()

	output, err := grepCommand.CombinedOutput()
	if err != nil {
		// grep returns err status 1 if not found
		return 0, nil
	}

	if len(output) > 2 {
		lineNum, _ := strconv.Atoi(string(output[0]))
		return lineNum, nil
	}
	return 0, nil
}

// make this a function for easier testing
func (pMgr *PolicyManager) getCreatorAndChainsForReset() (*ioutil.FileCreator, []string) {
	oldPolicyChains, err := pMgr.getIngressOrEgressPolicyChainNames()
	if err != nil {
		metrics.SendErrorLogAndMetric(util.IptmID, "Error: failed to determine NPM ingress/egress policy chains to delete")
	}
	chainsToFlush := append(iptablesOldAndNewChainList, oldPolicyChains...) // will work even if oldPolicyChains is nil
	creator := pMgr.getNewCreatorWithChains(chainsToFlush)
	creator.AddLine("", nil, util.IptablesRestoreCommit)
	return creator, chainsToFlush
}

func (pMgr *PolicyManager) getIngressOrEgressPolicyChainNames() ([]string, error) {
	iptablesListCommand := pMgr.ioShim.Exec.Command(util.Iptables, util.IptablesTableFlag, util.IptablesFilterTable, util.IptablesNumericFlag, util.IptablesListFlag)
	grepCommand := pMgr.ioShim.Exec.Command("grep", ingressOrEgressPolicyChainPattern)
	pipe, err := iptablesListCommand.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer pipe.Close()
	grepCommand.SetStdin(pipe)

	if err = iptablesListCommand.Start(); err != nil {
		return nil, err
	}
	// Without this wait, defunct iptable child process are created
	defer iptablesListCommand.Wait()

	output, err := grepCommand.CombinedOutput()
	if err != nil {
		// grep returns err status 1 if not found
		return nil, nil
	}

	lines := strings.Split(string(output), "\n")
	fmt.Println(lines)
	chainNames := make([]string, len(lines))
	for k, line := range lines {
		if len(line) < 7 {
			log.Errorf("got unexpected grep output for ingress/egress chains")
		} else {
			chainNames[k] = line[6:]
		}
	}
	return chainNames, nil
}

func getDropOnMatchSpecsForAzureChain(mark string) []string {
	return []string{
		util.IptablesAppendFlag,
		util.IptablesAzureChain,
		util.IptablesJumpFlag,
		util.IptablesDrop,
		util.IptablesModuleFlag,
		util.IptablesMarkVerb,
		util.IptablesMarkFlag,
		mark,
	}
}
