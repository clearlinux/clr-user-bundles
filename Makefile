.PHONY: test clean all

MANPAGES := \
	3rd-party-post.1 \
	mixer-user-bundler.1 \
	swupd-3rd-party.1

all: vendor man
	(cd ./swupd-wrapper && go build -mod=vendor -o ../swupd-3rd-party)
	(cd ./post-job && go build -mod=vendor -o ../3rd-party-post)

install: all
	install -D -m 00755 swupd-3rd-party $(DESTDIR)/usr/bin/swupd-3rd-party
	install -D -m 00755 3rd-party-post $(DESTDIR)/usr/bin/3rd-party-post
	install -D -m 00755 clr-user-bundles.py $(DESTDIR)/usr/bin/mixer-user-bundler
	install -D -m 00644 data/3rd-party-update.service $(DESTDIR)/usr/lib/systemd/system/3rd-party-update.service
	install -D -m 00644 data/3rd-party-update.timer $(DESTDIR)/usr/lib/systemd/system/3rd-party-update.timer
	install -D -m 00644 3rd-party-post.1 $(DESTDIR)/usr/share/man/man1/3rd-party-post.1
	install -D -m 00644 mixer-user-bundler.1 $(DESTDIR)/usr/share/man/man1/mixer-user-bundler.1
	install -D -m 00644 swupd-3rd-party.1 $(DESTDIR)/usr/share/man/man1/swupd-3rd-party.1

clean:
	rm -f swupd-3rd-party 3rd-party-post clr-user-bundles-*.tar.xz
	rm -fr vendor

vendor: go.mod
	go mod vendor

test: all
	tests/swupd-wrapper/test-runner.sh tests/swupd-wrapper

man: $(MANPAGES)

%: docs/%.rst
	rst2man.py "$<" > "$@.tmp" && mv -f "$@.tmp" "$@"

dist: vendor
	$(eval TMP := $(shell mktemp -d))
	cp -r . $(TMP)/clr-user-bundles-$(VERSION)
	(cd $(TMP)/clr-user-bundles-$(VERSION); git reset --hard $(VERSION); git clean -xf; rm -fr .git .gitignore)
	tar -C $(TMP) -cf clr-user-bundles-$(VERSION).tar .
	xz clr-user-bundles-$(VERSION).tar
	rm -fr $(TMP)
