package validate

import (
	"context"
	"encoding/json"
	"log"
	"net"

	k8sutils "github.com/Azure/azure-container-networking/test/internal/k8sutils"
	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	privilegedWindowsDaemonSetPath = "../manifests/load/privileged-daemonset-windows.yaml"
	windowsNodeSelector            = "kubernetes.io/os=windows"
)

var (
	hnsEndPointCmd   = []string{"powershell", "-c", "Get-HnsEndpoint | ConvertTo-Json"}
	azureVnetCmd     = []string{"powershell", "-c", "cat ../../k/azure-vnet.json"}
	azureVnetIpamCmd = []string{"powershell", "-c", "cat ../../k/azure-vnet-ipam.json"}
)

type WindowsValidator struct {
	clientset   *kubernetes.Clientset
	config      *rest.Config
	namespace   string
	cni         string
	restartCase bool
}

type HNSEndpoint struct {
	MacAddress       string `json:"MacAddress"`
	IPAddress        net.IP `json:"IPAddress"`
	IPv6Address      net.IP `json:",omitempty"`
	IsRemoteEndpoint bool   `json:",omitempty"`
}

type AzureVnet struct {
	NetworkInfo NetworkInfo `json:"Network"`
}

type NetworkInfo struct {
	ExternalInterfaces map[string]ExternalInterface `json:"ExternalInterfaces"`
}

type ExternalInterface struct {
	Networks map[string]Network `json:"Networks"`
}

type Network struct {
	Endpoints map[string]Endpoint `json:"Endpoints"`
}

type Endpoint struct {
	IPAddresses []net.IPNet `json:"IPAddresses"`
	IfName      string      `json:"IfName"`
}

type AzureVnetIpam struct {
	IPAM AddressSpaces `json:"IPAM"`
}

type AddressSpaces struct {
	AddrSpaces map[string]AddressSpace `json:"AddressSpaces"`
}

type AddressSpace struct {
	Pools map[string]AddressPool `json:"Pools"`
}

type AddressPool struct {
	Addresses map[string]AddressRecord `json:"Addresses"`
}

type AddressRecord struct {
	Addr  net.IP
	InUse bool
}

type check struct {
	name             string
	stateFileIps     func([]byte) (map[string]string, error)
	podLabelSelector string
	podNamespace     string
	cmd              []string
}

func CreateWindowsValidator(ctx context.Context, clienset *kubernetes.Clientset, config *rest.Config, namespace, cni string, restartCase bool) (*WindowsValidator, error) {
	// deploy privileged pod
	privilegedDaemonSet, err := k8sutils.MustParseDaemonSet(privilegedWindowsDaemonSetPath)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse daemonset")
	}
	daemonsetClient := clienset.AppsV1().DaemonSets(privilegedNamespace)
	if err := k8sutils.MustCreateDaemonset(ctx, daemonsetClient, privilegedDaemonSet); err != nil {
		return nil, errors.Wrap(err, "unable to create daemonset")
	}
	if err := k8sutils.WaitForPodsRunning(ctx, clienset, privilegedNamespace, privilegedLabelSelector); err != nil {
		return nil, errors.Wrap(err, "error while waiting for pods to be running")
	}
	return &WindowsValidator{
		clientset:   clienset,
		config:      config,
		namespace:   namespace,
		cni:         cni,
		restartCase: restartCase,
	}, nil
}

func (v *WindowsValidator) ValidateStateFile(ctx context.Context) error {
	checkSet := make(map[string][]check) // key is cni type, value is a list of check

	checkSet["cniv1"] = []check{
		{"hns", hnsStateFileIps, privilegedLabelSelector, privilegedNamespace, hnsEndPointCmd},
		{"azure-vnet", azureVnetIps, privilegedLabelSelector, privilegedNamespace, azureVnetCmd},
		{"azure-vnet-ipam", azureVnetIpamIps, privilegedLabelSelector, privilegedNamespace, azureVnetIpamCmd},
	}

	checkSet["cniv2"] = []check{
		{"azure-vnet", azureVnetIps, privilegedLabelSelector, privilegedNamespace, azureVnetCmd},
	}

	// this is checking all IPs of the pods with the statefile
	for _, check := range checkSet[v.cni] {
		err := v.validateIPs(ctx, check.stateFileIps, check.cmd, check.name, check.podNamespace, check.podLabelSelector)
		if err != nil {
			return err
		}
	}

	return nil
}

func hnsStateFileIps(result []byte) (map[string]string, error) {
	var hnsResult []HNSEndpoint
	err := json.Unmarshal(result, &hnsResult)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal hns endpoint list")
	}

	hnsPodIps := make(map[string]string)
	for _, v := range hnsResult {
		if !v.IsRemoteEndpoint {
			hnsPodIps[v.IPAddress.String()] = v.MacAddress
		}
	}
	return hnsPodIps, nil
}

func azureVnetIps(result []byte) (map[string]string, error) {
	var azureVnetResult AzureVnet
	err := json.Unmarshal(result, &azureVnetResult)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal azure vnet")
	}

	azureVnetPodIps := make(map[string]string)
	for _, v := range azureVnetResult.NetworkInfo.ExternalInterfaces {
		for _, v := range v.Networks {
			for _, e := range v.Endpoints {
				for _, v := range e.IPAddresses {
					azureVnetPodIps[v.IP.String()] = e.IfName
				}
			}
		}
	}
	return azureVnetPodIps, nil
}

func azureVnetIpamIps(result []byte) (map[string]string, error) {
	var azureVnetIpamResult AzureVnetIpam
	err := json.Unmarshal(result, &azureVnetIpamResult)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal azure vnet ipam")
	}

	azureVnetIpamPodIps := make(map[string]string)

	for _, v := range azureVnetIpamResult.IPAM.AddrSpaces {
		for _, v := range v.Pools {
			for _, v := range v.Addresses {
				if v.InUse {
					azureVnetIpamPodIps[v.Addr.String()] = v.Addr.String()
				}
			}
		}
	}
	return azureVnetIpamPodIps, nil
}

func (v *WindowsValidator) validateIPs(ctx context.Context, stateFileIps stateFileIpsFunc, cmd []string, checkType, namespace, labelSelector string) error {
	log.Println("Validating ", checkType, " state file")
	nodes, err := k8sutils.GetNodeListByLabelSelector(ctx, v.clientset, windowsNodeSelector)
	if err != nil {
		return errors.Wrapf(err, "failed to get node list")
	}
	for index := range nodes.Items {
		// get the privileged pod
		pod, err := k8sutils.GetPodsByNode(ctx, v.clientset, namespace, labelSelector, nodes.Items[index].Name)
		if err != nil {
			return errors.Wrapf(err, "failed to get privileged pod")
		}
		podName := pod.Items[0].Name
		// exec into the pod to get the state file
		result, err := k8sutils.ExecCmdOnPod(ctx, v.clientset, namespace, podName, cmd, v.config)
		if err != nil {
			return errors.Wrapf(err, "failed to exec into privileged pod")
		}
		filePodIps, err := stateFileIps(result)
		if err != nil {
			return errors.Wrapf(err, "failed to get pod ips from state file")
		}
		if len(filePodIps) == 0 && v.restartCase {
			log.Printf("No pods found on node %s", nodes.Items[index].Name)
			continue
		}
		// get the pod ips
		podIps := getPodIPsWithoutNodeIP(ctx, v.clientset, nodes.Items[index])

		check := compareIPs(filePodIps, podIps)

		if !check {
			return errors.Wrapf(errors.New("State file validation failed"), "for %s on node %s", checkType, nodes.Items[index].Name)
		}
	}
	log.Printf("State file validation for %s passed", checkType)
	return nil
}

func (v *WindowsValidator) ValidateRestartNetwork(context.Context) error {
	return errors.New("not implemented")
}