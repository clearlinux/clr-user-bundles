==============
3rd-party-post
==============

------------------------------
Post job for 3rd-party content
------------------------------

:Copyright: \(C) 2019 Intel Corporation, CC-BY-SA-3.0
:Manual section: 1


SYNOPSIS
========

``3rd-party-post <flags>``


DESCRIPTION
===========

``3rd-party-post``\(1) is a tool for creating, updating and cleaning up
3rd-party content artifacts that were added through ``swupd-3rd-party``\(1).

Contents installed (by default) under /opt/3rd-party are processed and
configured applications will have runner scripts generated for them under
/opt/3rd-party/bin (which should be added to the PATH as the last entry).


OPTIONS
=======

The following options are applicable to be used to modify the core behavior and
resources that ``3rd-party-post`` uses.

-  ``-h, --help``

   Display general help information.

-  ``-c, --contentdir``

   Changes the installation directory for 3rd-party content.

-  ``-s, --statedir``

   Changes the statedir used by ``swupd``\(1).


EXIT STATUS
===========

On success, 0 is returned. A non-zero return code indicates a failure.

SEE ALSO
--------

* ``swupd``\(1)
* ``swupd-3rd-party``\(1)
* https://github.com/clearlinux/swupd-client
* https://clearlinux.org/documentation/
