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
	"log"
	"path"
	"os"
	"os/exec"
	"strings"
	"clr-user-bundles/cublib"
)

func Add(uri string, statedir string, contentdir string, skipPost bool) {
	// GetLock causes program exit on failure to acquire lockfile
	cublib.GetLock(statedir)
	defer cublib.ReleaseLock(statedir)
	format, err := cublib.GetFormat()
	if err != nil {
		log.Fatalf("Unable to get format from filesystem: %s", err)
	}
	version, err := cublib.GetVersion(uri, statedir)
	if err != nil {
		log.Fatalf("Unable to get version from uri (%s): %s", uri, err)
	}

	configBasename := "user-config.toml"
	postfix := path.Join("/", version, configBasename)
	configURI := uri + postfix
	config, err := cublib.GetConfig(configURI)
	if err != nil {
		log.Fatalf("Error accessing configuration from (%s): %s", configURI, err)
	}

	if config.Bundle.URL != uri {
		log.Printf("WARNING: bundle configured url (%s) and url used to add bundle (%s) differ", config.Bundle.URL, uri)
	}

	chrootdir := path.Join(contentdir, "chroot")
	err = os.MkdirAll(chrootdir, 0755)
	if err != nil {
		log.Fatalf("Unable to make toplevel 3rd party content directory (%s): %s", contentdir, err)
	}

	bnameEncoded := cublib.GetEncodedBundleName(config.Bundle.URL, config.Bundle.Name)
	pstatedir := path.Join(statedir, "3rd-party", bnameEncoded)
	err = os.MkdirAll(pstatedir, 0700)
	if err != nil {
		log.Fatalf("Unable to make 3rd party state directory (%s): %s", pstatedir, err)
	}
	configPath := path.Join(chrootdir, bnameEncoded) + ".toml"
	if _, err = os.Stat(configPath); !os.IsNotExist(err) {
		log.Fatalf("Config %s already exists, exiting", configPath)
	}
	err = cublib.WriteConfig(configPath, config, false)
	if err != nil {
		Remove(statedir, contentdir, config.Bundle.URL, config.Bundle.Name, false, false)
		log.Fatalf("Unable to save bundle configuration file to 3rd party state directory (%s): %s", pstatedir, err)
	}

	pchrootdir := path.Join(chrootdir, bnameEncoded)
	if _, err = os.Stat(pchrootdir); !os.IsNotExist(err) {
		log.Fatalf("Content path %s already exists, try running remove operation on partially installed content", pchrootdir)
	}
	err = os.MkdirAll(pchrootdir, 0755)
	if err != nil {
		Remove(statedir, contentdir, config.Bundle.URL, config.Bundle.Name, false, false)
		log.Fatalf("Unable to make 3rd party state directory (%s): %s", pstatedir, err)
	}

	certURI := config.Bundle.URL + path.Join("/", version, "Swupd_Root.pem")
	certPath, err := cublib.GetCert(pstatedir, certURI)
	if err != nil {
		Remove(statedir, contentdir, config.Bundle.URL, config.Bundle.Name, false, false)
		log.Fatalf("Unable to load certificate (%s): %s", certURI, err)
	}

	var cmd *exec.Cmd
	out := bytes.Buffer{}
	cmd = exec.Command("openssl", "verify", certPath)
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		Remove(statedir, contentdir, config.Bundle.URL, config.Bundle.Name, false, false)
		log.Printf("Certificate (%s) isn't trusted: %s", certPath, out.String())
		log.Fatalf("Please add certificate to trust chain")
	}

	if len(config.Bundle.Includes) > 0 {
		includes := strings.Join(config.Bundle.Includes, " ")
		cmd = exec.Command("swupd", "bundle-add", includes)
		out = bytes.Buffer{}
		cmd.Stdout = &out
		cmd.Stderr = &out
		err = cmd.Run()
		if err != nil {
			Remove(statedir, contentdir, config.Bundle.URL, config.Bundle.Name, false, false)
			log.Fatalf("Unable to install dependency bundle(s) %s to the base system: %s", includes, out.String())
		}
	}

	// -b and -N are essential, scripts are security dangerous since 3rd party content would get to run as root
	cmd = exec.Command("swupd", "verify", "-f", "-b", "-N", "-S", pstatedir, "-p", pchrootdir, "-u", config.Bundle.URL, "-F", format, "-m", version, "-x", "-B", config.Bundle.Name, "-C", certPath)
	out = bytes.Buffer{}
	cmd.Stdout = &out
	cmd.Stderr = &out
	err = cmd.Run()
	if err != nil {
		Remove(statedir, contentdir, config.Bundle.URL, config.Bundle.Name, false, false)
		log.Fatalf("Unable to install bundle %s from %s: %s", config.Bundle.Name, config.Bundle.URL, out.String())
	}
	if err = os.Remove(certPath); err != nil {
		log.Printf("WARNING: Unable to remove temporary cert (%s): %s", certPath, err)
	}

	if skipPost {
		return
	}
	if err = cublib.PostProcess(statedir, contentdir); err != nil {
			log.Fatalf("%s", err)
	}
}
