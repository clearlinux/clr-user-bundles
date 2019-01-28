==================
mixer-user-bundler
==================

------------------------
3rd-party bundle builder
------------------------

:Copyright: \(C) 2019 Intel Corporation, CC-BY-SA-3.0
:Manual section: 1


SYNOPSIS
========

``mixer-user-bundler [STATEDIR] [CHROOTDIR] [CONFIG] <flags>``


DESCRIPTION
===========

``mixer-user-bundler``\(1) is a software update generator that takes a content
root \(CHROOTDIR) and CONFIG to create update metadata consumable by ``swupd``\(1).

``mixer-user-bundler`` will modify chrootdir with additional content required
for ``swupd`` so it is recommended to make a copy of chrootdir that is used
only with ``mixer-user-bundler``. The STATEDIR content needs to be maintained
between runs of ``mixer-user-bundler`` and the ``swupd`` update content is
located within it at STATEDIR/www/update which needs to be available to clients
to make use of this content.

The output of ``mixer-user-bundler`` is a set of manifests readable by ``swupd``
as well as all the OS content ``swupd`` needs to perform its update operations.
The OS content includes all the files in an update as well as zero and
delta-packs for improved update performance. The content that
``mixer-user-bundler`` produces is tied to a specific format so that ``swupd``
is guaranteed to understand it if the client is using the right version of
``swupd``. The way users are expected to interact with the content is through
``swupd-3rd-party``\(1) and ``3rd-party-post``\(1) commands that setup the
content on the end users system based on the configuration provided when
creating the user bundle. See ``3rd-party-post``\(1), ``swupd-3rd-party``\(1),
``swupd``\(1) and ``os-format``\(7) for more details.


OPTIONS
=======

The following options are applicable to be used to modify the core behavior and
resources that ``mixer-user-bundler`` uses.

-  ``-h, --help``

   Display general help information.


FILES
=====

`CONFIG.toml`

    The mixer-user-bundler configuration file. This is a toml* formatted file
    containing two tables. The first is an upstream table with url and version
    keys, where url is a string that references the upstream content that
    the user bundle is based on and version is an integer of the upstream
    version the user bundle is to be built against. The second table is a
    bundle table with name, description, includes, url and bin keys. The name
    key is a string with the name of the user bundle being created, the
    description key is a short string describing what the bundle's purpose is
    for an end user, the includes key is an array of strings that are names
    of upstream bundles that are to be installed in order for the user bundle
    to function, the url key is a string which has the update url for the user
    bundle content (pointing to the STATEDIR/www/update content) and the bin
    key is an array of strings with absolute paths that resolve to executables
    in the CHROOTDIR that are to be exposed as executables to the end user.


EXIT STATUS
===========

On success, 0 is returned. A non-zero return code indicates a failure.

SEE ALSO
--------

* ``3rd-party-post``\(1)
* ``swupd``\(1)
* ``swupd-3rd-party``\(1)
* ``os-format``\(7)
* https://github.com/clearlinux/swupd-client
* https://clearlinux.org/documentation/
* https://github.com/toml-lang/toml
