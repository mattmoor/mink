// +build !linux

package main

import "os/exec"

// The implementation of this currently only works on Linux.
// This is a placeholder for compilation/testing.
func dropNetworking(cmd *exec.Cmd) {
	panic("only implemented on linux")
}
