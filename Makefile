mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(patsubst %/,%,$(dir $(mkfile_path)))

GOPATH := $(current_dir)
GOBIN := $(current_dir)/bin

export GOBIN GOPATH

.PHONY: help

help:
	@echo "Run make install"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


install: bindata ## Install the software in bin/ltxdoc
	go install ltxdoc/ltxdoc

local: bindata-debug ## Internal use
	go install ltxdoc/ltxdoc

formatxml: ## re-format the XML file, needs xmllint
	xmllint --format ltxref.xml > temp.xml
	mv temp.xml ltxref.xml

validatexml: ## Check if the XML is valid, needs xmllint
	xmllint --relaxng src/github.com/speedata/ltxref/schema/commandlist.rng ltxref.xml  --noout

clean: ## Remove intermediate files
	-rm -rf $(GOBIN)/ltxdoc pkg

# Create go file for assets
bindata:
	if [ -e "insertatend.txt" ] ; then \
		sed -i.tmp -e  '/attheend/rinsertatend.txt' templates/layout.html; \
	fi
	$(GOBIN)/go-bindata -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml ; \
	if [ -e "insertatend.txt" ] ; then \
		cp templates/layout.html.tmp templates/layout.html ; \
	fi
	cd src/github.com/speedata/ltxref;  make bindata

# Don't include the asset files in the binary
bindata-debug:
	$(GOBIN)/go-bindata -debug -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml
	cd src/github.com/speedata/ltxref;  make bindata-debug

buildwindows64: bindata
	GOOS=windows GOARCH=amd64 go build ltxdoc/ltxdoc

buildwindows32: bindata
	GOOS=windows GOARCH=386 go build ltxdoc/ltxdoc

buildlinux64: bindata
	GOOS=linux GOARCH=amd64 go build ltxdoc/ltxdoc

buildlinux32: bindata
	GOOS=linux GOARCH=386 go build ltxdoc/ltxdoc
