project=hyperon
org=github.com/hyperon

workspace=$(shell pwd)
gosrc=$$GOPATH/src

orgdir=$(gosrc)/$(org)
builddir=$(gosrc)/$(org)/$(project)
buildflags="GO15VENDOREXPERIMENT=1 CGO_ENABLED=0"

.PHONY:default init build run clean

default:run

init:
	-mkdir -p $(orgdir)
	-rm   -rf $(builddir)
	ln -s $(workspace) $(builddir)

build:
	@cd $(builddir); $(buildflag) go build

run:
	@cd $(builddir); $(buildflag) go run main.go

clean:
	@cd $(builddir); $(buildflag) go clean
