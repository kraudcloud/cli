package main

import (
	"github.com/fatih/color"
	"github.com/rodaine/table"
)

func NewTable(header ...interface{}) table.Table {

	t := table.New(header...)
	t.WithHeaderFormatter(color.New(color.Bold).SprintfFunc())

	return t
}
