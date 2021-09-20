package common

import (
	"github.com/Azure/azure-container-networking/netlink"
	"github.com/Azure/azure-container-networking/network/netlinkinterface"
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
