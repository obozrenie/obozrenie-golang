PKGNAME=github.com/obozrenie/obozrenie-golang

BINDATA_PKGNAME=main
FILES=assets/...

BINDATA_DIR=.
BINDATA_OUTFILE=$(BINDATA_DIR)/bindata.go

BUILD_DIR=./build
OUTFILE=$(BUILD_DIR)/obozrenie

build: bindata
	GOOS=linux CGO_ENABLED=0 go build -o $(OUTFILE)
run: build
	$(OUTFILE)

debug:
	dlv debug $(PKGNAME)

clean:
	rm -rf $(BUILD_DIR)/*
	rm $(BINDATA_OUTFILE)
bindata:
	go-bindata -pkg $(BINDATA_PKGNAME) -o $(BINDATA_OUTFILE) $(FILES)
