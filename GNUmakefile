TEST?=$$(go list ./...)
GOFMT_FILES?=$$(find . -name '*.go')
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

test: fmtcheck
	go test -i $(TEST) || exit 1
	cd vcd ; VCD_SHORT_TEST=1 go test -v . -timeout 3m

testacc: fmtcheck
	if [ ! -f vcd/vcd_test_config.json -a -z "${VCD_CONFIG}" ] ; then \
		echo "ERROR: test configuration file vcd/vcd_test_config.json is missing"; \
		exit 1; \
	fi
	cd vcd ; TF_ACC=1 go test -v . -timeout 60m

vet:
	@echo "go vet ."
	@go vet $$(go list ./...) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)

website:
	@$(CURDIR)/scripts/web-site.sh $(WEBSITE_REPO) $(shell pwd) $(PKG_NAME) website-provider

website-test:
	@$(CURDIR)/scripts/web-site.sh $(WEBSITE_REPO) $(shell pwd) $(PKG_NAME) website-provider-test

.PHONY: build test testacc vet fmt fmtcheck test-compile website website-test

