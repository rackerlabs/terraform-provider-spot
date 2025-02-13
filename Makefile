PROVIDER_NAME := spot
RENDERED_PROVIDER_NAME := "Rackspace Spot"
GOBIN := $(PWD)/bin
PATH := $(GOBIN):$(PATH)

.PHONY: generate generate-code scaffold-ds scaffold-rs scaffold-provider build install apply destroy clean uninstall test fmt lint check-versions docs timeouts-fix

generate:
	@go generate ./...

generate-code:
	@echo "Generating provider, resources, data-sources schema files from provider_code_spec.json..."
	tfplugingen-framework generate all --input ./provider_code_spec.json --output internal/provider
	@echo "Running script to fix timeout attrib issue..."
	@bash ./hacks/insert-timeouts-attribute.sh

install:
	@echo "Installing provider..."
	go install .

apply:
	@echo "Applying example terraform config..."
	cd _examples/simple-cloudspace && terraform apply -auto-approve && cd ../..

destroy:
	echo "Destroying example terraform config..."
	cd _examples/simple-cloudspace && terraform destroy -auto-approve && cd ../..

clean:
	@echo "Cleaning up..."
	rm -rf ~/go/bin/terraform-provider-$(PROVIDER_NAME)

uninstall:
	@echo "Uninstalling provider..."
	go clean -i .
	rm -f ~/go/bin/terraform-provider-$(PROVIDER_NAME)

test:
	@echo "Running tests..."
	go test ./...

fmt:
	@echo "Formatting code..."
	go fmt ./...

docs:
	@echo "Generating docs..."
	tfplugindocs generate --rendered-provider-name $(RENDERED_PROVIDER_NAME)

lint:
	@echo "Linting code..."
	go vet ./...
	terraform fmt -check=true -recursive examples/
	tfplugindocs validate

scaffold-ds:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Data-source name not provided. Usage: NAME=<data-source-name> make scaffold-ds"; \
		exit 1; \
	fi
	@echo "Scaffolding code for data-source $(NAME)..."
	tfplugingen-framework scaffold data-source --name $(NAME) --output-dir ./internal/provider --force

scaffold-rs:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: resource name not provided. Usage: NAME=<resource-name> make scaffold-rs"; \
		exit 1; \
	fi
	@echo "Scaffolding code for resource $(NAME)..."
	tfplugingen-framework scaffold resource --name $(NAME) --output-dir ./internal/provider --force

scaffold-provider:
	@echo "Scaffolding provider code..."
	tfplugingen-framework scaffold provider --name $(PROVIDER_NAME) --output-dir ./internal/provider --force

dependencies:
	@echo "Installing dependencies..."
	go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

check-versions:
	@echo "Checking Go version..."
	@echo "Go version must be at least 1.21.4"
	@echo "Checking Terraform version..."
	@echo "Terraform version must be at least 1.6.6"
