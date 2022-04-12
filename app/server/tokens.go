/*
 * Copyright 2022 OpsMx, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License")
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type updateMessage struct {
	action string
	name   string
	token  string
}

type tokenResponse struct {
	token string
}

type tokenRequest struct {
	name         string
	responseChan chan tokenResponse
}

// TokenHandler is the magic token handler.
type TokenHandler struct {
	tokens      map[string]string
	tokenNames  []string
	updateChan  chan updateMessage
	requestChan chan tokenRequest
	quit        chan bool
	running     bool
}

// MakeTokenHandler will return a new TokenHandler, ready to have Start() called.
func MakeTokenHandler() *TokenHandler {
	return &TokenHandler{
		tokens:      make(map[string]string),
		updateChan:  make(chan updateMessage),
		requestChan: make(chan tokenRequest),
		quit:        make(chan bool),
	}
}

// Reconfig will reload the list of tokens, returning an error if any
// cannot be loaded or the handler is not running.  This should
// be called periodically to refresh tokens as needed.
func (th *TokenHandler) Reconfig(config []TokenConfig) error {
	seen := make(map[string]bool, len(th.tokenNames))
	for _, name := range th.tokenNames {
		seen[name] = false
	}

	for _, t := range config {
		data, err := os.ReadFile(t.Path)
		if err != nil {
			return err
		}
		token := strings.TrimSpace(string(data))
		err = th.UpdateToken(t.Name, token)
		if err != nil {
			return err
		}
		seen[t.Name] = true
	}

	for name, v := range seen {
		if !v {
			err := th.DeleteToken(name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateToken will update the contents of the token with new data.
func (th *TokenHandler) UpdateToken(name string, token string) error {
	if !th.running {
		return fmt.Errorf("TokenHandler not running")
	}
	th.updateChan <- updateMessage{
		action: "update",
		name:   name,
		token:  token,
	}
	return nil
}

// DeleteToken will safely delete a token that is no longer needed.
func (th *TokenHandler) DeleteToken(name string) error {
	if !th.running {
		return fmt.Errorf("TokenHandler not running")
	}
	th.updateChan <- updateMessage{action: "delete", name: name}
	return nil
}

// GetToken will return the found token, or an error if not present.
func (th *TokenHandler) GetToken(name string) (string, error) {
	if !th.running {
		return "", fmt.Errorf("TokenHandler not running")
	}
	c := make(chan tokenResponse)
	th.requestChan <- tokenRequest{
		name:         name,
		responseChan: c,
	}
	resp := <-c
	if resp.token == "" {
		return "", fmt.Errorf("unknown token name %s", name)
	}
	return resp.token, nil
}

// Start fires up a periodic refresher, and
func (th *TokenHandler) Start() error {
	if th.running {
		return fmt.Errorf("TokenHandler already running")
	}
	th.running = true
	go th.updater()
	return nil
}

// Stop turns off the reading system.
func (th *TokenHandler) Stop() error {
	if !th.running {
		return fmt.Errorf("TokenHandler not running")
	}
	th.quit <- true
	th.running = false
	return nil
}

func (th *TokenHandler) updater() {
	for {
		select {
		case <-th.quit:
			log.Printf("TokenHandler exiting")
			return
		case update := <-th.updateChan:
			switch update.action {
			case "delete":
				delete(th.tokens, update.name)
			case "update":
				th.tokens[update.name] = update.token
			default:
				log.Printf("TokenUpdater: unknown action %s, ignoring", update.action)
			}
		case req := <-th.requestChan:
			req.responseChan <- tokenResponse{token: th.tokens[req.name]}
		}
	}
}
