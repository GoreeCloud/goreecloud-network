package controlplane

import (
	"fmt"
	"strings"
	"sync"
)

type Device struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	State    string `json:"state"`
}

type Resource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type AllowPolicy struct {
	ID          string `json:"id"`
	PrincipalID string `json:"principalId"`
	DeviceID    string `json:"deviceId,omitempty"`
	ResourceID  string `json:"resourceId"`
}

type AccessRequest struct {
	PrincipalID string `json:"principalId"`
	DeviceID    string `json:"deviceId,omitempty"`
	ResourceID  string `json:"resourceId"`
}

type AccessDecision struct {
	Decision   string `json:"decision"`
	ReasonCode string `json:"reasonCode"`
	Reason     string `json:"reason"`
	PolicyID   string `json:"policyId,omitempty"`
}

type Overview struct {
	DeviceCount    int    `json:"deviceCount"`
	ResourceCount  int    `json:"resourceCount"`
	PolicyCount    int    `json:"policyCount"`
	PolicyMode     string `json:"policyMode"`
	Persistence    string `json:"persistence"`
	Authentication string `json:"authentication"`
}

type State struct {
	mu        sync.RWMutex
	devices   map[string]Device
	resources map[string]Resource
	policies  map[string]AllowPolicy
}

func NewState() *State {
	return &State{
		devices:   make(map[string]Device),
		resources: make(map[string]Resource),
		policies:  make(map[string]AllowPolicy),
	}
}

func (s *State) Devices() []Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	devices := make([]Device, 0, len(s.devices))
	for _, device := range s.devices {
		devices = append(devices, device)
	}
	return devices
}

func (s *State) Overview() Overview {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Overview{
		DeviceCount:    len(s.devices),
		ResourceCount:  len(s.resources),
		PolicyCount:    len(s.policies),
		PolicyMode:     "deny_by_default",
		Persistence:    "volatile_development_memory",
		Authentication: "not_implemented",
	}
}

func (s *State) PutDevice(device Device) error {
	device.ID = strings.TrimSpace(device.ID)
	device.Name = strings.TrimSpace(device.Name)
	device.Platform = strings.TrimSpace(device.Platform)
	device.State = strings.TrimSpace(device.State)
	if device.ID == "" || device.Name == "" || device.Platform == "" || device.State == "" {
		return fmt.Errorf("device id, name, platform, and state are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[device.ID] = device
	return nil
}

func (s *State) PutResource(resource Resource) error {
	resource.ID = strings.TrimSpace(resource.ID)
	resource.Name = strings.TrimSpace(resource.Name)
	resource.Type = strings.TrimSpace(resource.Type)
	if resource.ID == "" || resource.Name == "" || resource.Type == "" {
		return fmt.Errorf("resource id, name, and type are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.resources[resource.ID] = resource
	return nil
}

func (s *State) PutAllowPolicy(policy AllowPolicy) error {
	policy.ID = strings.TrimSpace(policy.ID)
	policy.PrincipalID = strings.TrimSpace(policy.PrincipalID)
	policy.DeviceID = strings.TrimSpace(policy.DeviceID)
	policy.ResourceID = strings.TrimSpace(policy.ResourceID)
	if policy.ID == "" || policy.PrincipalID == "" || policy.ResourceID == "" {
		return fmt.Errorf("policy id, principal id, and resource id are required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[policy.ID] = policy
	return nil
}

func (s *State) EvaluateAccess(request AccessRequest) AccessDecision {
	request.PrincipalID = strings.TrimSpace(request.PrincipalID)
	request.DeviceID = strings.TrimSpace(request.DeviceID)
	request.ResourceID = strings.TrimSpace(request.ResourceID)
	if request.PrincipalID == "" || request.ResourceID == "" {
		return AccessDecision{
			Decision:   "deny",
			ReasonCode: "INVALID_CONTEXT",
			Reason:     "A principalId and resourceId are required for an access decision.",
		}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, policy := range s.policies {
		if policy.PrincipalID != request.PrincipalID || policy.ResourceID != request.ResourceID {
			continue
		}
		if policy.DeviceID != "" && policy.DeviceID != request.DeviceID {
			continue
		}
		return AccessDecision{
			Decision:   "allow",
			ReasonCode: "MATCHING_ALLOW_POLICY",
			Reason:     "An explicit matching allow policy was found.",
			PolicyID:   policy.ID,
		}
	}

	return AccessDecision{
		Decision:   "deny",
		ReasonCode: "NO_MATCHING_ALLOW_POLICY",
		Reason:     "No explicit matching allow policy exists; deny-by-default applies.",
	}
}
