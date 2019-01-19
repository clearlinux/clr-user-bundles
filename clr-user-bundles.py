#!/usr/bin/env python3

# Copyright © 2019 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from io import BytesIO

import argparse
import hashlib
import os
import shutil
import stat
import struct
import subprocess
import sys
import tempfile
import time

import pycurl
import toml

ZERO_HASH = '0000000000000000000000000000000000000000000000000000000000000000'


def parse_args():
    """Handle arguments."""
    p = argparse.ArgumentParser(description="Build Clear Linux user bundles.")
    p.add_argument('statedir', help="Directory to store user bundle creation output.")
    p.add_argument('chrootdir', help="Directory containing the content to be turned into a bundle.")
    p.add_argument('config', help="Configuration file for generating the user bundle.")
    return p.parse_args()


def do_curl(url):
    """Perform a curl operation for `url`."""
    c = pycurl.Curl()
    c.setopt(c.URL, url)
    c.setopt(c.FOLLOWLOCATION, True)
    c.setopt(c.FAILONERROR, True)
    buf = BytesIO()
    c.setopt(c.WRITEDATA, buf)
    try:
        c.perform()
    except pycurl.error as exptn:
        print(f"Unable to fetch {url}: {exptn}")
        return None
    finally:
        c.close()

    return buf.getvalue().decode('utf-8')


def load_manifest(url, version, manifest_name):
    """Download and parse manifest."""
    manifest_raw = do_curl(f"{url}/{version}/Manifest.{manifest_name}")
    manifest = {}
    if not manifest_raw:
        raise Exception(f"Unable to load manifest {manifest_name}")

    try:
        lines = manifest_raw.splitlines()
        for idx, line in enumerate(lines):
            content = line.split('\t')
            if content[0] == "MANIFEST":
                manifest['format'] = content[1]
            elif content[0] == "version:":
                manifest['version'] = content[1]
            elif content[0] == "previous:":
                manifest['previous'] = content[1]
            elif content[0] == "minversion:":
                manifest['minversin'] = content[1]
            elif content[0] == "filecount:":
                manifest['filecount'] = content[1]
            elif content[0] == "timestamp:":
                manifest['timestamp'] = content[1]
            elif content[0] == "contentsize:":
                manifest['contentsize'] = content[1]
            elif content[0] == "includes":
                if not manifest.get('includes'):
                    manifest['includes'] = []
                manifest['includes'].append(content[1])
            elif len(content) == 4:
                if not manifest.get('files'):
                    manifest['files'] = {}
                manifest['files'][content[3]] = content
    except Exception as _:
        raise Exception(f"Unable to parse manifest {manifest_name} at line {idx+1}: {line}")

    if not manifest.get('includes'):
        manifest['includes'] = []
    if not manifest.get('files'):
        raise Exception(f"Invalid manifest {manifest_name}, missing file section")

    return manifest


def copy_certificate(chrootdir, statedir, name):
    """Add certificate to chroot contents."""
    if not os.path.isfile("privatekey.pem") or not os.path.isfile("Swupd_Root.pem"):
        subprocess.run(["openssl", "req", "-x509", "-sha256", "-nodes", "-newkey", "rsa:4096",
                        "-keyout", "privatekey.pem", "-out", "Swupd_Root.pem", "-days", "1825",
                        "-subj", "/C=US/ST=Oregon/L=Hillsboro/O=Example/CN=www.example.com"],
                       capture_output=True, check=True)
    chroot_path = os.path.join(chrootdir, "usr", "share", "clear", "update-ca")
    os.makedirs(chroot_path, exist_ok=True)
    shutil.copyfile("Swupd_Root.pem", os.path.join(chroot_path, "Swupd_Root.pem"))


def copy_config(full_config, chrootdir, statedir, version):
    """Add config to chroot contents."""
    chroot_path = os.path.join(chrootdir, "usr")
    state_path = os.path.join(statedir, "www", "update", version)
    os.makedirs(chroot_path, exist_ok=True)
    os.makedirs(state_path, exist_ok=True)
    cpath = os.path.join(chroot_path, "user-config.toml")
    config = {}
    config['bundle'] = full_config['bundle']
    with open(cpath, "w") as cfile:
        cfile.write(toml.dumps(config))
    shutil.copyfile(cpath, os.path.join(state_path, "user-config.toml"))


def get_base_manifests(includes, url, version):
    """Load upstream manifest files."""
    manifests = {}
    mom = load_manifest(url, version, "MoM")
    for include in includes:
        try:
            include_version = mom['files'][include][2]
        except Exception as _:
            raise Exception(f"Included bundle {include} not found in upstream {url} for version {version}")
        manifests[include] = load_manifest(url, include_version, include)
        for recursive_include in manifests[include]['includes']:
            if manifests.get(recursive_include):
                continue
            try:
                rinclude_version = mom['files'][recursive_include][2]
            except Exception as _:
                raise Exception(f"Bundle {recursive_include}, included by {include} not found in upstream {url} for version {version}")
            manifests[recursive_include] = load_manifest(url, rinclude_version, recursive_include)

    return manifests


def get_previous_version(statedir):
    """Parse version of previous user content release."""
    version_path = os.path.join(statedir, "www", "update", "version")
    version = 0
    try:
        latest_format = sorted(os.listdir(version_path), key=lambda x: int(x.strip("format")))[-1]
        with open(os.path.join(version_path, f"{latest_format}", "latest"), "r") as lfile:
            version = int(lfile.read().strip())
    except FileNotFoundError as exptn:
        pass
    return version


def get_hash(path):
    """Get hash for the file contents."""
    proc = subprocess.run(["swupd", "hashdump", path], encoding="utf-8", capture_output=True)
    if proc.returncode != 0:
        print(f"Failed to get hash of file {path}, swupd returned: {proc.stderr}, skipping")
        return None
    return proc.stdout.strip()


def get_flags(path, cpath, lstat):
    """Get flags for the file."""
    mode = list("....")
    if stat.S_ISDIR(lstat.st_mode):
        mode[0] = "D"
    elif stat.S_ISLNK(lstat.st_mode):
        mode[0] = "L"
    elif stat.S_ISREG(lstat.st_mode):
        mode[0] = "F"

    if mode[0] == ".":
        print(f"Invalide mode for {path}, skipping")
        return None

    if cpath.startswith("/usr/lib/kernel/") or cpath.startswith("/usr/lib/modules"):
        mode[2] = "b"

    return ''.join(mode)


def add_metadata(chrootdir, url, manifest_format, version, bundle_name):
    """Create chroot config files."""
    version_path = os.path.join(chrootdir, "usr", "lib")
    os.makedirs(version_path, exist_ok=True)
    with open(os.path.join(version_path, "os-release"), "w") as osfile:
        osfile.write(f"VERSION_ID={version}\n")
    sub_path = os.path.join(chrootdir, "usr", "share", "clear", "bundles")
    os.makedirs(sub_path, exist_ok=True)
    with open(os.path.join(sub_path, bundle_name), "w") as sfile:
        sfile.write("")
    # defaults_path = os.path.join(chrootdir, "usr", "share", "defaults", "swupd")
    # os.makedirs(defaults_path, exist_ok=True)
    # with open(os.path.join(defaults_path, "contenturl"), "w") as cfile:
    #     cfile.write(f"{url}")
    # with open(os.path.join(defaults_path, "versionurl"), "w") as vfile:
    #     vfile.write(f"{url}")
    # with open(os.path.join(defaults_path, "format"), "w") as ffile:
    #     ffile.write(f"{manifest_format}")


def scan_chroot(chrootdir, version, previous_version, previous_manifest, manifest_format):
    """Build manifest based off of a chroot directory."""
    manifest = {}
    manifest['format'] = manifest_format
    manifest['version'] = version
    manifest['previous'] = previous_version
    manifest['filecount'] = 0
    manifest['timestamp'] = int(time.time())
    manifest['contentsize'] = 0
    manifest['includes'] = []
    manifest['files'] = {}
    for root, dirs, files in os.walk(chrootdir):
        for dname in dirs:
            entry = []
            path = os.path.join(root, dname)
            cpath = path.lstrip(chrootdir)
            dstat = os.lstat(path)
            dhash = get_hash(path)
            if not dhash:
                continue
            flags = get_flags(path, cpath, dstat)
            if not flags:
                continue
            version = manifest['version']
            manifest['files'][cpath] = [flags, dhash, version, cpath]
            manifest['filecount'] += 1
        for fname in files:
            path = os.path.join(root, fname)
            cpath = path.lstrip(chrootdir)
            fstat = os.lstat(path)
            fhash = get_hash(path)
            if not fhash:
                continue
            flags = get_flags(path, cpath, fstat)
            if not flags:
                continue
            version = manifest['version']
            manifest['files'][cpath] = [flags, fhash, version, cpath]
            manifest['filecount'] += 1
            manifest['contentsize'] += os.lstat(path).st_size

    return manifest


def combine_manifests(new_manifest, previous_manifest):
    """Create a combined manifest by modifying old and new manifest files."""
    if not previous_manifest:
        return new_manifest

    for _, entry in new_manifest['files'].items():
        # hash equal, use previous manifest entry's version
        if not previous_manifest['files'].get(entry[3]):
            continue
        if entry[1] == previous_manifest['files'][entry[3]][1]:
            entry[2] = previous_manifest['files'][entry[3]][2]

    return new_manifest


def create_tar(input_path, output_path, input_name, output_name, transform=False, pack=False):
    """Create compressed tarfile in state directory."""
    tar_name = f"{output_name}.tar"
    tar_path = os.path.join(output_path, tar_name)
    if pack:
        tar_cmd = f"tar -C {input_path} -cf {tar_path} {input_name[0]} {input_name[1]}"
    elif transform:
        tar_cmd = f"tar --no-recursion -C {input_path} -cf {tar_path} {input_name} --transform s/{input_name}/{output_name}/"
    else:
        tar_cmd = f"tar --no-recursion -C {input_path} -cf {tar_path} {input_name}"
    proc = subprocess.run(tar_cmd, shell=True, capture_output=True)
    if proc.returncode != 0:
        raise Exception(f"Unable to create tar file for {os.path.join(input_path, input_name)}")
    proc = subprocess.run(f"xz {tar_path}", shell=True, capture_output=True)
    if proc.returncode != 0:
        raise Exception(f"Unable to compress tar file for {os.path.join(input_path, input_name)}")
    os.rename(f"{tar_path}.xz", f"{tar_path}")


def write_manifest(statedir, version, manifest, name):
    """Output final manifest file to the state directory."""
    mname = f"Manifest.{name}"
    mtmp = [f"MANIFEST\t{manifest['format']}\n",
            f"version:\t{manifest['version']}\n",
            f"previous:\t{manifest['previous']}\n",
            f"filecount:\t{manifest['filecount']}\n",
            f"timestamp:\t{manifest['timestamp']}\n",
            f"contentsize:\t{manifest['contentsize']}\n"]
    for include in sorted(manifest['includes']):
        mtmp.append(f"includes:\t{include}\n")

    mtmp.append('\n')
    for key in sorted(manifest['files'], key=lambda k: (manifest['files'][k][2],
                                                        manifest['files'][k][3])):
        mtmp.append(f"{manifest['files'][key][0]}\t{manifest['files'][key][1]}\t{manifest['files'][key][2]}\t{manifest['files'][key][3]}\n")

    out_path = os.path.join(statedir, "www", "update", version)
    if not os.path.isdir(out_path):
        os.makedirs(out_path, exist_ok=True)
    with open(os.path.join(out_path, mname), "w") as mout:
        mout.writelines(mtmp)
    create_tar(out_path, out_path, mname, mname)


def write_fullfiles(statedir, chrootdir, version, manifest):
    """Output fullfiles content to the state directory."""
    out_path = os.path.join(statedir, "www", "update", version, "files")
    if not os.path.isdir(out_path):
        os.makedirs(out_path, exist_ok=True)
    for val in manifest['files'].values():
        if val[2] != version or val[1] == ZERO_HASH:
            continue
        out_file = os.path.join(out_path, val[1])
        if os.path.isfile(f"{out_file}.tar"):
            continue
        in_file = os.path.join(chrootdir, val[3][1:])
        create_tar(os.path.dirname(in_file), out_path, os.path.basename(in_file), val[1], True)


def extract_file(statedir, file_entry, out_path):
    """Extract file in statedir to out_path."""
    tar_file = os.path.join(statedir, "www", "update", file_entry[2], "files", f"{file_entry[1]}.tar")
    proc = subprocess.run(["tar", "-C", out_path, "-xf", tar_file], capture_output=True)
    if proc.returncode != 0:
        raise Exception("Unable to extract previously created fullfile {file_entry[3]} from version {file_entry[2]}")
    return os.path.join(out_path, file_entry[1])


def write_deltafiles(statedir, chrootdir, version, manifest, previous_manifest):
    """Output deltafiles content to the state directory."""
    out_path = os.path.join(statedir, "www", "update", version, "delta")
    if not os.path.isdir(out_path):
        os.makedirs(out_path, exist_ok=True)
    for val in manifest['files'].values():
        if val[2] != version or val[0] != "F..." or val[1] == ZERO_HASH:
            continue
        if not previous_manifest['files'].get(val[3]):
            continue
        pval = previous_manifest['files'][val[3]]
        out_file = os.path.join(out_path, f"{pval[2]}-{val[2]}-{pval[1]}-{val[1]}")
        if os.path.isfile(out_file):
            continue
        with tempfile.TemporaryDirectory(dir=os.getcwd()) as odir:
            previous_file = extract_file(statedir, pval, odir)
            current_file = os.path.join(chrootdir, val[3][1:])
            try:
                proc = subprocess.run(["bsdiff", previous_file, current_file, out_file, "xz"],
                                      timeout=10, capture_output=True)
                if proc.returncode != 0:
                    shutil.rmtree(out_file, ignore_errors=True)
            except subprocess.TimeoutExpired as exptn:
                shutil.rmtree(out_file, ignore_errors=True)


def write_deltapack(statedir, chrootdir, version, manifest, previous_manifest, bundle_name):
    """Output deltapack to the statedir."""
    out_path = os.path.join(statedir, "www", "update", version)
    delta_path = os.path.join(out_path, "delta")
    if not os.path.isdir(out_path):
        os.makedirs(out_path, exist_ok=True)
    with tempfile.TemporaryDirectory(dir=os.getcwd()) as odir:
        staged = os.path.join(odir, "staged")
        delta = os.path.join(odir, "delta")
        os.makedirs(staged)
        os.makedirs(delta)
        for val in manifest['files'].values():
            if val[2] != version or val[1] == ZERO_HASH:
                continue
            delta_file = None
            if previous_manifest and previous_manifest['files'].get(val[3]):
                pval = previous_manifest['files'][val[3]]
                fname = f"{pval[2]}-{val[2]}-{pval[1]}-{val[1]}"
                delta_file = os.path.join(delta_path, fname)
                if not os.path.isfile(delta_file):
                    delta_file = None
            copy_file = os.path.join(chrootdir, val[3][1:])
            if delta_file:
                out_file = os.path.join(delta, fname)
                if os.path.isfile(f"{out_file}"):
                    continue
                shutil.copyfile(delta_file, out_file)
            else:
                out_file = os.path.join(staged, val[1])
                if os.path.exists(f"{out_file}"):
                    continue
                extract_file(statedir, val, staged)
        if previous_manifest:
            out_file = f"pack-{bundle_name}-from-{previous_manifest['version']}"
        else:
            out_file = f"pack-{bundle_name}-from-0"
        create_tar(odir, out_path, (staged, delta), out_file, pack=True)


def write_mom(statedir, manifest_format, version, previous_version, bundle_name):
    """Create a wrapper MoM for the user bundle."""
    bundle_hash = get_hash(os.path.join(statedir, "www", "update", version, f"Manifest.{bundle_name}"))
    manifest = {'format': manifest_format,
                'version': version,
                'previous': previous_version,
                'filecount': 1,
                'timestamp': int(time.time()),
                'contentsize': 0,
                'includes': [],
                'files': {bundle_name: ["M...", bundle_hash, version, bundle_name]}}
    write_manifest(statedir, version, manifest, "MoM")
    subprocess.run(["openssl", "smime", "-sign", "-binary", "-in",
                    os.path.join(statedir, "www", "update", version, "Manifest.MoM"),
                    "-signer", "Swupd_Root.pem", "-inkey", "privatekey.pem",
                    "-outform", "DER", "-out",
                    os.path.join(statedir, "www", "update", version, "Manifest.MoM.sig")],
                   check=True)
    shutil.copyfile("Swupd_Root.pem", os.path.join(statedir, "www", "update", version, "Swupd_Root.pem"))


def write_versions(statedir, manifest_format, version):
    """Create version files."""
    out_path = os.path.join(statedir, "www", "update", "version", f"format{manifest_format}")
    os.makedirs(out_path, exist_ok=True)
    first_path = os.path.join(out_path, "first")
    latest_path = os.path.join(out_path, "latest")
    if not os.path.isfile(os.path.join(out_path, "first")):
        with open(first_path, "w") as fout:
            fout.write(f"{version}")
    with open(latest_path, "w") as lout:
        lout.write(f"{version}")


def build_user_bundle(statedir, chrootdir, config):
    """Create manifests and other user bundle artifacts."""
    base_version = config['upstream']['version']
    bundle_includes = config['bundle']['includes']
    bundle_name = config['bundle']['name']
    bundle_url = config['bundle']['url']
    included_manifests = get_base_manifests(bundle_includes,
                                            config['upstream']['url'],
                                            base_version)
    manifest_format = included_manifests['os-core']['format']
    previous_version = get_previous_version(statedir)
    version = str(previous_version + 10)
    previous_version = str(previous_version)
    manifest_dir = os.path.join(os.getcwd(), statedir, "www", "update")
    if previous_version == "0":
        previous_manifest = None
    else:
        previous_manifest = get_base_manifests([bundle_name], f"file:///{manifest_dir}", previous_version)[bundle_name]
    copy_certificate(chrootdir, statedir, bundle_name)
    copy_config(config, chrootdir, statedir, version)
    add_metadata(chrootdir, bundle_url, manifest_format, version, bundle_name)
    new_manifest = scan_chroot(chrootdir, version, previous_version, previous_manifest, manifest_format)
    user_manifest = combine_manifests(new_manifest, previous_manifest)
    write_manifest(statedir, version, user_manifest, bundle_name)
    write_fullfiles(statedir, chrootdir, version, user_manifest)
    if previous_manifest:
        write_deltafiles(statedir, chrootdir, version, user_manifest, previous_manifest)
    write_deltapack(statedir, chrootdir, version, user_manifest, previous_manifest, bundle_name)
    write_mom(statedir, user_manifest['format'], version, previous_version, bundle_name)
    write_versions(statedir, user_manifest['format'], version)


def main():
    """Entry point for Clear Linux user bundle creation."""
    args = parse_args()
    try:
        config = toml.load(args.config)
    except Exception as exptn:
        print(f"Unable to load configuration file: {exptn}")
        sys.exit(-1)

    try:
        build_user_bundle(args.statedir, args.chrootdir, config)
    except Exception as exptn:
        print(f"Unable to create user bundle: {exptn}")
        sys.exit(-1)

if __name__ == '__main__':
    main()
