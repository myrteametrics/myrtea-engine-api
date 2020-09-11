COMPONENT = assistant

.PHONY: build
build:
	protoc -I proto/ proto/$(COMPONENT).proto --go_out=plugins=grpc:proto/