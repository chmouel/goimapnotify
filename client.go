package main

// This file is part of goimapnotify
// Copyright (C) 2017-2019  Jorge Javier Araya Navarro

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
	"crypto/tls"
	"fmt"
	"os"

	idle "github.com/emersion/go-imap-idle"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-sasl"
)

type IMAPIDLEClient struct {
	*client.Client
	*idle.IdleClient
}

func newClient(conf NotifyConfig) (c *client.Client, err error) {
	if conf.TLS {
		c, err = client.DialTLS(conf.Host+fmt.Sprintf(":%d", conf.Port), &tls.Config{
			ServerName:         conf.Host,
			InsecureSkipVerify: !conf.TLSOptions.RejectUnauthorized,
		})
	} else {
		c, err = client.Dial(conf.Host + fmt.Sprintf(":%d", conf.Port))
	}
	if err != nil {
		return c, err
	}

	// turn on debugging
	if conf.Debug {
		c.SetDebug(os.Stdout)
	}

	if conf.PasswordCMD != "" {
		conf = retrievePasswordCmd(conf)
	}

	if conf.GmailOauth {
		sasl_oauth := sasl.OAuthBearerOptions{
			Username: conf.Username,
			// You should use something like https://github.com/google/oauth2l
			// in your Passcmd to grab the token as Passsword
			Token: conf.Password,
			Host:  conf.Host,
			Port:  conf.Port,
		}
		sasl_client := sasl.NewOAuthBearerClient(&sasl_oauth)
		err = c.Authenticate(sasl_client)
	} else {
		err = c.Login(conf.Username, conf.Password)
	}

	return c, err
}

func newIMAPIDLEClient(conf NotifyConfig) (c *IMAPIDLEClient, err error) {
	i, err := newClient(conf)
	if err != nil {
		return c, err
	}
	return &IMAPIDLEClient{i, idle.NewClient(i)}, nil
}
