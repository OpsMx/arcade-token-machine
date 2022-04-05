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
	"time"

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
		queries := req.URL.Query()
		provider := queries.Get("provider")
		if provider != "multipass" {
			http.Error(w, "provider must be 'multipass'", http.StatusUnprocessableEntity)
			return
		}
		providerContext := queries.Get("providerContext")
		// Get an actual token, for now...  echo the ENVAR...
		var ret struct {
			Token    string
			Context  string
			Provider string
		}
		ret.Token = getEnvar("KUBERNETES_TOKEN", "KUBERNETES_TOKEN-envar-not-set")
		ret.Context = providerContext
		ret.Provider = provider

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

func runHTTPServer(ctx context.Context, listenPort uint16, healthchecker *health.Health) {
	s := &srv{
		listenPort: listenPort,
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
