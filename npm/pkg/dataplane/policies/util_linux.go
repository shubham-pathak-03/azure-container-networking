package policies

import (
	"strconv"

	"github.com/Azure/azure-container-networking/npm/util"
)

var matchTypeStrings = make(map[MatchType]string)

func initMatchTypeStrings() {
	if len(matchTypeStrings) == 0 {
		matchTypeStrings[SrcMatch] = util.IptablesSrcFlag
		matchTypeStrings[DstMatch] = util.IptablesDstFlag
		matchTypeStrings[SrcSrcMatch] = util.IptablesSrcFlag + "," + util.IptablesSrcFlag
		matchTypeStrings[DstDstMatch] = util.IptablesDstFlag + "," + util.IptablesDstFlag
		matchTypeStrings[SrcDstMatch] = util.IptablesSrcFlag + "," + util.IptablesDstFlag
		matchTypeStrings[DstSrcMatch] = util.IptablesDstFlag + "," + util.IptablesSrcFlag
	}
}

// match type is only used in Linux
func (setInfo *SetInfo) hasKnownMatchType() bool {
	initMatchTypeStrings()
	_, exists := matchTypeStrings[setInfo.MatchType]
	return exists
}

func (matchType MatchType) toIPTablesString() string {
	initMatchTypeStrings()
	return matchTypeStrings[matchType]
	// switch matchType {
	// case SrcMatch:
	// 	return util.IptablesSrcFlag
	// case DstMatch:
	// 	return util.IptablesDstFlag
	// case SrcSrcMatch:
	// 	return util.IptablesSrcFlag + "," + util.IptablesSrcFlag
	// case DstDstMatch:
	// 	return util.IptablesDstFlag + "," + util.IptablesDstFlag
	// case SrcDstMatch:
	// 	return util.IptablesSrcFlag + "," + util.IptablesDstFlag
	// case DstSrcMatch:
	// 	return util.IptablesDstFlag + "," + util.IptablesSrcFlag
	// default:
	// 	return ""
	// }
}

func (portRange *Ports) toIPTablesString() string {
	start := strconv.Itoa(int(portRange.Port))
	if portRange.Port == portRange.EndPort {
		return start
	}
	end := strconv.Itoa(int(portRange.EndPort))
	return start + ":" + end
}

func (policy *ACLPolicy) satisifiesPortAndProtocolConstraints() bool {
	return len(policy.SrcPorts) == 0 &&
		len(policy.DstPorts) == 0 &&
		policy.Protocol != AnyProtocol
}

func (target Verdict) toIPTablesString() string {
	switch target {
	case Allowed:
		return "ACCEPT"
	case Dropped:
		return "DROP"
	default:
		return ""
	}
}

func (networkPolicy *NPMNetworkPolicy) hasSamePodSelector(otherNetworkPolicy *NPMNetworkPolicy) bool {
	if len(networkPolicy.PodSelectorIPSets) != len(otherNetworkPolicy.PodSelectorIPSets) {
		return false
	}
	for k, ipset := range networkPolicy.PodSelectorIPSets {
		otherIPSet := otherNetworkPolicy.PodSelectorIPSets[k]
		if !ipset.Equals(otherIPSet) {
			return false
		}
	}
	return true
}
