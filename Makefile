SQLC=docker run --rm -v $(PWD):$(PWD) -w $(PWD) sqlc/sqlc:1.25.0
sqlc:
	$(SQLC) generate -f pkg/layoutstore/sqlite/sqlc.yaml


NOW := $(shell date +"%Y%m%d%H%M%S")
migration:
ifeq ($(name),)
	@echo "name is required"
else
	@echo "Creating migration $(name)"
	touch internal/db/migrations/$(NOW)_$(name).up.sql
	touch internal/db/migrations/$(NOW)_$(name).down.sql
endif
