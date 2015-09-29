
all:
	@echo "make [getdependencies | install | formatxml | validatexml | clean | bindata | local]"
	@echo "'make install' installs a binary for your platform in the $GOBIN directory"


getdependencies:
	go get -u github.com/jteeuwen/go-bindata/...
	bin/go-bindata -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml
	go get ./...

install: bindata
	go install ltxdoc/ltxdoc

formatxml:
	xmllint --format ltxref.xml > temp.xml
	mv temp.xml ltxref.xml

validatexml:
	xmllint --relaxng src/github.com/speedata/ltxref/schema/commandlist.rng ltxref.xml  --noout

clean:
	-rm -rf bin/ltxdoc pkg

bindata:
	bin/go-bindata -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml
	cd src/github.com/speedata/ltxref;  make bindata

bindata-debug:
	bin/go-bindata -debug -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml
	cd src/github.com/speedata/ltxref;  make bindata-debug

local: bindata-debug
	go install ltxdoc/ltxdoc

buildwindows64: bindata
	GOOS=windows GOARCH=amd64 go build ltxdoc/ltxdoc

buildwindows32: bindata
	GOOS=windows GOARCH=386 go build ltxdoc/ltxdoc

buildlinux64: bindata
	GOOS=linux GOARCH=amd64 go build ltxdoc/ltxdoc

buildlinux32: bindata
	GOOS=linux GOARCH=386 go build ltxdoc/ltxdoc
