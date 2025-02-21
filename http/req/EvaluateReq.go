package req

type EvaluateReq struct {
	Expression     string  `json:"expression"`
	DefaultTable   *string `json:"defaultTable"`
	DefaultContext string  `json:"defaultContext"`
}

func NewEvaluateReq(expression string) *EvaluateReq {
	return &EvaluateReq{Expression: expression, DefaultTable: nil, DefaultContext: ""}
}
