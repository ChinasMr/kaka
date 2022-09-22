INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
API_PROTO_FILES=$(shell find api -name *.proto)



.PHONY: config
config:
	protoc --proto_path=./internal \
		   --proto_path=./third_party \
		   --go_out=paths=source_relative:./internal \
		   $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./api \
 	       --go-http_out=paths=source_relative:./api \
 	       --go-grpc_out=paths=source_relative:./api \
	       $(API_PROTO_FILES)