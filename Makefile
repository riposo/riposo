default: test

.minimal.makefile:
	curl -fsSL -o $@ https://gitlab.com/bsm/misc/raw/master/make/go/minimal.makefile
include .minimal.makefile

test.integration.server:
	RIPOSO_ACCOUNT_CREATE_PRINCIPALS=system.Everyone \
	RIPOSO_BUCKET_CREATE_PRINCIPALS=system.Authenticated \
	RIPOSO_PLUGIN_DIR=../accounts/,../flush \
	go run cmd/riposo/main.go server

test.integration:
	bundle exec rspec -fd

rubocop:
	bundle exec rubocop --auto-correct

db.create:
	echo "CREATE DATABASE riposo_test ENCODING = 'UTF8'"  | psql -q postgres
db.drop:
	echo "DROP DATABASE riposo_test" | psql -q postgres

build: bin/riposo

bin/riposo: go.mod go.sum $(shell find . -name '*.go')
	@mkdir -p $(dir $@)
	go build -ldflags '-s -w' -o $@ cmd/riposo/main.go
