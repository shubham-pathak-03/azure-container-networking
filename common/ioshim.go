package common

import (
	"github.com/Azure/azure-container-networking/netlink"
	"github.com/Azure/azure-container-networking/network/netlinkinterface"
	testutils "github.com/Azure/azure-container-networking/test/utils"
	utilexec "k8s.io/utils/exec"
)

type IOShim struct {
	Exec    utilexec.Interface
	Netlink netlinkinterface.NetlinkInterface
}

func NewIOShim() *IOShim {
	return &IOShim{
		Exec:    utilexec.New(),
		Netlink: netlink.NewNetlink(),
	}
}

func NewMockIOShim(calls []testutils.TestCmd, returnError bool, errorString string) *IOShim {
	// netlink.NewMockNetlink(returnError, errorString)
	return &IOShim{
		Exec:    testutils.GetFakeExecWithScripts(calls),
		Netlink: netlink.NewNetlink(),
	}
}
