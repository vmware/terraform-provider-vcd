TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)
WEBSITE_REPO=github.com/hashicorp/terraform-website
PKG_NAME=vcd

default: build

# builds the plugin
build: fmtcheck
	go install

# builds the plugin with race detector enabled
buildrace: fmtcheck
	go install --race

# creates a .zip archive of the code
dist:
	git archive --format=zip -o source.zip HEAD
	git archive --format=tar HEAD | gzip -c > source.tar.gz

# builds and deploys the plugin
install: build
	@sh -c "'$(CURDIR)/scripts/install-plugin.sh'"

# builds and deploys the plugin with race detector enabled (useful for troubleshooting)
installrace: buildrace
	@sh -c "'$(CURDIR)/scripts/install-plugin.sh'"

# makes .tf files from test templates
test-binary-prepare: install
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short-provider"
	@sh -c "'$(CURDIR)/scripts/runtest.sh' binary-prepare"

# runs test using Terraform binary as Org user
test-binary-orguser: install
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short-provider-orguser"
	@sh -c "'$(CURDIR)/scripts/runtest.sh' binary"

# runs upgrade test using Terraform binary
test-upgrade:
	@sh -c "'$(CURDIR)/scripts/test-upgrade.sh'"

# makes .tf files from test templates for upgrade testing, but does not execute them
test-upgrade-prepare:
	@sh -c "skip_upgrade_execution=1 '$(CURDIR)/scripts/test-upgrade.sh'"

# runs test using Terraform binary as system administrator using binary with race detection enabled
test-binary: installrace
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short-provider"
	@sh -c "'$(CURDIR)/scripts/runtest.sh' binary"

# prepares to build the environment in a new vCD (run once before test-env-apply)
test-env-init: install
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short-provider"
	@sh -c "'$(CURDIR)/scripts/runtest.sh' test-env-init"

# build the environment in a new vCD (run once before testacc)
test-env-apply:
	@sh -c "'$(CURDIR)/scripts/runtest.sh' test-env-apply"

# destroys the environment built with 'test-env-apply' (warning: can't be undone)
test-env-destroy:
	@sh -c "'$(CURDIR)/scripts/runtest.sh' test-env-destroy"

# Retrieves the authorization token
token:
	@sh -c "'$(CURDIR)/scripts/runtest.sh' token"

# runs staticcheck
static: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' static"

# runs the unit tests
testunit: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' unit"

# Runs the basic execution test
test: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' short"

# Runs the full acceptance test as Org user
testacc-orguser: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' acceptance-orguser"

# Runs the full acceptance test as system administrator
testacc: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' acceptance"

# Runs the acceptance test as system administrator for search label
test-search: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' search"

# Runs full acceptance test sequentially (using "-parallel 1" flag for go test)
testacc-race-seq: testunit
	@sh -c "'$(CURDIR)/scripts/runtest.sh' sequential-acceptance"

# Runs the acceptance test with tag 'multiple'
testmulti: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' multiple"

# Runs the acceptance test for org
testorg: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' org"

# Runs the acceptance test for catalog
testcatalog: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' catalog"

# Runs the acceptance test for vapp
testvapp: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' vapp"

# Runs the acceptance test for lb
testlb: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' lb"

# Runs the acceptance test for user
testuser: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' user"

# Runs the acceptance test for vm
testvm: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' vm"

# Runs the acceptance test for gateway
testgateway: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' gateway"

# Runs the acceptance test for network
testnetwork: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' network"

# Runs the acceptance test for external network
testextnetwork: fmtcheck
	@sh -c "'$(CURDIR)/scripts/runtest.sh' extnetwork"

# vets all .go files
vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -ne 0 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

# formats all .go files
fmt:
	gofmt -w $(GOFMT_FILES)

# runs a Go format check
fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

# runs the vendor directory check
vendor-check:
	go mod tidy
	go mod vendor
	git diff --exit-code

# checks that the code can compile
test-compile:
	cd vcd && go test -race -tags ALL -c .

# builds the website and allows running it from localhost
website:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

# tests the website files for broken link
website-test:
ifeq (,$(wildcard $(GOPATH)/src/$(WEBSITE_REPO)))
	echo "$(WEBSITE_REPO) not found in your GOPATH (necessary for layouts and assets), get-ting..."
	git clone https://$(WEBSITE_REPO) $(GOPATH)/src/$(WEBSITE_REPO)
endif
	@$(MAKE) -C $(GOPATH)/src/$(WEBSITE_REPO) website-provider-test PROVIDER_PATH=$(shell pwd) PROVIDER_NAME=$(PKG_NAME)

.PHONY: build test testacc-race-seq testacc vet static fmt fmtcheck vendor-check test-compile website website-test

