package cache

import (
	"github.com/Azure/azure-container-networking/npm"
	"github.com/Azure/azure-container-networking/npm/ipsm"
)

type NPMCache struct {
	Nodename string
	NsMap    map[string]*npm.Namespace
	PodMap   map[string]*npm.NpmPod
	ListMap  map[string]*ipsm.Ipset
	SetMap   map[string]*ipsm.Ipset
}
