DB_DIR=./pkg/layoutstore/sqlite

SQLC=docker run --rm -v $(PWD):$(PWD) -w $(PWD) sqlc/sqlc:1.25.0
sqlc: schemadump
	$(SQLC) generate -f $(DB_DIR)/sqlc.yaml

schemadump:
	go run $(DB_DIR)/schemadump -path "$(DB_DIR)/schema.sql"

NOW := $(shell date +"%Y%m%d%H%M%S")
migration:
ifeq ($(name),)
	@echo "name is required"
else
	@echo "Creating migration $(name)"
	touch $(DB_DIR)/migrations/$(NOW)_$(name).up.sql
	touch $(DB_DIR)/migrations/$(NOW)_$(name).down.sql
endif

clean-deps:
	go mod tidy
	go mod vendor
