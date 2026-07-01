//go:build windows

package main

import "syscall"

func hideFile(path string) error {
	pathW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return syscall.SetFileAttributes(pathW, syscall.FILE_ATTRIBUTE_HIDDEN)
}
