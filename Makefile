watch:
	air

run:
	go run main.go

generate:
	oapi-codegen -generate types -o oapi/openapi_types.gen.go -package oapi oapi/model.yml
	oapi-codegen -generate gin -o oapi/openapi_server.gen.go -package oapi oapi/model.yml
	oapi-codegen -generate spec -o oapi/openapi_server.spec.go -package oapi oapi/model.yml
