COMPONENT = baseline
COMPONENT_U = Baseline

.PHONY: build
build:
	protoc -I proto/ proto/$(COMPONENT).proto --go_out=plugins=grpc:proto/