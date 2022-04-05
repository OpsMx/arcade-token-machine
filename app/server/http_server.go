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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/skandragon/gohealthcheck/health"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

type srv struct {
	listenPort uint16
}

func (*srv) tokenRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("content-type", "application/json")
		// TODO: implement actual responses
		ret := false
		json, err := json.Marshal(ret)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(json)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.LoggingHandler(os.Stdout, next)
}

func (s *srv) routes(r *mux.Router) {
	// internal handlers
	r.HandleFunc("/tokens", s.tokenRequest()).Methods(http.MethodGet)
}

func runHTTPServer(ctx context.Context, listenPort uint16, healthchecker *health.Health) {
	s := &srv{
		listenPort: listenPort,
	}

	r := mux.NewRouter()
	// added first because order matters.
	r.HandleFunc("/health", healthchecker.HTTPHandler()).Methods(http.MethodGet)
	s.routes(r)

	r.Use(loggingMiddleware)
	r.Use(otelmux.Middleware("arcade-multipass"))

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.listenPort),
		Handler: r,
		// Disable HTTP/2.
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
	log.Fatal(srv.ListenAndServe())
}
