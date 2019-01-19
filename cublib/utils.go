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
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

var lockfd = -1

func GetFormat() (string, error) {
	format, err := ioutil.ReadFile("/usr/share/defaults/swupd/format")
	if err != nil {
		return "", err
	}
	return string(format), nil
}

func GetVersion(uri string, statedir string) (string, error) {
	cmd := exec.Command("swupd", "update", "-S", statedir, "-s", "-u", uri)
	// cmd := exec.Command("swupd", "update", "-s", "-u", uri)
	var out bytes.Buffer
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return "", err
		}
	}
	// Output is of the form:
	// Current OS version: XXX
	// Latest server version: YYY
	// Grab YYY
	// TODO make a swupd-client check-update command to just give us this value.
	if _, err = out.ReadBytes(':'); err != nil {
		return "", err
	}
	if _, err = out.ReadBytes(':'); err != nil {
		return "", err
	}
	if _, err = out.ReadByte(); err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

// Take lock for a given statedir, causes program to exit if it would fail to get the lock.
func GetLock(statedir string) {
	if !path.IsAbs(statedir) {
		log.Fatalf("Error: state directory path (%s) is not absolute", statedir)
	}

	err := os.MkdirAll(statedir, 0700)
	if err != nil {
		log.Fatalf("Unable to create statedir (%s): %s", statedir, err)
	}

	lockfile := path.Join(statedir, "3rd-party.lock")
	flock := syscall.Flock_t{
		Type: syscall.F_WRLCK,
		Start: 0, Len: 0, Whence: 0, Pid: int32(syscall.Getpid()),
	}
	fd, err := syscall.Open(lockfile, syscall.O_CREAT|syscall.O_RDWR|syscall.O_CLOEXEC, 0600)
	if err != nil {
		syscall.Close(fd)
		log.Fatalf("Lockfile (%s) open failed: %s", lockfile, err)
	}
	if err := syscall.FcntlFlock(uintptr(fd), syscall.F_SETLK, &flock); err != nil {
		log.Fatalf("Unable to set flock on %s: %s", lockfile, err)
	}
	lockfd = fd
}

func ReleaseLock(statedir string) {
	if lockfd >= 0 {
		syscall.Close(lockfd)
	}
}

func GetCert(statedir string, uri string) (string, error) {
	url, err := url.ParseRequestURI(uri)
	if err != nil {
		return "", err
	}
	var breader io.Reader
	if url.Scheme == "file" {
		buffer, err := ioutil.ReadFile(url.Path)
		if err != nil {
			return "", err
		}
		breader = bytes.NewReader(buffer)
	} else {
		resp, err := http.Get(uri)
		if err != nil {
			return "", err
		} else {
			defer resp.Body.Close()
			breader = resp.Body
		}
	}

	outpath := path.Join(statedir, "Swupd_Root.pem")
	out, err := os.Create(outpath)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(out, breader); err != nil {
		return "", err
	}

	return outpath, nil
}

func GetEncodedBundleName(url string, name string) string {
	return base64.StdEncoding.EncodeToString([]byte(url + name))
}
