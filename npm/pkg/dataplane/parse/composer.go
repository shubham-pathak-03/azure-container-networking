package parse

import (
	"github.com/Azure/azure-container-networking/common"
	"github.com/Azure/azure-container-networking/npm/pkg/dataplane/ioutil"
	NPMIPtable "github.com/Azure/azure-container-networking/npm/pkg/dataplane/iptables"
)

// NOTE doesn't specify ACCEPT/DROP for any chain specification e.g. for FORWARD
// NOTE also doesn't support flags that parser ignores e.g. -o, -i, -s, -d
func ChainsToSaveFile(chains []*NPMIPtable.Chain) *ioutil.FileCreator {
	fileCreator := ioutil.NewFileCreator(common.NewMockIOShim(nil), 1)

	// fileCreator.AddLine("*filter") // specify table

	// // specify chains modified
	// for _, chain := range chains {
	// 	fileCreator.AddLine(":"+chain.Name, "-", "[0:0]") // TODO get original count from iptables-save
	// }

	// // add rules
	// for _, chain := range chains {
	// 	for _, rule := range chain.Rules {
	// 		specs := []string{"-A", chain.Name}

	// 		if rule.Protocol != "" {
	// 			specs = append(specs, util.IptablesProtFlag, rule.Protocol)
	// 		}
	// 		if rule.Modules != nil {
	// 			for _, module := range rule.Modules {
	// 				specs = append(specs, util.IptablesModuleFlag, module.Verb) // is there always a verb??
	// 				specs = append(specs, getOptionValueStrings(module.OptionValueMap)...)
	// 			}
	// 		}
	// 		if rule.Target != nil { // there should always be a target though...
	// 			specs = append(specs, util.IptablesJumpFlag, rule.Target.Name)
	// 			specs = append(specs, getOptionValueStrings(rule.Target.OptionValueMap)...)
	// 		}

	// 		fileCreator.AddLine(specs...)
	// 	}
	// }
	return fileCreator
}

// can have module options starting with not-X (meaning "! --X"), can also have module options with no values
func getOptionValueStrings(optionValues map[string][]string) []string {
	result := make([]string, 0)
	for option, values := range optionValues { // values can be an empty slice if it's a lone option
		if len(option) > 4 && option[:4] == "not-" { // TODO make this a constant
			result = append(result, "!")
			option = option[4:]
		}
		result = append(result, "--"+option)
		result = append(result, values...)
	}
	return result
}
