===============
swupd-3rd-party
===============

-----------------------------------
swupd wrapper for 3rd-party content
-----------------------------------

:Copyright: \(C) 2019 Intel Corporation, CC-BY-SA-3.0
:Manual section: 1


SYNOPSIS
========

``swupd-3rd-party [SUBCOMMANDS] <flags>``


DESCRIPTION
===========

``swupd-3rd-party``\(1) is a wrapper for ``swupd``\(1) that enables working with
3rd-party content in a way that conforms to the ``stateless``\(7) configuration
and content repositories built by ``mixer-user-bundler``\(1).

Contents are installed by default under /opt/3rd-party with hooks using
``3rd-party-post``\(1) for configured applications to be available under
/opt/3rd-party/bin which should be added to the PATH as the last entry.


OPTIONS
=======

The following options are applicable to all subcommands, and can be
used to modify the core behavior and resources that ``swupd-3rd-party``
uses.

-  ``-h, --help``

   Display general help information.

-  ``-c, --contentdir``

   Changes the installation directory for 3rd-party content.

-  ``-s, --statedir``

   Changes the statedir used by ``swupd``.


SUBCOMMANDS
===========

``add`` [URI] <addflags>

    Add 3rd-party repo based on URI of the content. Content must be signed
    with a certificate trusted by the system trust store.

    addflags:
    - ``-p, --skip-post``
      Skip running ``3rd-party-post`` processing.

``list``

    Display installed 3rd-party content and its configured settings.

``remove`` [URI] [BUNDLE] <removeflags>

    Remove 3rd-party repo based on URI and BUNDLE name of the content.

    removeflags:
    - ``-p, --skip-post``
      Skip running ``3rd-party-post`` processing.

``update`` <updateflags>

    Update all 3rd-party repositories on the system.

    updateflags:
    - ``-p, --skip-post``
      Skip running ``3rd-party-post`` processing.


EXIT STATUS
===========

On success, 0 is returned. A non-zero return code indicates a failure.

SEE ALSO
--------

* ``3rd-party-post``\(1)
* ``mixer-user-bundler``\(1)
* ``swupd``\(1)
* https://github.com/clearlinux/swupd-client
* https://clearlinux.org/documentation/
