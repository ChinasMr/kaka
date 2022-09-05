INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)

.PHONY: config
config:
	protoc --proto_path=./internal \
		   --proto_path=./third_party \
		   --go_out=paths=source_relative:./internal \
		   $(INTERNAL_PROTO_FILES)
