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
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/skandragon/gohealthcheck/health"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

type srv struct {
	listenPort   uint16
	tokenHandler *TokenHandler
}

func (s *srv) tokenRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "application/json")
		queries := req.URL.Query()

		providerQuery := queries.Get("provider")
		if providerQuery == "" {
			http.Error(w, "provider query option not set", http.StatusUnprocessableEntity)
			return
		}

		queryParts := strings.SplitN(providerQuery, ".", 2)
		if len(queryParts) != 2 {
			http.Error(w, "provider format is wrong, should be 'filesystem.tokenname", http.StatusUnprocessableEntity)
			return
		}
		provider := queryParts[0]
		providerContext := queryParts[1]
		if provider != "filesystem" {
			http.Error(w, "provider must be 'filesystem.tokenname'", http.StatusUnprocessableEntity)
			return
		}
		if providerContext == "" {
			http.Error(w, "provider format is wrong, should be 'filesystem.tokenname", http.StatusUnprocessableEntity)
			return
		}

		var ret struct {
			Token    string
			Provider string
		}
		token, err := s.tokenHandler.GetToken(providerContext)
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		ret.Token = token
		ret.Provider = providerQuery

		json, err := json.Marshal(ret)
		if err != nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}

func runHTTPServer(ctx context.Context, listenPort uint16, healthchecker *health.Health, tokenHandler *TokenHandler) {
	s := &srv{
		listenPort:   listenPort,
		tokenHandler: tokenHandler,
	}

	r := mux.NewRouter()

	// Order matters.
	r.HandleFunc("/health", healthchecker.HTTPHandler()).Methods(http.MethodGet)

	r.HandleFunc("/tokens", s.tokenRequest()).
		Methods(http.MethodGet)

	r.Use(loggingMiddleware)
	r.Use(otelmux.Middleware("arcade-multipass"))

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.listenPort),
		Handler:      r,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
	}
	log.Fatal(srv.ListenAndServe())
}
