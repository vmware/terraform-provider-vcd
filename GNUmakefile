TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=vcd

default: build

build: fmtcheck
	go install

dist:
	git archive --format=zip -o source.zip HEAD
	git archive --format=tar HEAD | gzip -c > source.tar.gz

install: build
	@$(CURDIR)/scripts/install-plugin.sh

test-binary-prepare: install
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short-provider"
	@sh -c "'$(CURDIR)/scripts/runtest.sh' binary-prepare"

test-binary: install
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short-provider"
	@sh -c "'$(CURDIR)/scripts/runtest.sh' binary"

testunit: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' unit"

test: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short"

testacc: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' acceptance"

testmulti: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' multiple"

testcatalog: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' catalog"

testvapp: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' vapp"

testvm: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' vm"

testgateway: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' gateway"

testnetwork: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' network"

testextnetwork: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' extnetwork"

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

errcheck:
	@sh -c "'$(CURDIR)/scripts/errcheck.sh'"

vendor-check:
	go mod tidy
	go mod vendor
	git diff --exit-code

test-compile:
	cd vcd && go test -tags ALL -c .

website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build test testacc vet fmt fmtcheck errcheck vendor-check test-compile website website-test

