package models

type Function struct {
	Name         string
	Group        *string
	Concurrent   bool
	InputFormat  *string
	OutputFormat *string
}
