.PHONY: generate

# Ensure protoc plugins installed via `go install` are on PATH
export PATH := $(shell go env GOPATH)/bin:$(PATH)

# Well-known types (descriptor.proto, timestamp.proto, …)
PROTOC_SYS_INCLUDES :=
ifneq ($(wildcard /usr/include/google/protobuf),)
PROTOC_SYS_INCLUDES += --proto_path=/usr/include
endif
ifneq ($(wildcard /opt/homebrew/include/google/protobuf),)
PROTOC_SYS_INCLUDES += --proto_path=/opt/homebrew/include
endif
ifneq ($(wildcard /usr/local/include/google/protobuf),)
PROTOC_SYS_INCLUDES += --proto_path=/usr/local/include
endif

generate:
	# Очищаем старую генерацию, чтобы избежать конфликтов
	rm -rf gen/
	mkdir -p gen
	# Генерируем код с путями по умолчанию (paths=import)
	protoc --proto_path=proto \
		--proto_path=third_party/googleapis \
		$(PROTOC_SYS_INCLUDES) \
		--go_out=gen \
		--go-grpc_out=gen \
		--grpc-gateway_out=gen \
		proto/*.proto
