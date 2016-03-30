#VERSION := $(shell ./bytemine-idbvamp -version)
VENDOR=vendor
VENDORGIT=git clone --depth=1

vendors:
	rm -rf vendor
	mkdir -p vendor/git.office.bytemine.net/idb
	cd vendor/git.office.bytemine.net/idb && $(VENDORGIT) --branch v2 gitlab@git.office.bytemine.net:idb/idbclient.git && cd idbclient && rm -rf .git
	mkdir -p vendor/github.com/go-sql-driver
	cd vendor/github.com/go-sql-driver && $(VENDORGIT) https://github.com/go-sql-driver/mysql && cd mysql && rm -rf .git

build:
	go build -o bytemine-idbvamp

doc:
	go doc > README

distfile: doc build
	$(eval VERSION := $(shell ./bytemine-idbvamp -version))
	rm -rf /tmp/bytemine-idbvamp-$(VERSION)
	mkdir /tmp/bytemine-idbvamp-$(VERSION)
	cp README bytemine-idbvamp /tmp/bytemine-idbvamp-$(VERSION)/
	cd /tmp && tar czfv /tmp/bytemine-idbvamp-$(VERSION).tgz \
		bytemine-idbvamp-$(VERSION)/
	sha256sum /tmp/bytemine-idbvamp-$(VERSION).tgz

upload: distfile
	scp /tmp/bytemine-idbvamp-$(VERSION).tgz \
		bytemine-www@appliance.bytemine.net:/data/www/allgemein/files.bytemine.net/

