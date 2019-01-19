#!/bin/bash

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

RET=0
BASEDIR="$(pwd)"
TESTDIR="$1"

cleanup() {
    sudo rm -fr c c2 s o test.toml privatekey.pem Swupd_Root.pem
    if [ -f /etc/ca-certs/trusted/Swupd_Root.pem ]; then
        sudo clrtrust remove /etc/ca-certs/trusted/Swupd_Root.pem &> /dev/null
    fi
    popd &> /dev/null
    exit $RET
}

pushd "$TESTDIR" &> /dev/null

# Create test content chroots
sudo cp -r content c
sudo cp -r content2 c2

# Generate template
source /usr/lib/os-release
sed -e "s|@VERSION@|${VERSION_ID}|" example-config.toml > test.toml
sed -i "s|@TESTDIR@|file:///${PWD}/s/www/update|" test.toml

# Build content
sudo "${BASEDIR}/clr-user-bundles.py" s c test.toml
if [ $? -ne 0 ]; then
    echo "Build content failed"
    ret=1
    cleanup
fi

# Install cert to trust store
sudo clrtrust add Swupd_Root.pem &> /dev/null
if [ $? -ne 0 ]; then
    echo "Certificate couldn't be added to trust store"
    ret=1
    cleanup
fi

# Install content
sudo "${BASEDIR}/swupd-3rd-party" add "file:///${PWD}/s/www/update" -c "${PWD}/o"
if [ $? -ne 0 ]; then
    echo "Install content failed"
    ret=1
    cleanup
fi

# Validate install worked
diff o/chroot/*/usr/bin/test.sh c/usr/bin/test.sh
if [ $? -ne 0 ]; then
    echo "Install content failed to verify"
    ret=1
    cleanup
fi

# Run post process job
sudo "${BASEDIR}/3rd-party-post" -c "${PWD}/o"
if [ $? -ne 0 ]; then
    echo "Post process job failed"
    ret=1
    cleanup
fi

# Verify post process job worked
PATH="$PWD/o/bin:$PATH" test.sh | grep baz -q
if [ $? -ne 0 ]; then
    echo "Post process job failed to verify"
    ret=1
    cleanup
fi
PATH=$PWD/o/bin:$PATH test.sh | grep fooenv -q
if [ $? -ne 0 ]; then
    echo "Post process job failed to setup environment"
    ret=1
    cleanup
fi

# Build content2
sudo "${BASEDIR}/clr-user-bundles.py" s c2 test.toml
if [ $? -ne 0 ]; then
    echo "Build content2 failed"
    ret=1
    cleanup
fi

# Update to content2
sudo "${BASEDIR}/swupd-3rd-party" update -c "${PWD}/o"
if [ $? -ne 0 ]; then
    echo "Update to content2 failed"
    ret=1
    cleanup
fi

# Validate update worked
diff o/chroot/*/usr/bin/test.sh c2/usr/bin/test.sh
if [ $? -ne 0 ]; then
    echo "Update to content2 failed to verify"
    ret=1
    cleanup
fi

# Run post process on update
sudo "${BASEDIR}/3rd-party-post" -c "${PWD}/o"
if [ $? -ne 0 ]; then
    echo "Post process of update failed"
    ret=1
    cleanup
fi

# Verify post process of update worked
PATH="$PWD/o/bin:$PATH" test.sh | grep zab -q
if [ $? -ne 0 ]; then
    echo "Post process of update failed to verify"
    ret=1
    cleanup
fi
PATH="$PWD/o/bin:$PATH" test.sh | grep barenv -q
if [ $? -ne 0 ]; then
    echo "Post process of update failed to setup environment"
    ret=1
    cleanup
fi

# Verify remove works
sudo "${BASEDIR}/swupd-3rd-party" remove "file:///${PWD}/s/www/update" test -c "${PWD}/o"
if [ $? -ne 0 ]; then
    echo "Remove content failed"
    ret=1
    cleanup
fi

# Run post process on removal
sudo "${BASEDIR}/3rd-party-post" -c "${PWD}/o"
if [ $? -ne 0 ]; then
    echo "Post process of removal failed"
    ret=1
    cleanup
fi

# Verify post process of removal worked
if [ -d o/bin ]; then
    echo "Post process of removal failed to verify"
    ret=1
    cleanup
fi

cleanup
