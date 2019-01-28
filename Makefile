.PHONY: test clean all

MANPAGES := \
	3rd-party-post.1 \
	mixer-user-bundler.1 \
	swupd-3rd-party.1

all: vendor man
	go build -mod=vendor -o swupd-3rd-party ./swupd-wrapper
	go build -mod=vendor -o 3rd-party-post ./post-job

install: all
	install -D -m 00755 swupd-3rd-party $(DESTDIR)/usr/bin/swupd-3rd-party
	install -D -m 00755 3rd-party-post $(DESTDIR)/usr/bin/3rd-party-post
	install -D -m 00755 clr-user-bundles.py $(DESTDIR)/usr/bin/mixer-user-bundler
	install -D -m 00644 3rd-party-post.1 $(DESTDIR)/usr/share/man/man1/3rd-party-post.1
	install -D -m 00644 mixer-user-bundler.1 $(DESTDIR)/usr/share/man/man1/mixer-user-bundler.1
	install -D -m 00644 swupd-3rd-party.1 $(DESTDIR)/usr/share/man/man1/swupd-3rd-party.1

clean:
	rm -f swupd-3rd-party 3rd-party-post
	rm -fr vendor

vendor: go.mod
	go mod vendor

test: all
	tests/swupd-wrapper/test-runner.sh tests/swupd-wrapper

man: $(MANPAGES)

%: docs/%.rst
	rst2man.py "$<" > "$@.tmp" && mv -f "$@.tmp" "$@"
