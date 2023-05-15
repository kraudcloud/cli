//go:build darwin
// +build darwin

package main

// We have to use FileDir because MacOS doesn't support keyring (we need to cross compile cgo and I'm hitting a linker wall)
// https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/FileSystemOverview/FileSystemOverview.html
var FileDir = "~/Library/Application Support/kraudcloud"
