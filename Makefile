.PHONY: all validate test audit report help

all: validate test

validate:
	bazel run //:validate

test:
	bazel test //...

audit: validate

report:
	cat validator_report.json

help:
	@echo "Enterprise Skill Builder - Bazel 9.2 Build Targets"
	@echo "================================================="
	@echo "  bazel run //:validate  (or make validate) - Run 5-point SDLC validator"
	@echo "  bazel test //...       (or make test)     - Run Bazel test suites"
	@echo "  make report                               - Display latest audit report"
