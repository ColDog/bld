test:
	go test -cover ./pkg/...

test-examples: install
	@echo "testing examples/node"
	@docker run --rm -it \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v /tmp/:/tmp \
		-v $$PWD/examples/node:/mnt \
		--workdir /mnt \
		coldog/bld:latest

install:
	@echo "installing..."
	@go install ./cmd/bld
	@bld

save-schema:
	@pkg/builder/save-schema.sh
