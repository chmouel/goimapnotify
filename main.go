package main

// Execute scripts on events using IDLE imap command (Go version)
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// NotifyConfig holds the configuration
type NotifyConfig struct {
	Host       string
	Port       int  `json:",omitempty"`
	TLS        bool `json:",omitempty"`
	TLSOptions struct {
		RejectUnauthorized bool
	} `json:",omitempty"`
	Username      string
	Password      string
	OnNewMail     string
	OnNewMailPost string
	Boxes         []string
}

// IDLEEvent models an IDLE event
type IDLEEvent struct {
	Mailbox   string
	EventType string
}

// PrepareCommand parse a string and return a command executable by Go
func PrepareCommand(command string, rsp IDLEEvent) *exec.Cmd {
	var commandstr string
	if strings.Contains("%s", command) {
		commandstr = fmt.Sprintf(command, rsp.Mailbox)
	} else {
		commandstr = command
	}
	commandsplt := strings.Split(commandstr, " ")
	commandhead := commandsplt[0]
	args := commandsplt[:1]
	cmd := exec.Command(commandhead, args...)
	return cmd
}

func main() {
	// imap.DefaultLogMask = imap.LogConn | imap.LogRaw
	raw, err := ioutil.ReadFile("/home/jorge/.config/imapnotify/jorge.conf.private")
	if err != nil {
		log.Fatalf("[ERR] Can't read file: %s", err)
	}
	var conf NotifyConfig
	_ = json.Unmarshal(raw, &conf)

	events := make(chan IDLEEvent, 100)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	guard := guardian{
		mx: &sync.Mutex{},
		wg: &sync.WaitGroup{},
	}

	NewWatchMailBox(conf, events, quit, &guard)

	// Process incoming events from the mailboxes
	for rsp := range events {
		log.Printf("[DBG] Event %s for %s", rsp.EventType, rsp.Mailbox)
		if rsp.EventType == "EXPUNGE" || rsp.EventType == "EXISTS" || rsp.EventType == "RECENT" {
			cmd := PrepareCommand(conf.OnNewMail, rsp)
			err := cmd.Run()
			if err != nil {
				log.Printf("[ERR] OnNewMail command failed: %s", err)
			} else {
				// execute the post command thing
				cmd := PrepareCommand(conf.OnNewMailPost, rsp)
				err := cmd.Run()
				if err != nil {
					log.Printf("[WARN] OnNewMailPost failed: %s", err)
				}
			}
		}
	}

	log.Println("[INF] Waiting for goroutines to finish...")
	guard.Wait()
}
