// Copyright 2017 Microsoft. All rights reserved.
// MIT License

package platform

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/Azure/azure-container-networking/log"
	"golang.org/x/sys/windows"
)

const (
	// CNMRuntimePath is the path where CNM state files are stored.
	CNMRuntimePath = ""

	// CNIRuntimePath is the path where CNI state files are stored.
	CNIRuntimePath = ""

	// CNILockPath is the path where CNI lock files are stored.
	CNILockPath = ""

	// CNI runtime path on a Kubernetes cluster
	K8SCNIRuntimePath = "C:\\k\\azurecni\\bin"

	// Network configuration file path on a Kubernetes cluster
	K8SNetConfigPath = "C:\\k\\azurecni\\netconf"

	// CNSRuntimePath is the path where CNS state files are stored.
	CNSRuntimePath = ""

	// NPMRuntimePath is the path where NPM state files are stored.
	NPMRuntimePath = ""

	// DNCRuntimePath is the path where DNC state files are stored.
	DNCRuntimePath = ""

	// SDNRemoteArpMacAddress is the registry key for the remote arp mac address.
	// This is set for multitenancy to get arp response from within VM
	// for vlan tagged arp requests
	SDNRemoteArpMacAddress = "12-34-56-78-9a-bc"

	// Command to get SDNRemoteArpMacAddress registry key
	GetSdnRemoteArpMacAddressCommand = "(Get-ItemProperty " +
		"-Path HKLM:\\SYSTEM\\CurrentControlSet\\Services\\hns\\State -Name SDNRemoteArpMacAddress).SDNRemoteArpMacAddress"

	// Command to set SDNRemoteArpMacAddress registry key
	SetSdnRemoteArpMacAddressCommand = "Set-ItemProperty " +
		"-Path HKLM:\\SYSTEM\\CurrentControlSet\\Services\\hns\\State -Name SDNRemoteArpMacAddress -Value \"12-34-56-78-9a-bc\""

	// Command to restart HNS service
	RestartHnsServiceCommand = "Restart-Service -Name hns"
)

// Flag to check if sdnRemoteArpMacAddress registry key is set
var sdnRemoteArpMacAddressSet = false

// GetOSInfo returns OS version information.
func (Platform) GetOSInfo() string {
	return "windows"
}

func (p Platform) GetProcessSupport() error {
	cmd := fmt.Sprintf("Get-Process -Id %v", os.Getpid())
	_, err := p.ExecutePowershellCommand(cmd)
	return err
}

var tickCount = syscall.NewLazyDLL("kernel32.dll").NewProc("GetTickCount64")

// GetLastRebootTime returns the last time the system rebooted.
func  (Platform) GetLastRebootTime() (time.Time, error) {
	currentTime := time.Now()
	output, _, err := tickCount.Call()
	if errno, ok := err.(syscall.Errno); !ok || errno != 0 {
		log.Printf("Failed to call GetTickCount64, err: %v", err)
		return time.Time{}.UTC(), err
	}
	rebootTime := currentTime.Add(-time.Duration(output) * time.Millisecond).Truncate(time.Second)
	log.Printf("Formatted Boot time: %s", rebootTime.Format(time.RFC3339))
	return rebootTime.UTC(), nil
}

func (Platform) SetOutboundSNAT(subnet string) error {
	return nil
}

// ClearNetworkConfiguration clears the azure-vnet.json contents.
// This will be called only when reboot is detected - This is windows specific
func (p Platform) ClearNetworkConfiguration() (bool, error) {
	jsonStore := CNIRuntimePath + "azure-vnet.json"
	log.Printf("Deleting the json store %s", jsonStore)
	cmd := p.exec.Command("cmd", "/c", "del", jsonStore)

	if err := cmd.Run(); err != nil {
		log.Printf("Error deleting the json store %s", jsonStore)
		return true, err
	}

	return true, nil
}

func (p Platform) KillProcessByName(processName string) error {
	cmd := fmt.Sprintf("taskkill /IM %v /F", processName)
	p.ExecuteCommand(cmd)
	return nil
}

// ExecutePowershellCommand executes powershell command
func (p Platform) ExecutePowershellCommand(command string) (string, error) {
	ps, err := exec.LookPath("powershell.exe")
	if err != nil {
		return "", fmt.Errorf("Failed to find powershell executable")
	}

	log.Printf("[Azure-Utils] %s", command)

	out, err := p.exec.Command(ps, command).CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// SetSdnRemoteArpMacAddress sets the regkey for SDNRemoteArpMacAddress needed for multitenancy
func (p Platform) SetSdnRemoteArpMacAddress() error {
	if sdnRemoteArpMacAddressSet == false {
		result, err := p.ExecutePowershellCommand(GetSdnRemoteArpMacAddressCommand)
		if err != nil {
			return err
		}

		// Set the reg key if not already set or has incorrect value
		if result != SDNRemoteArpMacAddress {
			if _, err = p.ExecutePowershellCommand(SetSdnRemoteArpMacAddressCommand); err != nil {
				log.Printf("Failed to set SDNRemoteArpMacAddress due to error %s", err.Error())
				return err
			}

			log.Printf("[Azure CNS] SDNRemoteArpMacAddress regKey set successfully. Restarting hns service.")
			if _, err := p.ExecutePowershellCommand(RestartHnsServiceCommand); err != nil {
				log.Printf("Failed to Restart HNS Service due to error %s", err.Error())
				return err
			}
		}

		sdnRemoteArpMacAddressSet = true
	}

	return nil
}

func (Platform) GetOSDetails() (map[string]string, error) {
	return nil, nil
}

func (p Platform) GetProcessNameByID(pidstr string) (string, error) {
	pidstr = strings.Trim(pidstr, "\r\n")
	cmd := fmt.Sprintf("Get-Process -Id %s|Format-List", pidstr)
	out, err := p.ExecutePowershellCommand(cmd)
	if err != nil {
		log.Printf("Process is not running. Output:%v, Error %v", out, err)
		return "", err
	}

	if len(out) <= 0 {
		log.Printf("Output length is 0")
		return "", fmt.Errorf("get-process output length is 0")
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Name") {
			pName := strings.Split(line, ":")
			if len(pName) > 1 {
				return strings.TrimSpace(pName[1]), nil
			}
		}
	}

	return "", fmt.Errorf("Process not found")
}

func (Platform) PrintDependencyPackageDetails() {
}

// https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw
func (Platform) ReplaceFile(source, destination string) error {
	src, err := syscall.UTF16PtrFromString(source)
	if err != nil {
		return err
	}

	dest, err := syscall.UTF16PtrFromString(destination)
	if err != nil {
		return err
	}

	return windows.MoveFileEx(src, dest, windows.MOVEFILE_REPLACE_EXISTING|windows.MOVEFILE_WRITE_THROUGH)
}
