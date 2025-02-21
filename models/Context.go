package models

type Context struct {
	Name        string
	Description string
	Path        string
	Parent      string
	Children    []Context
}
