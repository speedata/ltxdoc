
all:
	@echo "make [getdependencies | install | formatxml | validatexml]"

getdependencies:
	go get ./...

install:
	go install ltxdoc/ltxdoc

formatxml:
	xmllint --format ltxref.xml > temp.xml
	mv temp.xml ltxref.xml

validatexml:
	xmllint --relaxng schema/commandlist.rng ltxref.xml  --noout

clean:
	-rm -rf bin/ltxdoc pkg
