.PHONY: all
all:
	@rm -rf /tmp/go-build*/
	@make mod
	@go generate ./...

.PHONY: test
test:
	@go test -cover -covermode=atomic -coverprofile=coverage.out -vet= -failfast -timeout=90s -count=3 -cpu=3 -parallel=3 --race ./...

.PHONY: cover
cover:
	@rm -rf /tmp/cover*/
	@go tool cover -html=coverage.out

.PHONY: mod
mod:
	@go mod vendor
	@go mod tidy

.PHONY: mod.upgrade
mod.upgrade:
	@go get -t -u ./...
	@make mod
	@go test all
