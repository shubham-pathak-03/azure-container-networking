package cache

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/Azure/azure-container-networking/npm"
	"github.com/Azure/azure-container-networking/npm/ipsm"
	"github.com/stretchr/testify/assert"
	k8sversion "k8s.io/apimachinery/pkg/version"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	fakeexec "k8s.io/utils/exec/testing"
)

func NPMEncoder(nodeName string) *npm.NetworkPolicyManager {
	noResyncPeriodFunc := func() time.Duration { return 0 }
	kubeclient := k8sfake.NewSimpleClientset()
	kubeInformer := kubeinformers.NewSharedInformerFactory(kubeclient, noResyncPeriodFunc())
	fakeK8sVersion := &k8sversion.Info{
		GitVersion: "v1.20.2",
	}
	exec := &fakeexec.FakeExec{}
	npmVersion := "npm-ut-test"

	npMgr := npm.NewNetworkPolicyManager(kubeInformer, exec, npmVersion, fakeK8sVersion)
	npMgr.NodeName = nodeName

	return npMgr
}

func TestDecode(t *testing.T) {
	encodedNPMCacheData := []byte(`{"ListMap":{},"Nodename":"abc","NsMap":{},"PodMap":{},"SetMap":{}}`)
	decodedNPMCache := NPMCache{}
	if err := json.Unmarshal(encodedNPMCacheData, &decodedNPMCache); err != nil {
		t.Errorf("failed to decode %s to NPMCache", encodedNPMCacheData)
	}

	expected := NPMCache{
		ListMap:  make(map[string]*ipsm.Ipset),
		Nodename: "abc",
		NsMap:    make(map[string]*npm.Namespace),
		PodMap:   make(map[string]*npm.NpmPod),
		SetMap:   make(map[string]*ipsm.Ipset),
	}

	if !reflect.DeepEqual(decodedNPMCache, expected) {
		t.Errorf("got '%+v', expected '%+v'", decodedNPMCache, expected)
	}
}

func TestEncode(t *testing.T) {
	expect := []byte(`{"ListMap":{},"Nodename":"abc","NsMap":{},"PodMap":{},"SetMap":{}}`)
	nodeName := "abc"
	npmEncoder := NPMEncoder(nodeName)
	npmCacheRaw, err := json.Marshal(npmEncoder)
	assert.NoError(t, err)
	assert.ElementsMatch(t, expect, npmCacheRaw)
}

func TestEncodeDecode(t *testing.T) {
	npmEncoder := NPMEncoder("abc")
	npmCacheRaw, err := json.Marshal(npmEncoder)
	assert.NoError(t, err)

	decodedNPMCache := NPMCache{}
	if err := json.Unmarshal(npmCacheRaw, &decodedNPMCache); err != nil {
		t.Errorf("failed to decode %s to NPMCache", npmCacheRaw)
	}

	expected := NPMCache{
		ListMap:  make(map[string]*ipsm.Ipset),
		Nodename: "abc",
		NsMap:    make(map[string]*npm.Namespace),
		PodMap:   make(map[string]*npm.NpmPod),
		SetMap:   make(map[string]*ipsm.Ipset),
	}

	if !reflect.DeepEqual(decodedNPMCache, expected) {
		t.Errorf("got '%+v', expected '%+v'", decodedNPMCache, expected)
	}
}
