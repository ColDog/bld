build:
	@docker run --rm -i \
		-v /tmp/:/tmp/ \
		-v $$PWD:/mnt \
		-v /var/run/docker.sock:/var/run/docker.sock \
		--workdir /mnt \
		coldog/bld:latest
