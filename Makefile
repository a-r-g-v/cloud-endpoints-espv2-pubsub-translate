.PHONY: protoc
protoc:
	rm -rf ./pb
	mkdir -p ./pb/testv1

	protoc \
		-I.:$(CURDIR)/protobuf/lib \
		--go_out=./pb/testv1 \
		--go-grpc_out=./pb/testv1 \
		--validate_out='lang=go:./pb/testv1' \
		--include_imports \
		--include_source_info \
		--descriptor_set_out=pb/api_descriptor.pb \
		./protobuf/src/*.proto
