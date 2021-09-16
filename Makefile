default: test

.minimal.makefile:
	curl -fsSL -o $@ https://gitlab.com/bsm/misc/raw/master/make/go/minimal.makefile
include .minimal.makefile

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
	go build -ldflags '-s -w' -o $@ cmd/riposo/*.go
