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
	"log"
	"os"
	"path"
	"clr-user-bundles/cublib"
)

func Remove(statedir string, contentdir string, uri string, name string, postJob bool) {
	// GetLock causes program exit on failure to acquire lockfile
	cublib.GetLock(statedir)
	defer cublib.ReleaseLock(statedir)
	encodedName := cublib.GetEncodedBundleName(uri, name)
	pstatedir := path.Join(statedir, "3rd-party", encodedName)
	chrootdir := path.Join(contentdir, "chroot", encodedName)
	err := os.RemoveAll(pstatedir)
	if err != nil {
		log.Printf("WARNING: Unable to remove 3rd-party state directory (%s): %s", pstatedir, err)
	}
	err = os.RemoveAll(chrootdir)
	if err != nil {
		log.Printf("WARNING: Unable to remove 3rd-party content directory (%s): %s", chrootdir, err)
	}
	err = os.Remove(chrootdir + ".toml")
	if err != nil {
		log.Printf("WARNING: Unable to remove 3rd-party config (%s): %s", chrootdir + ".toml", err)
	}
	if postJob {
		if err = cublib.PostProcess(statedir, contentdir); err != nil {
			log.Fatalf("%s", err)
		}
	}
}
