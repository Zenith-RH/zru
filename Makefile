TARGET 		= zru
SRC 		= main.go

.PHONY: all
all: compile

.PHONY: help
help:
	@echo "\033[34mzru (Zenith Release Utility) targets:\033[0m"
	@perl -nle'print $& if m{^[a-zA-Z_-\d]+:.*?## .*$$}' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

.PHONY: compile
compile: ## compile the project
	@go build -o $(TARGET) $(SRC)
	@echo "[+] Build completed"

.PHONY: clean
clean: ## clean up the project directory
	@rm -f $(TARGET)

.PHONY: docker
docker: ## build a local docker image
	@docker build --network host . -t zru:latest

.PHONY: install
install: compile ## install the application locally
	@mv $(TARGET) /usr/local/bin/
	@echo "[+] Successfully installed to /usr/local/bin/$(TARGET)"

.PHONY: pull
pull:
	@git pull

.PHONY: update
update: pull tidy install ## pulls new version and installs it

.PHONE: tidy
tidy: ## runs gofmt & go mod tidy
	@go mod tidy
	@gofmt -s -w .
