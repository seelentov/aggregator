package models

type Variable struct {
	Name         string
	Group        *string
	TableFormat  *string
	DefaultValue string
	Readable     bool
	Writable     bool
}
