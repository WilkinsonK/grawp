.PHONY: all build

all: build

build:
	go build -C ./grawpadmin -o ../grawpa
