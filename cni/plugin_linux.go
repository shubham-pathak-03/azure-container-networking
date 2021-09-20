package cni

import (
	"github.com/Azure/azure-container-networking/log"
	"github.com/Azure/azure-container-networking/platform"
	"k8s.io/utils/exec"
)

func removeLockFileAfterReboot(plugin *Plugin) {
	pf := platform.New(exec.New())
	rebootTime, _ := pf.GetLastRebootTime()
	log.Printf("[cni] reboot time %v", rebootTime)
}
