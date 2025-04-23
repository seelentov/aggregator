package models

type Rule struct {
	Condition  string `json:"condition"`
	Expression string `json:"expression"`
	Comment    string `json:"comment"`
	Target     string `json:"target"`
}
