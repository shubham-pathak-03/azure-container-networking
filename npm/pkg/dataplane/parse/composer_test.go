package parse

import (
	"fmt"
	"testing"

	NPMIPtable "github.com/Azure/azure-container-networking/npm/pkg/dataplane/iptables"
)

var chainNames = []string{
	"INPUT",
	"FORWARD",
	"OUTPUT",
	"AZURE-NPM",
	"AZURE-NPM-ACCEPT",
	"AZURE-NPM-EGRESS",
	"AZURE-NPM-EGRESS-DROPS",
	"AZURE-NPM-EGRESS-PORT",
	"AZURE-NPM-EGRESS-TO",
	"AZURE-NPM-INGRESS",
	"AZURE-NPM-INGRESS-DROPS",
	"AZURE-NPM-INGRESS-FROM",
	"AZURE-NPM-INGRESS-PORT",
	"DOCKER",
	"DOCKER-ISOLATION-STAGE-1",
	"DOCKER-ISOLATION-STAGE-2",
	"DOCKER-USER",
	"KUBE-EXTERNAL-SERVICES",
	"KUBE-FIREWALL",
	"KUBE-FORWARD",
	"KUBE-KUBELET-CANARY",
	"KUBE-PROXY-CANARY",
	"KUBE-SERVICES",
}

func getTestChains() []*NPMIPtable.Chain {
	table, _ := IptablesFile("filter", "../testdata/iptablesave")
	// fmt.Println(table.Chains["FORWARD"].Rules[0].Protocol)
	// fmt.Println(table)

	chains := make([]*NPMIPtable.Chain, len(chainNames))
	for k, name := range chainNames {
		chain, exists := table.Chains[name]
		if !exists {
			panic(fmt.Sprintf("chain %s doesn't exist", name))
		}
		chains[k] = chain
	}
	return chains
}

func TestCompose(t *testing.T) {
	fileCreator := ChainsToSaveFile(getTestChains())
	fmt.Println(fileCreator.ToString())
}
