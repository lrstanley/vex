// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var config = newConfig()

type Config struct {
	mu sync.RWMutex

	UnsealKeys              []string          `json:"unseal_keys"`
	RootToken               string            `json:"root_token"`
	BootstrappedAuthEngines bool              `json:"bootstrapped_auth_engines"`
	BootstrappedPolicies    bool              `json:"bootstrapped_policies"`
	BootstrappedUsers       bool              `json:"bootstrapped_users"`
	BootstrappedMounts      bool              `json:"bootstrapped_mounts"`
	BootstrappedKVSecrets   bool              `json:"bootstrapped_kv_secrets"`
	Users                   map[string]string `json:"users"`
}

func (c *Config) SetUser(username, password string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Users[username] = password
}

func newConfig() *Config {
	return &Config{
		UnsealKeys: []string{},
		Users:      make(map[string]string),
	}
}

func ConfigPath() string {
	dirRaw, err := exec.CommandContext(
		context.Background(),
		"git",
		"rev-parse",
		"--show-toplevel",
	).Output()
	if err != nil {
		logger.Error("failed to get git directory", "error", err)
		os.Exit(1)
	}

	return filepath.Join(strings.TrimSpace(string(dirRaw)), ".mock-cluster.json")
}

func LoadConfig() {
	cfgPath := ConfigPath()
	logger.Info("loading config", "path", cfgPath)

	_, err := os.Stat(cfgPath)
	if err != nil && os.IsNotExist(err) {
		return
	}

	f, err := os.Open(cfgPath)
	if err != nil {
		logger.Error("failed to open config file", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	err = json.NewDecoder(f).Decode(config)
	_ = f.Close()
	if err != nil {
		logger.Error("failed to decode config file", "path", cfgPath, "error", err)
		os.Exit(1)
	}
}

func SaveConfig() {
	cfgPath := ConfigPath()
	logger.Info("saving config", "path", cfgPath)

	cfgData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		logger.Error("failed to marshal config", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	err = os.WriteFile(cfgPath, cfgData, 0o600)
	if err != nil {
		logger.Error("failed to write config file", "path", cfgPath, "error", err)
		os.Exit(1)
	}
}

func RemoveConfig() {
	cfgPath := ConfigPath()
	logger.Info("removing config", "path", cfgPath)

	err := os.Remove(cfgPath)
	if err != nil {
		logger.Error("failed to remove config file", "path", cfgPath, "error", err)
		os.Exit(1)
	}
}
