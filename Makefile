WD=$$(pwd)
DEV_MODE=1

.PHONY: all build

all: build

build:
	@go build -C $(WD)/grawpadmin -ldflags="-X 'main.ProjectRootPath=$(WD)' -X 'main.DeveloperMode=$(DEV_MODE)'" -o ../grawpa
