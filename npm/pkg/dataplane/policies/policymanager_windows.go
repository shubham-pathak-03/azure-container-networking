package policies

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Microsoft/hcsshim/hcn"
	"k8s.io/klog"
)

var (
	FailedMarshalACLSettings = errors.New("Failed to marshal ACL settings")
)

func (pMgr *PolicyManager) addPolicy(policy *NPMNetworkPolicy, endpointList []string) error {
	endpointList, ok := checkEndpointsList(policy, endpointList)
	if !ok {
		klog.Infof("[DataPlane Windows] No Endpoints to apply policy %s on", policy.Name)
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

	var tempErr error
	for _, epID := range endpointList {
		err = pMgr.addEPPolicyWithEpID(epID, epPolicyRequest)
		if err != nil {
			klog.Infof("[DataPlane Windows] Failed to add policy on %s ID Endpoint with %s err", epID, err.Error())
			// Do not return if one endpoint fails, try all endpoints.
			// aggregate the error message and return it at the end
			tempErr = fmt.Errorf("Failed to add policy on %s ID Endpoint with %s err \n Previous %w", epID, err.Error(), tempErr)
		}
	}

	return tempErr
}

func (pMgr *PolicyManager) removePolicy(name string, endpointList []string) error {
	policy, ok := pMgr.GetPolicy(name)
	if !ok {
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
	klog.Infof("[DataPlane Windows] To Remove Policy: %s \n To Delete ACLs: %+v \n To Remove From endpoints", policy.Name, rulesToRemove, endpointList)
	// TODO now there are two paths we can take.
	// If remove bug is solved we can directly remove the exact policy from the endpoint
	// but if the bug is not solved then get all existing policies and remove relevant policies from list
	// then apply remaining policies onto the endpoint
	return nil
}

// addEPPolicyWithEpID given an EP ID and a list of policies, add the policies to the endpoint
func (pMgr *PolicyManager) addEPPolicyWithEpID(epID string, policies hcn.PolicyEndpointRequest) error {
	epObj, err := pMgr.getEndpointByID(epID)
	if err != nil {
		klog.Infof("[DataPlane Windows] Skipping applying policies %s ID Endpoint with %s err", err.Error())
		return err
	}

	err = pMgr.ioShim.Hns.ApplyEndpointPolicy(epObj, hcn.RequestTypeAdd, policies)
	if err != nil {
		klog.Infof("[DataPlane Windows]Failed to apply policies on %s ID Endpoint with %s err", err.Error())
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
func getEPPolicyReqFromACLSettings(settings []NPMACLPolSettings) (hcn.PolicyEndpointRequest, error) {
	policyToAdd := hcn.PolicyEndpointRequest{
		Policies: make([]hcn.EndpointPolicy, len(settings)),
	}

	for i, acl := range settings {
		byteACL, err := json.Marshal(acl)
		if err != nil {
			klog.Infof("[DataPlane Windows] Failed to marshall ACL settings %+v", acl)
			return hcn.PolicyEndpointRequest{}, FailedMarshalACLSettings
		}

		epPolicy := hcn.EndpointPolicy{
			Type:     hcn.ACL,
			Settings: byteACL,
		}
		policyToAdd.Policies[i] = epPolicy
	}
	return policyToAdd, nil
}

func getSettingsFromACL(acls []*ACLPolicy) ([]NPMACLPolSettings, error) {
	rulesToRemove := make([]NPMACLPolSettings, len(acls))
	for i, acl := range acls {
		rule, err := acl.convertToAclSettings()
		if err != nil {
			// TODO need some retry mechanism to check why the translations failed
			return rulesToRemove, err
		}
		rulesToRemove[i] = rule
	}
	return rulesToRemove, nil
}

func getEndpointPolicyByID(id string) (hcn.AclPolicySetting, error) {
	return hcn.AclPolicySetting{}, nil
}
