SHELL:=/bin/bash

export PYTHONDONTWRITEBYTECODE := 1


help: ### Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
.PHONY: help


ci: lint typecheck test ### Run all checks
.PHONY: ci


lint: .venv ### Thoroughly lint source files
	( \
		source $(CURDIR)/.venv/bin/activate; \
		autoflake --in-place --recursive --remove-unused-variables --remove-all-unused-imports $(CURDIR)/instaproxy; \
		black $(CURDIR)/instaproxy; \
		flake8 --max-line-length 120 $(CURDIR)/instaproxy; \
	)
.PHONY: lint


test: .venv ### Run unit tests
	( \
		source $(CURDIR)/.venv/bin/activate; \
		pytest -p no:cacheprovider $(CURDIR)/ \
	)
.PHONY: test


typecheck: .venv ### Run mypy and its static type analyser
	( \
		source $(CURDIR)/.venv/bin/activate; \
		mypy --cache-dir=/dev/null $(CURDIR)/instaproxy \
	)
.PHONY: typecheck


.venv: requirements.txt requirements-dev.txt ### Initialise the virtual environment and install all dependencies
	( \
		python3 -m venv $(CURDIR)/.venv; \
		source $(CURDIR)/.venv/bin/activate; \
		pip3 install -r $(CURDIR)/requirements.txt -r $(CURDIR)/requirements-dev.txt; \
	);
	touch .venv;