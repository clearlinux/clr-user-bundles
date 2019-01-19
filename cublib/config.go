// Copyright © 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cublib

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"github.com/BurntSushi/toml"
)

type TomlConfig struct {
 	Bundle BundleConfig
}

type BundleConfig struct {
	Name        string
	Description string
	Includes    []string
	URL         string
	Bin         []string
}

func ReadConfig(buffer io.Reader) (TomlConfig, error) {
	var config TomlConfig

	if _, err := toml.DecodeReader(buffer, &config); err != nil {
		return TomlConfig{}, err
	}

	return config, nil
}

func GetConfig(uri string) (TomlConfig, error) {
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		return TomlConfig{}, err
	}
	var breader io.Reader
	if url.Scheme == "file" {
		buffer, err := ioutil.ReadFile(url.Path)
		if err != nil {
			return TomlConfig{}, err
		}
		breader = bytes.NewReader(buffer)
	} else {
		resp, err := http.Get(uri)
		if err != nil {
			return TomlConfig{}, err
		} else {
			defer resp.Body.Close()
			breader = resp.Body
		}
	}

	return ReadConfig(breader)
}

func WriteConfig(outPath string, config TomlConfig, overwrite bool) error {
	out, err := os.Create(outPath)
	if err != nil {
		return err
	}
	if err = toml.NewEncoder(out).Encode(config); err != nil {
		return err
	}
	return nil
}
