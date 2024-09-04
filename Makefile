













# Validate a provider
check: bin/validate
	@if [ $(SOURCE) == "" ]; then echo 'Missing $$(SOURCE)'; exit 1; fi
	bin/validate check --source $(SOURCE)

bin/validate: bin
	cd tools/validate && go build -o ../../bin/validate .

.PHONY: bin/validate

bin:
	mkdir -p bin
