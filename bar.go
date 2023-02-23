package main

import (
	"fmt"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
)

func NewBar(max int, desc string) *progressbar.ProgressBar {

	w := ansi.NewAnsiStderr()
	opts := []progressbar.Option{
		//progressbar.OptionUseANSICodes(true),
		progressbar.OptionSetWriter(w),
		progressbar.OptionSetWidth(80),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(15),
		progressbar.OptionSetDescription(fmt.Sprintf("%-60s", desc)),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetElapsedTime(false),
		progressbar.OptionSetPredictTime(false),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(w, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	}

	return progressbar.NewOptions(max, opts...)
}
