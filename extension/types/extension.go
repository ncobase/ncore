package types

// Metadata represents the metadata of an extension
type Metadata struct {
	// Name is the name of the extension
	Name string `json:"name,omitempty"`
	// Version is the version of the extension
	Version string `json:"version,omitempty"`
	// Dependencies are the dependencies of the extension
	Dependencies []string `json:"dependencies,omitempty"`
	// Description is the description of the extension
	Description string `json:"description,omitempty"`
	// Type is the type of the extension, e.g. core, business, plugin, module, etc
	Type string `json:"type,omitempty"`
	// Group is the belong group of the extension, e.g. iam, res, flow, sys, org, rt, plug, etc
	Group string `json:"group,omitempty"`
}
