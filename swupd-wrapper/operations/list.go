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
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"clr-user-bundles/cublib"
)

func List(statedir string, contentdir string) {
	// GetLock causes program exit on failure to acquire lockfile
	cublib.GetLock(statedir)
	defer cublib.ReleaseLock(statedir)
	chrootdir := path.Join(contentdir, "chroot")
	dlist, err := ioutil.ReadDir(chrootdir)
	if err != nil {
		log.Fatalf("Unable to read 3rd-party content directory (%s): %s", chrootdir, err)
	}

	fmt.Println("Installed 3rd-party bundles")
	for _, p := range dlist {
		// chroot dir should only be chroot directories and conf files so skip the conf files
		// as it is easier to make those names from the directory names
		if ext := filepath.Ext(p.Name()); ext != "" {
			continue
		}
		confPath := "file://" + path.Join(chrootdir, p.Name()) + ".toml"
		conf, err := cublib.GetConfig(confPath)
		if err != nil {
			log.Printf("WARNING: Unable to read 3rd-party config (%s): %s", confPath, err)
			continue
		}
		// Includes can be updated by the 3rd-party repo so show the updated config in that case
		newConfPath := "file://" + path.Join(chrootdir, p.Name(), "usr", "user-config.toml")
		newConf, err := cublib.GetConfig(newConfPath)
		if err != nil {
			log.Printf("WARNING: Unable to read updated 3rd-party config (%s): %s", newConfPath, err)
			continue
		}
		fmt.Println("")
		fmt.Println("Included Bundles:")
		fmt.Printf("Name:              %-28s\n", conf.Bundle.Name)
		fmt.Printf("Description:       %-28s\n", conf.Bundle.Description)
		fmt.Printf("URL:               %-28s\n", conf.Bundle.URL)
		if len(conf.Bundle.Bin) > 0 {
			fmt.Println("Applications:")
			for _, app := range conf.Bundle.Bin {
				fmt.Printf("                   %-28s\n", app)
			}
		}
		if len(newConf.Bundle.Includes) > 0 {
			fmt.Println("Included Bundles:")
			for _, include := range newConf.Bundle.Includes {
				fmt.Printf("                   %-28s\n", include)
			}
		}
	}
}
