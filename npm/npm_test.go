package npm

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/Azure/azure-container-networking/npm/ipsm"
	"github.com/Azure/azure-container-networking/npm/iptm"
	"github.com/Azure/azure-container-networking/npm/metrics"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/exec"
)

// To indicate the object is needed to be DeletedFinalStateUnknown Object
type IsDeletedFinalStateUnknownObject bool

const (
	DeletedFinalStateUnknownObject IsDeletedFinalStateUnknownObject = true
	DeletedFinalStateknownObject   IsDeletedFinalStateUnknownObject = false
)

func getKey(obj interface{}, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
	if err != nil {
		t.Errorf("Unexpected error getting key for obj %v: %v", obj, err)
		return ""
	}
	return key
}

func TestMarshalJSON(t *testing.T) {
	nodeName := "nodename"
	npmCacheEncoder := NPMCacheEncoder(nodeName)
	npmCacheRaw, err := npmCacheEncoder.MarshalJSON()

	assert.NoError(t, err)

	// TODO(junguk): better to use const in NPMCache and nodeName variable
	expect := []byte(`{"ListMap":{},"NodeName":"nodename","NsMap":{},"PodMap":{},"SetMap":{}}`)
	assert.ElementsMatch(t, expect, npmCacheRaw)
}

func TestMarshalUnMarshalJSON(t *testing.T) {
	nodeName := "nodename"
	npmCacheEncoder := NPMCacheEncoder(nodeName)

	npmCacheRaw, err := npmCacheEncoder.MarshalJSON()
	assert.NoError(t, err)

	decodedNPMCache := NPMCache{}
	if err := json.Unmarshal(npmCacheRaw, &decodedNPMCache); err != nil {
		t.Errorf("failed to decode %s to NPMCache", npmCacheRaw)
	}

	expected := NPMCache{
		ListMap:  make(map[string]*ipsm.Ipset),
		NodeName: "nodename",
		NsMap:    make(map[string]*Namespace),
		PodMap:   make(map[string]*NpmPod),
		SetMap:   make(map[string]*ipsm.Ipset),
	}

	if !reflect.DeepEqual(decodedNPMCache, expected) {
		t.Errorf("got '%+v', expected '%+v'", decodedNPMCache, expected)
	}
}

func TestMain(m *testing.M) {
	metrics.InitializeAll()
	exec := exec.New()
	iptMgr := iptm.NewIptablesManager(exec, iptm.NewFakeIptOperationShim())
	iptMgr.UninitNpmChains()

	ipsMgr := ipsm.NewIpsetManager(exec)
	// Do not check returned error here to proceed all UTs.
	// TODO(jungukcho): are there any side effect?
	ipsMgr.DestroyNpmIpsets()

	exitCode := m.Run()
	os.Exit(exitCode)
}
