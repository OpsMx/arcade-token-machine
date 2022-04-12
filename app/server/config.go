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
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds overall application config.
type Config struct {
	CheckIntervalMinutes int           `yaml:"checkIntervalMinutes,omitempty" json:"checkIntervalMinutes,omitempty"`
	Tokens               []TokenConfig `yaml:"tokens,omitempty" json:"tokens,omitempty"`
}

// TokenConfig holds the configuration for a specific token.
type TokenConfig struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
}

// LoadConfig will load the provided config file.
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if config.CheckIntervalMinutes <= 0 {
		config.CheckIntervalMinutes = 10
	}

	return &config, nil
}
