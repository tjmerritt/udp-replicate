
CMDS=udp-replicate read-test write-test

build:
	@for d in $(CMDS); do ( echo "Building $$d"; cd cmd/$$d; go build; ); done
