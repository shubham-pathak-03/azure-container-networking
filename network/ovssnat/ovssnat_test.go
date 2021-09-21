package ovssnat

import (
	"os"
	"testing"

	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/netlink"
	testutils "github.com/Azure/azure-container-networking/test/utils"
)

var anyInterface = "dummy"

func TestMain(m *testing.M) {
	exitCode := m.Run()

	// Create a dummy test network interface.

	os.Exit(exitCode)
}

func TestAllowInboundFromHostToNC(t *testing.T) {
	client := &OVSSnatClient{
		snatBridgeIP:          "169.254.0.1/16",
		localIP:               "169.254.0.4/16",
		containerSnatVethName: anyInterface,
		ioShim:                common.NewMockIOShim([]testutils.TestCmd{}, false, ""),
	}

	if err := client.ioShim.Netlink.AddLink(&netlink.DummyLink{
		LinkInfo: netlink.LinkInfo{
			Type: netlink.LINK_TYPE_DUMMY,
			Name: anyInterface,
		},
	}); err != nil {
		t.Errorf("Error adding dummy interface %v", err)
	}

	if err := client.ioShim.Netlink.AddLink(&netlink.DummyLink{
		LinkInfo: netlink.LinkInfo{
			Type: netlink.LINK_TYPE_DUMMY,
			Name: SnatBridgeName,
		},
	}); err != nil {
		t.Errorf("Error adding dummy interface %v", err)
	}

	if err := client.AllowInboundFromHostToNC(); err != nil {
		t.Errorf("Error adding inbound rule: %v", err)
	}

	if err := client.AllowInboundFromHostToNC(); err != nil {
		t.Errorf("Error adding existing inbound rule: %v", err)
	}

	if err := client.DeleteInboundFromHostToNC(); err != nil {
		t.Errorf("Error removing inbound rule: %v", err)
	}

	if err := client.ioShim.Netlink.DeleteLink(anyInterface); err != nil {
		t.Errorf("Error removing any interface link: %v", err)
	}
	if err := client.ioShim.Netlink.DeleteLink(SnatBridgeName); err != nil {
		t.Errorf("Error removing snat bridge: %v", err)
	}
}

func TestAllowInboundFromNCToHost(t *testing.T) {
	client := &OVSSnatClient{
		snatBridgeIP:          "169.254.0.1/16",
		localIP:               "169.254.0.4/16",
		containerSnatVethName: anyInterface,
		ioShim:                common.NewMockIOShim([]testutils.TestCmd{}, false, ""),
	}

	if err := client.ioShim.Netlink.AddLink(&netlink.DummyLink{
		LinkInfo: netlink.LinkInfo{
			Type: netlink.LINK_TYPE_DUMMY,
			Name: anyInterface,
		},
	}); err != nil {
		t.Errorf("Error adding dummy interface %v", err)
	}

	if err := client.ioShim.Netlink.AddLink(&netlink.DummyLink{
		LinkInfo: netlink.LinkInfo{
			Type: netlink.LINK_TYPE_DUMMY,
			Name: SnatBridgeName,
		},
	}); err != nil {
		t.Errorf("Error adding dummy interface %v", err)
	}

	if err := client.AllowInboundFromNCToHost(); err != nil {
		t.Errorf("Error adding inbound rule: %v", err)
	}

	if err := client.AllowInboundFromNCToHost(); err != nil {
		t.Errorf("Error adding existing inbound rule: %v", err)
	}

	if err := client.DeleteInboundFromNCToHost(); err != nil {
		t.Errorf("Error removing inbound rule: %v", err)
	}

	if err := client.ioShim.Netlink.DeleteLink(anyInterface); err != nil {
		t.Errorf("Error removing any interface link: %v", err)
	}
	if err := client.ioShim.Netlink.DeleteLink(SnatBridgeName); err != nil {
		t.Errorf("Error removing snat bridge: %v", err)
	}
}
