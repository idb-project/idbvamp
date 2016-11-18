NAME=idbvamp
VENDORGIT=git clone --depth=1

vendors:
	rm -rf vendor
	mkdir -p vendor/github.com/go-sql-driver
	cd vendor/github.com/go-sql-driver && $(VENDORGIT) https://github.com/go-sql-driver/mysql && cd mysql && rm -rf .git

doc:
	godoc . | head -n -4 > README

build:
	go install

distfile: build doc
	$(eval VERSION := $(shell $(GOPATH)/bin/$(NAME) -version))
	rm -rf /tmp/bytemine-$(NAME)-$(VERSION)
	mkdir /tmp/bytemine-$(NAME)-$(VERSION)
	cp $(GOPATH)/bin/$(NAME) /tmp/bytemine-$(NAME)-$(VERSION)/bytemine-$(NAME)
	cp README /tmp/bytemine-$(NAME)-$(VERSION)/
	cd /tmp && tar czfv /tmp/bytemine-$(NAME)-$(VERSION).tgz bytemine-$(NAME)-$(VERSION)/
	sha256sum /tmp/bytemine-$(NAME)-$(VERSION).tgz
