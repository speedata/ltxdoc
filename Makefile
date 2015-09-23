
all:
	@echo "make [getdependencies | install | formatxml | validatexml | clean | bindata | local]"


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
	xmllint --relaxng schema/commandlist.rng ltxref.xml  --noout

clean:
	-rm -rf bin/ltxdoc pkg

bindata:
	bin/go-bindata -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml

local:
	bin/go-bindata -debug -o src/ltxdoc/bindata.go -pkg ltxdoc  -ignore=\\.DS_Store  httproot/... templates/... ltxref.xml
	go install ltxdoc/ltxdoc

