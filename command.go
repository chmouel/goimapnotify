// +build !windows

package main

// This file is part of goimapnotify
// Copyright (C) 2017  Jorge Javier Araya Navarro

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// PrepareCommand parse a string and return a command executable by Go
func PrepareCommand(command string, rsp IDLEEvent) *exec.Cmd {
	var commandstr string
	if strings.Contains(command, "%s") {
		commandstr = fmt.Sprintf(command, rsp.Mailbox)
	} else {
		commandstr = command
	}

	commandsplt := append([]string{"sh", "-c"}, commandstr)
	log.Printf("[DBG] Command: %s", strings.Join(commandsplt, " "))
	// #nosec
	cmd := exec.Command(commandsplt[0], commandsplt[1:]...)
	cmd.Stdout = os.Stdout
	return cmd
}
