package policies

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/npm/util"
	testutils "github.com/Azure/azure-container-networking/test/utils"
	"github.com/stretchr/testify/require"
)

func TestInitChains(t *testing.T) {
	calls := []testutils.TestCmd{
		fakeIPTablesRestoreCommand,
		{
			Cmd:      []string{util.Iptables, util.IptablesTableFlag, util.IptablesFilterTable, util.IptablesNumericFlag, util.IptablesListFlag, util.IptablesForwardChain, util.IptablesLineNumbersFlag},
			Stdout:   "",
			ExitCode: 1, // i.e. grep finds nothing
		},
		{Cmd: []string{"grep", util.IptablesKubeServicesChain}},
		{Cmd: append([]string{util.Iptables, util.IptablesWaitFlag, defaultlockWaitTimeInSeconds, util.IptablesCheckFlag}, jumpToAzureChainArgs...)},
		{Cmd: []string{util.Iptables, util.IptablesWaitFlag, defaultlockWaitTimeInSeconds, util.IptablesInsertionFlag, "1", util.IptablesForwardChain, util.IptablesJumpFlag, util.IptablesAzureChain, util.IptablesCtstateFlag, util.IptablesNewState}},
	}
	pMgr := NewPolicyManager(common.NewMockIOShim(calls))
	err := pMgr.InitializeNPMChains()
	require.NoError(t, err)
}

func TestRemoveChains(t *testing.T) {
	calls := []testutils.TestCmd{
		{Cmd: append([]string{util.Iptables, util.IptablesWaitFlag, defaultlockWaitTimeInSeconds, util.IptablesDeletionFlag}, jumpToAzureChainArgs...)},
		{
			Cmd: []string{util.Iptables, util.IptablesTableFlag, util.IptablesFilterTable, util.IptablesNumericFlag, util.IptablesListFlag},
			Stdout: strings.Join(
				[]string{
					fmt.Sprintf("Chain %s-123456", util.IptablesAzureIngressPolicyChainPrefix),
					fmt.Sprintf("Chain %s-123456", util.IptablesAzureEgressPolicyChainPrefix),
				},
				"\n",
			),
		},
		{Cmd: []string{"grep", ingressOrEgressPolicyChainPattern}},
		fakeIPTablesRestoreCommand,
	}
	for _, chain := range iptablesOldAndNewChainList {
		calls = append(calls, getFakeDestroyCommand(chain))
	}
	calls = append(calls, getFakeDestroyCommand(fmt.Sprintf("%s-123456", util.IptablesAzureIngressPolicyChainPrefix)))
	calls = append(calls, getFakeDestroyCommand(fmt.Sprintf("%s-123456", util.IptablesAzureEgressPolicyChainPrefix)))
	pMgr := NewPolicyManager(common.NewMockIOShim(calls))
	err := pMgr.RemoveNPMChains()
	require.NoError(t, err)
}

func getFakeDestroyCommand(chain string) testutils.TestCmd {
	return testutils.TestCmd{Cmd: []string{util.Iptables, util.IptablesWaitFlag, defaultlockWaitTimeInSeconds, util.IptablesDestroyFlag, chain}}
}

func TestGetChainLineNumber(t *testing.T) {
	// TODO (see iptm_test.go)
}
