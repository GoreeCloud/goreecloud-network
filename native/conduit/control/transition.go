package control

import "errors"

const MigrationStageIsolatedValidation = "isolated-validation"

// AdvanceToIsolatedValidation applies only the source-level stage transition
// authorized by a complete isolated-acceptance decision. It never transfers
// authority, removes a compatibility bridge, or authorizes production cutover.
func AdvanceToIsolatedValidation(capability CapabilityState, evidence IsolatedAcceptanceEvidence) (CapabilityState, error) {
	decision, err := EvaluateIsolatedAcceptance(capability, evidence)
	if err != nil {
		return CapabilityState{}, err
	}
	if !decision.EligibleForIsolatedValidation {
		return CapabilityState{}, errors.New("conduit control: isolated-validation gates are incomplete")
	}
	if capability.MigrationStage != "implementation" {
		return CapabilityState{}, errors.New("conduit control: capability must be at implementation before isolated validation")
	}

	next := capability
	next.MigrationStage = MigrationStageIsolatedValidation
	next.ProductionCutoverAuthorized = false
	return next, nil
}
