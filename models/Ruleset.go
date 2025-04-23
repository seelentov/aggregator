package models

type RuleSet struct {
	Name        string  `json:"name"`
	Parent      string  `json:"parent"`
	Description string  `json:"description"`
	Rules       []*Rule `json:"rules"`
}
