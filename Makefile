.PHONY: generate

# Ensure protoc plugins installed via `go install` are on PATH
export PATH := $(shell go env GOPATH)/bin:$(PATH)

generate:
	# Очищаем старую генерацию, чтобы избежать конфликтов
	rm -rf gen/
	mkdir -p gen
	# Генерируем код с путями по умолчанию (paths=import)
	protoc --proto_path=proto \
		--proto_path=third_party/googleapis \
		--go_out=gen \
		--go-grpc_out=gen \
		--grpc-gateway_out=gen \
		proto/*.proto
