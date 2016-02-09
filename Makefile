# ensure that the binary is compiled from the current sources because we get
# the VERSION from it (single source of truth).
IGNORE := $(shell go build -o bytemine-idbvamp)
VERSION := $(shell ./bytemine-idbvamp -version)

doc:
	go doc > README

distfile: doc
	rm -rf /tmp/bytemine-idbvamp-$(VERSION)
	mkdir /tmp/bytemine-idbvamp-$(VERSION)
	cp README bytemine-idbvamp /tmp/bytemine-idbvamp-$(VERSION)/
	cd /tmp && tar czfv /tmp/bytemine-idbvamp-$(VERSION).tgz \
		bytemine-idbvamp-$(VERSION)/
	sha256sum /tmp/bytemine-idbvamp-$(VERSION).tgz

upload: distfile
	scp /tmp/bytemine-idbvamp-$(VERSION).tgz \
		bytemine-www@appliance.bytemine.net:/data/www/allgemein/files.bytemine.net/

