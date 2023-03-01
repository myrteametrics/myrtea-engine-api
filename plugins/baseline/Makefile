COMPONENT = baseline
COMPONENT_U = Baseline

# go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

.PHONY: build
build:
	protoc -I proto/ proto/$(COMPONENT).proto --go_out=. --go-grpc_out=.

# protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/baseline.proto