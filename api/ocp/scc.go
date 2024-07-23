package ocp

import (
	"bytes"
	"fmt"
	"html/template"

	secv1 "github.com/openshift/api/security/v1"
	"gopkg.in/yaml.v2"
)

// SecurityContextConstraintsValues contains values that need replacing in the
// template.
type SecurityContextConstraintsValues struct {
	// Namespace contains the OpenShift Namespace where the SCC will be
	// used.
	Namespace string
	// Deployer refers to the Operator that creates the SCC and
	// ServiceAccounts. This is an optional option.
	Deployer string
}

// SecurityContextConstraintsDefaults can be used for generating deployment
// artifacts with details values.
var SecurityContextConstraintsDefaults = SecurityContextConstraintsValues{
	Namespace: "ceph-csi-operator-system",
	Deployer:  "",
}

// NewSecurityContextConstraints creates a new SecurityContextConstraints
// object by replacing variables in the template by the values set in the
// SecurityContextConstraintsValues.
//
// The deployer parameter (when not an empty string) is used as a prefix for
// the name of the SCC and the linked ServiceAccounts.
func NewSecurityContextConstraints(values SecurityContextConstraintsValues) (*secv1.SecurityContextConstraints, error) {
	data, err := NewSecurityContextConstraintsYAML(values)
	if err != nil {
		return nil, err
	}

	scc := &secv1.SecurityContextConstraints{}
	err = yaml.Unmarshal([]byte(data), scc)
	if err != nil {
		return nil, fmt.Errorf("failed convert YAML to %T: %w", scc, err)
	}

	return scc, nil
}

// internalSecurityContextConstraintsValues extends
// SecurityContextConstraintsValues with some private attributes that may get
// set based on other values.
type internalSecurityContextConstraintsValues struct {
	SecurityContextConstraintsValues

	// Prefix is based on SecurityContextConstraintsValues.Deployer.
	Prefix string
}

// NewSecurityContextConstraintsYAML returns a YAML string where the variables
// in the template have been replaced by the values set in the
// SecurityContextConstraintsValues.
func NewSecurityContextConstraintsYAML(values SecurityContextConstraintsValues) (string, error) {
	var buf bytes.Buffer

	// internalValues is a copy of values, but will get extended with
	// API-internal values
	internalValues := internalSecurityContextConstraintsValues{
		SecurityContextConstraintsValues: values,
	}

	if internalValues.Deployer != "" {
		internalValues.Prefix = internalValues.Deployer + "-"
	}

	tmpl, err := template.New("SCC").Parse(securityContextConstraints)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}
	err = tmpl.Execute(&buf, internalValues)
	if err != nil {
		return "", fmt.Errorf("failed to replace values in template: %w", err)
	}

	return buf.String(), nil
}
