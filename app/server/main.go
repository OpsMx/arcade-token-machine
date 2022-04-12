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
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/skandragon/gohealthcheck/health"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const port uint16 = 1982

var (
	configFile = flag.String("configFile", "/app/config/arcade-token-machine.yaml", "Configuration file location")

	// eg, http://localhost:14268/api/traces
	jaegerEndpoint = flag.String("jaeger-endpoint", "", "Jaeger collector endpoint")

	healthchecker = health.MakeHealth()
	tracer        trace.Tracer
	tokenHandler  *TokenHandler
	config        *Config
	quitRefresher = make(chan bool)
)

func getEnvar(name string, defaultValue string) string {
	value, found := os.LookupEnv(name)
	if !found {
		return defaultValue
	}
	return value
}

func gitBranch() string {
	return getEnvar("GIT_BRANCH", "dev")
}

func gitHash() string {
	return getEnvar("GIT_HASH", "dev")
}

func showGitInfo() {
	log.Printf("GIT Version: %s @ %s", gitBranch(), gitHash())
}

func main() {
	showGitInfo()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGTERM, syscall.SIGINT)

	flag.Parse()
	if len(*jaegerEndpoint) == 0 {
		*jaegerEndpoint = getEnvar("JAEGER_TRACE_URL", "")
	}

	tracerProvider, err := newTracerProvider(*jaegerEndpoint, gitHash())
	if err != nil {
		log.Fatal(err)
	}
	otel.SetTracerProvider(tracerProvider)
	tracer = tracerProvider.Tracer("main")

	otel.SetTextMapPropagator(propagation.TraceContext{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			if *jaegerEndpoint != "" {
				log.Printf("shutting down tracer: %v", err)
			}
		}
	}(ctx)

	go healthchecker.RunCheckers(15)

	log.Printf("Loading config from %s", *configFile)
	config, err = LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Unable to load config: %v", err)
	}

	log.Printf("Starting TokenHandler")
	tokenHandler = MakeTokenHandler()
	err = tokenHandler.Start()
	if err != nil {
		log.Fatalf("Unable to start TokenHandler: %v", err)
	}
	log.Printf("Loading Tokens")
	err = tokenHandler.Reconfig(config.Tokens)
	if err != nil {
		log.Fatalf("Unable to configure TokenHandler: %v", err)
	}

	log.Printf("Starting config and token refresher")
	go startRefresher(*configFile, time.Duration(config.CheckIntervalMinutes)*time.Minute)

	log.Printf("Listening for HTTP requests on port %d", port)
	go runHTTPServer(ctx, port, healthchecker, tokenHandler)

	<-sigchan
	quitRefresher <- true

	err = tokenHandler.Stop()
	if err != nil {
		log.Printf("Unable to stop TokenHandler")
	}

	log.Printf("Exiting Cleanly")
}

func startRefresher(cf string, timerDuration time.Duration) {
	timer := time.NewTimer(timerDuration)
	for {
		select {
		case <-quitRefresher:
			timer.Stop()
			log.Printf("Config refresher stopping")
			return
		case <-timer.C:
			log.Printf("Refreshing config and tokens")
			newConfig, err := LoadConfig(cf)
			if err != nil {
				log.Printf("Unable to refresh config file: %v", err)
				timer.Reset(timerDuration)
				break
			}
			timerDuration := time.Duration(newConfig.CheckIntervalMinutes) * time.Minute
			err = tokenHandler.Reconfig(newConfig.Tokens)
			if err != nil {
				log.Printf("Unable to refresh tokens: %v", err)
				timer.Reset(timerDuration)
				break
			}
			timer.Reset(timerDuration)
		}
	}
}
