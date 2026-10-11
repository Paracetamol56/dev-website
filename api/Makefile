PACKAGES = ./cmd/... ./internal/...
COVERED = ./internal/controllers,./internal/db,./internal/middlewares,./internal/models,./internal/server,./internal/utils
OUT = coverage

.PHONY: test test-e2e coverage coverage-e2e

test:
	go test $(PACKAGES)

test-e2e:
	go test -tags e2e ./internal/e2e

coverage:
	@mkdir -p $(OUT)
	go test $(PACKAGES) -coverprofile=$(OUT)/unit.raw
	@grep -v /internal/testutil/ $(OUT)/unit.raw > $(OUT)/unit.out
	@go tool cover -func=$(OUT)/unit.out | tail -1
	@go tool cover -html=$(OUT)/unit.out -o $(OUT)/unit.html

coverage-e2e:
	@mkdir -p $(OUT)
	go test -tags e2e ./internal/e2e -coverpkg=$(COVERED) -coverprofile=$(OUT)/e2e.out
	@go tool cover -func=$(OUT)/e2e.out | tail -1
	@go tool cover -html=$(OUT)/e2e.out -o $(OUT)/e2e.html
