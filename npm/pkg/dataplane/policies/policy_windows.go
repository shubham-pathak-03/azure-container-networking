package policies

import (
	"fmt"

	"github.com/Microsoft/hcsshim/hcn"
)

var protocolNumMap = map[Protocol]string{
	TCP:  "6",
	UDP:  "17",
	ICMP: "1",
	SCTP: "132",
	// HNS thinks 256 as ANY protocol
	AnyProtocol: "256",
}

type NPMACLPolSettings struct {
	Id              string            `json:",omitempty"`
	Protocols       string            `json:",omitempty"` // EX: 6 (TCP), 17 (UDP), 1 (ICMPv4), 58 (ICMPv6), 2 (IGMP)
	Action          hcn.ActionType    `json:","`
	Direction       hcn.DirectionType `json:","`
	LocalAddresses  string            `json:",omitempty"`
	RemoteAddresses string            `json:",omitempty"`
	LocalPorts      string            `json:",omitempty"`
	RemotePorts     string            `json:",omitempty"`
	RuleType        hcn.RuleType      `json:",omitempty"`
	Priority        uint16            `json:",omitempty"`
}

func (acl ACLPolicy) convertToAclSettings() (hcn.AclPolicySetting, error) {
	policySettings := hcn.AclPolicySetting{}
	for _, setInfo := range acl.SrcList {
		if !setInfo.Included {
			return policySettings, fmt.Errorf("Windows Dataplane does not support negative matches. ACL: %+v", acl)
		}
	}

	// TODO complete this convertsion logic

	return policySettings, nil
}

func getHCNDirection(direction Direction) hcn.DirectionType {
	switch direction {
	case Ingress:
		return hcn.DirectionTypeIn
	case Egress:
		return hcn.DirectionTypeOut
	}
	return ""
}

func getHCNAction(verdict Verdict) hcn.ActionType {
	switch verdict {
	case Allowed:
		return hcn.ActionTypeAllow
	case Dropped:
		return hcn.ActionTypeBlock
	}
	return ""
}
