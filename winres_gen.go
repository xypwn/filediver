package main

//go:generate go tool go-winres make --product-version=git-tag
//go:generate cp rsrc_windows_*.syso ./cmd/filediver-gui
//go:generate cp rsrc_windows_*.syso ./cmd/filediver-cli
