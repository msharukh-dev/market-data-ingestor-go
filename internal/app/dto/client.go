package dto

type ClientConfig struct {
	Symbols map[string]SymbolConfig `json:"symbols"`
}

type SymbolConfig struct {
	RenameFields   map[string]string    `json:"rename_fields"`
	ValueRules     map[string]ValueRule `json:"value_rules"`
	OverrideFields map[string]any       `json:"override_fields"`
	RemoveFields   []string             `json:"remove_fields"`
	UseCurrentTS   bool                 `json:"use_current_ts"`
}

type ValueRule struct {
	Op    string  `json:"op"`
	Value float64 `json:"value"`
}

type ValueTransform struct {
	Operation string  `json:"operation"` // "multiply", "add", "subtract", "divide"
	Value     float64 `json:"value"`
}
