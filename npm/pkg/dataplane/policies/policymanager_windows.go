package policies

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Microsoft/hcsshim/hcn"
	"k8s.io/klog"
)

var (
	ErrFailedMarshalACLSettings   = errors.New("Failed to marshal ACL settings")
	ErrFailedUnMarshalACLSettings = errors.New("Failed to unmarshal ACL settings")
)

type endpointPolicyBuilder struct {
	aclPolicies   []*NPMACLPolSettings
	otherPolicies []hcn.EndpointPolicy
}

func (pMgr *PolicyManager) addPolicy(policy *NPMNetworkPolicy, endpointList []string) error {
	endpointList, ok := checkEndpointsList(policy, endpointList)
	if !ok {
		klog.Infof("[DataPlane Windows] No Endpoints to apply policy %s on", policy.Name)
		return nil
	}

	if len(policy.ACLs) == 0 {
		klog.Infof("[DataPlane Windows] No ACLs in policy %s to apply", policy.Name)
		return nil
	}

	rulesToAdd, err := getSettingsFromACL(policy.ACLs)
	if err != nil {
		return err
	}
	// TODO Make sure you add all allows before deny rules
	// also we will need some checks on if rules exists earlier.
	// We rely on remove policy for a clean slate reg this particular policy
	epPolicyRequest, err := getEPPolicyReqFromACLSettings(rulesToAdd)
	if err != nil {
		return err
	}

	var aggregateErr error
	for _, epID := range endpointList {
		err = pMgr.addEPPolicyWithEpID(epID, epPolicyRequest)
		if err != nil {
			klog.Infof("[DataPlane Windows] Failed to add policy on %s ID Endpoint with %s err", epID, err.Error())
			// Do not return if one endpoint fails, try all endpoints.
			// aggregate the error message and return it at the end
			aggregateErr = fmt.Errorf("Failed to add policy on %s ID Endpoint with %s err \n Previous %w", epID, err.Error(), aggregateErr)
		}
	}

	return aggregateErr
}

func (pMgr *PolicyManager) removePolicy(name string, endpointList []string) error {
	policy, ok := pMgr.GetPolicy(name)
	if !ok {
		return nil
	}

	if len(policy.ACLs) == 0 {
		klog.Infof("[DataPlane Windows] No ACLs in policy %s to remove", policy.Name)
		return nil
	}

	endpointList, ok = checkEndpointsList(policy, endpointList)
	if !ok {
		klog.Infof("[DataPlane Windows] No Endpoints to remove policy %s on", policy.Name)
		return nil
	}

	rulesToRemove, err := getSettingsFromACL(policy.ACLs)
	if err != nil {
		return err
	}
	klog.Infof("[DataPlane Windows] To Remove Policy: %s \n To Delete ACLs: %+v \n To Remove From %+v endpoints", policy.Name, rulesToRemove, endpointList)
	// If remove bug is solved we can directly remove the exact policy from the endpoint
	// but if the bug is not solved then get all existing policies and remove relevant policies from list
	// then apply remaining policies onto the endpoint
	var aggregateErr error
	for _, epID := range endpointList {
		epObj, err := pMgr.getEndpointByID(epID)
		if err != nil {
			// Do not return if one endpoint fails, try all endpoints.
			// aggregate the error message and return it at the end
			aggregateErr = fmt.Errorf("[DataPlane Windows] Skipping removing policies on %s ID Endpoint with %s err\n Previous %w", epID, err.Error(), aggregateErr)
		}
		if len(epObj.Policies) == 0 {
			klog.Infof("[DataPlanewindows] No Policies to remove on %s ID Endpoint", epID)
			continue
		}

		epBuilder, err := splitEndpointPolicies(epObj.Policies)
		if err != nil {
			aggregateErr = fmt.Errorf("[DataPlane Windows] Skipping removing policies on %s ID Endpoint with %s err\n Previous %w", epID, err.Error(), aggregateErr)
			continue
		}

		err = epBuilder.compareAndRemovePolicies(rulesToRemove)
		if err != nil {
			aggregateErr = fmt.Errorf("[DataPlane Windows] Skipping removing policies on %s ID Endpoint with %s err\n Previous %w", epID, err.Error(), aggregateErr)
			continue
		}

		epPolicies, err := epBuilder.updatePolicies()
		if err != nil {
			aggregateErr = fmt.Errorf("[DataPlanewindows] Skipping removing policies on %s ID Endpoint with %s err\n Previous %w", epID, err.Error(), aggregateErr)
			continue
		}

		err = pMgr.updateEPPolicyWithEpObj(epObj, epPolicies)
		if err != nil {
			aggregateErr = fmt.Errorf("[DataPlanewindows] Skipping removing policies on %s ID Endpoint with %s err\n Previous %w", epID, err.Error(), aggregateErr)
			continue
		}
	}

	return aggregateErr
}

// addEPPolicyWithEpID given an EP ID and a list of policies, add the policies to the endpoint
func (pMgr *PolicyManager) addEPPolicyWithEpID(epID string, policies hcn.PolicyEndpointRequest) error {
	epObj, err := pMgr.getEndpointByID(epID)
	if err != nil {
		klog.Infof("[DataPlane Windows] Skipping applying policies %s ID Endpoint with %s err", epID, err.Error())
		return err
	}

	err = pMgr.ioShim.Hns.ApplyEndpointPolicy(epObj, hcn.RequestTypeAdd, policies)
	if err != nil {
		klog.Infof("[DataPlane Windows]Failed to apply policies on %s ID Endpoint with %s err", epID, err.Error())
		return err
	}
	return nil
}

// addEPPolicyWithEpID given an EP ID and a list of policies, add the policies to the endpoint
func (pMgr *PolicyManager) updateEPPolicyWithEpObj(epObj *hcn.HostComputeEndpoint, policies hcn.PolicyEndpointRequest) error {
	err := pMgr.ioShim.Hns.ApplyEndpointPolicy(epObj, hcn.RequestTypeUpdate, policies)
	if err != nil {
		klog.Infof("[DataPlane Windows]Failed to apply policies on %s ID Endpoint with %s err", epObj.Id, err.Error())
		return err
	}
	return nil
}

func (pMgr *PolicyManager) getEndpointByID(id string) (*hcn.HostComputeEndpoint, error) {
	epObj, err := pMgr.ioShim.Hns.GetEndpointByID(id)
	if err != nil {
		klog.Infof("[DataPlane Windows] Failed to get EndPoint object of %s ID from HNS", id)
		return nil, err
	}
	return epObj, nil
}

func checkEndpointsList(policy *NPMNetworkPolicy, endpointList []string) ([]string, bool) {
	if len(endpointList) > 0 {
		return endpointList, true
	}
	if len(policy.PodEndpoints) == 0 {
		return nil, false
	}
	endpointList = make([]string, len(policy.PodEndpoints))
	i := 0
	for _, epID := range policy.PodEndpoints {
		endpointList[i] = epID
		i++
	}

	return endpointList, true
}

// getEPPolicyReqFromACLSettings converts given ACLSettings into PolicyEndpointRequest
func getEPPolicyReqFromACLSettings(settings []*NPMACLPolSettings) (hcn.PolicyEndpointRequest, error) {
	policyToAdd := hcn.PolicyEndpointRequest{
		Policies: make([]hcn.EndpointPolicy, len(settings)),
	}

	for i, acl := range settings {
		byteACL, err := json.Marshal(acl)
		if err != nil {
			klog.Infof("[DataPlane Windows] Failed to marshall ACL settings %+v", acl)
			return hcn.PolicyEndpointRequest{}, ErrFailedMarshalACLSettings
		}

		epPolicy := hcn.EndpointPolicy{
			Type:     hcn.ACL,
			Settings: byteACL,
		}
		policyToAdd.Policies[i] = epPolicy
	}
	return policyToAdd, nil
}

func getSettingsFromACL(acls []*ACLPolicy) ([]*NPMACLPolSettings, error) {
	hnsRules := make([]*NPMACLPolSettings, len(acls))
	for i, acl := range acls {
		rule, err := acl.convertToAclSettings()
		if err != nil {
			// TODO need some retry mechanism to check why the translations failed
			return hnsRules, err
		}
		hnsRules[i] = rule
	}
	return hnsRules, nil
}

// splitEndpointPolicies this function takes in endpoint policies and separated ACL policies from other policies
func splitEndpointPolicies(endpointPolicies []hcn.EndpointPolicy) (*endpointPolicyBuilder, error) {
	epBuilder := newEndpointPolicyBuilder()
	for _, policy := range endpointPolicies {
		if policy.Type == hcn.ACL {
			var aclSettings *NPMACLPolSettings
			err := json.Unmarshal(policy.Settings, &aclSettings)
			if err != nil {
				return nil, ErrFailedUnMarshalACLSettings
			}
			epBuilder.aclPolicies = append(epBuilder.aclPolicies, aclSettings)
		} else {
			epBuilder.otherPolicies = append(epBuilder.otherPolicies, policy)
		}
	}
	return epBuilder, nil
}

func newEndpointPolicyBuilder() *endpointPolicyBuilder {
	return &endpointPolicyBuilder{
		aclPolicies:   []*NPMACLPolSettings{},
		otherPolicies: []hcn.EndpointPolicy{},
	}
}

func (epBuilder *endpointPolicyBuilder) updatePolicies() (hcn.PolicyEndpointRequest, error) {
	epPolReq, err := getEPPolicyReqFromACLSettings(epBuilder.aclPolicies)
	if err != nil {
		return hcn.PolicyEndpointRequest{}, err
	}

	// Make sure other policies are applied first
	epPolReq.Policies = append(epBuilder.otherPolicies, epPolReq.Policies...)
	return epPolReq, nil
}

func (epBuilder *endpointPolicyBuilder) compareAndRemovePolicies(rulesToRemove []*NPMACLPolSettings) error {
	lenOfRulesToRemove := len(rulesToRemove)
	for _, ruleToRemove := range rulesToRemove {
		for i, acl := range epBuilder.aclPolicies {
			// First check if ID is present and equal, this saves compute cycles to compare both objects
			if ruleToRemove.Id != "" && ruleToRemove.Id == acl.Id {
				epBuilder.removeACLPolicyAtIndex(i)
				lenOfRulesToRemove--
			} else if ruleToRemove.compare(acl) {
				epBuilder.removeACLPolicyAtIndex(i)
				lenOfRulesToRemove--
			}
		}
	}
	if lenOfRulesToRemove > 0 {
		return fmt.Errorf("[Dataplane Windows] did not find %d no of ACLs to remove", lenOfRulesToRemove)
	}
	return nil
}

func (epBuilder *endpointPolicyBuilder) removeACLPolicyAtIndex(i int) {
	klog.Infof("[DataPlane Windows] Found ACL with ID %s and removing it", epBuilder.aclPolicies[i].Id)
	epBuilder.aclPolicies = append(epBuilder.aclPolicies[:i], epBuilder.aclPolicies[i+1:]...)
}
