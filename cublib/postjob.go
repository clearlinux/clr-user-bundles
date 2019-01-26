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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
)

func setupBins(statedir string, contentdir string, installdir string, bins []string) error {
	scriptTemplate := `#!/bin/bash

export PATH=%s:%s
export LD_LIBRARY_PATH=%s:%s
`
	internalBinPath := fmt.Sprintf("%s/usr/bin", installdir)
	internalLdPath := fmt.Sprintf("%s/usr/lib64", installdir)
	targetPath := path.Join(contentdir, ".bin")
	err := os.MkdirAll(targetPath, 0755)
	if err != nil {
		return err
	}
	envPath := os.Getenv("PATH")
	envLdPath := os.Getenv("LD_LIBRARY_PATH")
	binScript := fmt.Sprintf(scriptTemplate, internalBinPath, envPath, internalLdPath, envLdPath)
	for _, b := range bins {
		if _, err = os.Lstat(path.Join(installdir, b)); err != nil {
			log.Printf("WARNING: Application %s set to be installed but not found in %s", b, installdir)
			continue
		}
		fullBinScript := fmt.Sprintf("%s%s \"$@\"\n", binScript, path.Join(installdir, b))
		err = ioutil.WriteFile(path.Join(targetPath, path.Base(b)), []byte(fullBinScript), 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func stageContent(contentdir string) error {
	items := []string{"bin"}
	for _, item := range items {
		old := path.Join(contentdir, item)
		new := path.Join(contentdir, fmt.Sprintf(".%s", item))
		err := os.RemoveAll(old)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		_, err = os.Stat(new)
		if os.IsNotExist(err) {
			continue
		}
		err = os.Rename(new, old)
		if err != nil {
			return err
		}
	}
	return nil
}

func PostProcess(statedir string, contentdir string) error {
	pstatedir := path.Join(statedir, "3rd-party")
	chrootdir := path.Join(contentdir, "chroot")
	dlist, err := ioutil.ReadDir(chrootdir)
	if err != nil {
		return fmt.Errorf("Unable to read 3rd-party content directory (%s): %s", pstatedir, err)
	}

	for _, p := range dlist {
		// chroot dir should only be chroot directories and conf files so skip the conf files
		// as it is easier to make those names from the directory names
		if ext := filepath.Ext(p.Name()); ext != "" {
			continue
		}
		confPath := "file://" + path.Join(chrootdir, p.Name()) + ".toml"
		conf, err := GetConfig(confPath)
		if err != nil {
			log.Printf("WARNING: Unable to read 3rd party config (%s): %s", confPath, err)
			continue
		}
		if err = setupBins(pstatedir, contentdir, path.Join(chrootdir, p.Name()), conf.Bundle.Bin); err != nil {
			log.Printf("WARNING: Unable to create bin scripts for %s: %s", conf.Bundle.Name, err)
		}
	}
	if err = stageContent(contentdir); err != nil {
		return fmt.Errorf("User content not staged successfully to %s: %s", contentdir, err)
	}

	return nil
}
