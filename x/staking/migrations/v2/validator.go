package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"sigs.k8s.io/yaml"
)

// String implements the Stringer interface for a Validator object.
func (v Validator) String() string {
	bz, err := codec.ProtoMarshalJSON(&v, nil)
	if err != nil {
		panic(err)
	}

	out, err := yaml.JSONToYAML(bz)
	if err != nil {
		panic(err)
	}

	return string(out)
}

// String returns a human readable string representation of a Delegation.
func (d Delegation) String() string {
	out, _ := yaml.Marshal(d)
	return string(out)
}
