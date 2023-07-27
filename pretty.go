package main

import (
	"github.com/fatih/color"
	"github.com/rodaine/table"
	"regexp"
	"unicode/utf8"
)

func NewTable(header ...interface{}) table.Table {

	t := table.New(header...)
	t.WithHeaderFormatter(color.New(color.Bold).SprintfFunc())
	t.WithWidthFunc(func(s string) int {
		return utf8.RuneCountInString(stripansi(s))
	})

	return t
}

const ansire = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[a-zA-Z\\d]*)*)?\u0007)|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PRZcf-ntqry=><~]))"

var re = regexp.MustCompile(ansire)

func stripansi(str string) string {
	return re.ReplaceAllString(str, "")
}
