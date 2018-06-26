test:
	go test -cover ./pkg/...

install:
	go install ./cmd/bld


save-schema:
	@pkg/builder/save-schema.sh
