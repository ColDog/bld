test:
	go test -cover ./pkg/...

test-examples: install
	bld -v 10 -root-dir $$PWD/examples/node -spec $$PWD/examples/node/.bld.yaml

install:
	go install ./cmd/bld

save-schema:
	@pkg/builder/save-schema.sh
