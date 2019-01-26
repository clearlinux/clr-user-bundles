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

package operations

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"path/filepath"
	"clr-user-bundles/cublib"
)

func updateContent(statedir string, contentdir string, config cublib.TomlConfig) error {
	// -b and -N are essential, scripts are security dangerous since 3rd party content would get to run as root
	format, err := cublib.GetFormat()
	if err != nil {
		return err;
	}
	var out bytes.Buffer
 	certPath := path.Join(contentdir, "/usr/share/clear/update-ca/Swupd_Root.pem")
	cmd := exec.Command("swupd", "update", "-b", "-N", "-F", format, "-S", statedir, "-p", contentdir, "-u", config.Bundle.URL, "-C", certPath)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		return errors.New(out.String())
	}

	return nil
}

func Update(statedir string, contentdir string, postJob bool) {
	// GetLock causes program exit on failure to acquire lockfile
	cublib.GetLock(statedir)
	defer cublib.ReleaseLock(statedir)
	pstatedir := path.Join(statedir, "3rd-party")
	chrootdir := path.Join(contentdir, "chroot")
	dlist, err := ioutil.ReadDir(chrootdir)
	if err != nil {
		log.Fatalf("Unable to read 3rd-party content directory (%s): %s", chrootdir, err)
	}

	for _, p := range dlist {
		// chroot dir should only be chroot directories and conf files so skip the conf files
		// as it is easier to make those names from the directory names
		if ext := filepath.Ext(p.Name()); ext != "" {
			continue
		}
		confPath := "file://" + path.Join(chrootdir, p.Name()) + ".toml"
		conf, err := cublib.GetConfig(confPath)
		if err != nil {
			log.Printf("WARNING: Unable to read 3rd party config (%s): %s", confPath, err)
			continue
		}
		// NOTE: content chroot exists but matching config doesn't => warning
		// BUT content chroot doesn't exist and config does => ignored, manual cleanup required
		if err = updateContent(path.Join(pstatedir, p.Name()), path.Join(chrootdir, p.Name()), conf); err != nil {
			log.Printf("WARNING: Unable to update (%s %s): %s", conf.Bundle.URL, conf.Bundle.Name, err)
		}
	}
	if postJob {
		if err = cublib.PostProcess(statedir, contentdir); err != nil {
			log.Fatalf("%s", err)
		}
	}
}
