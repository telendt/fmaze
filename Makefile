GO := go

pkgs = $(shell $(GO) list ./...)
src  = $(shell find . -name '*.go' -print)

build: vet style test fmaze

fmaze: $(src)
	@echo ">> building binaries"
	@$(GO) build -o $@ cmd/fmaze/main.go

fmaze-race: $(src)
	@echo ">> building race binaries"
	@$(GO) build -race -o $@ cmd/fmaze/main.go

test:
	@echo ">> running all tests"
	@$(GO) test -race -v $(pkgs)

vet:
	@echo ">> vetting code"
	@$(GO) vet $(pkgs)

style:
	@echo ">> checking code style"
	@! gofmt -d $(src) | grep '^'

clean:
	@echo ">> running cleanup"
	@rm -f fmaze fmaze-race coverage.txt profile.out

.PHONY: build test vet style
