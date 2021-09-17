// Copyright 2018 Microsoft. All rights reserved.
// MIT License
package npm

import (
	"encoding/json"
	"time"

	"github.com/Azure/azure-container-networking/npm/ipsm"
	k8sversion "k8s.io/apimachinery/pkg/version"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	fakeexec "k8s.io/utils/exec/testing"
)

type NPMCacheKey string

const (
	NodeName NPMCacheKey = "NodeName"
	NsMap    NPMCacheKey = "NsMap"
	PodMap   NPMCacheKey = "PodMap"
	ListMap  NPMCacheKey = "ListMap"
	SetMap   NPMCacheKey = "SetMap"
)

type NPMCache struct {
	NodeName string
	NsMap    map[string]*Namespace
	PodMap   map[string]*NpmPod
	ListMap  map[string]*ipsm.Ipset
	SetMap   map[string]*ipsm.Ipset
}

// NPMCacheEncoder is used only for unit tests to test encoding and decoding NPMCache.
func NPMCacheEncoder(nodeName string) json.Marshaler {
	noResyncPeriodFunc := func() time.Duration { return 0 }
	kubeclient := k8sfake.NewSimpleClientset()
	kubeInformer := kubeinformers.NewSharedInformerFactory(kubeclient, noResyncPeriodFunc())
	fakeK8sVersion := &k8sversion.Info{
		GitVersion: "v1.20.2",
	}
	exec := &fakeexec.FakeExec{}
	npmVersion := "npm-ut-test"

	npMgr := NewNetworkPolicyManager(kubeInformer, exec, npmVersion, fakeK8sVersion)
	npMgr.NodeName = nodeName
	return npMgr
}
