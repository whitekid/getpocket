.PHONY: test dep tidy
test:
	@go test ./...

dep:
	@rm -f go.mod go.sum
	@go mod init github.com/whitekid/getpocket

	@$(MAKE) tidy

tidy:
	@go mod tidy -v
