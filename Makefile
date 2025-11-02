.PHONY: all build

all: build

WD=$$(pwd)

build:
	go build -C $(WD)/grawpadmin -ldflags="-X 'main.ProjectRootPath=$(WD)'" -o ../grawpa
