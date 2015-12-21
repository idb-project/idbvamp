VERSION := 0.0.1

build:
	go build -o bytemine-idbvamp-$(VERSION)

distfile: build
	rm -rf /tmp/bytemine-idbvamp-$(VERSION)
	mkdir /tmp/bytemine-idbvamp-$(VERSION)
	cp bytemine-idbvamp-$(VERSION) /tmp/bytemine-idbvamp-$(VERSION)/bytemine-idbvamp-$(VERSION)
	cd /tmp && tar czfv /tmp/bytemine-idbvamp-$(VERSION).tgz \
		bytemine-idbvamp-$(VERSION)/
	sha256sum /tmp/bytemine-idbvamp-$(VERSION).tgz

upload: distfile
	scp /tmp/bytemine-idbvamp-$(VERSION).tgz \
		bytemine-www@appliance.bytemine.net:/data/www/allgemein/files.bytemine.net/

