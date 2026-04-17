.PHONY: check desktop-dev desktop-build

check:
	pwsh ./scripts/check.ps1

desktop-dev:
	wails dev

desktop-build:
	wails build -clean
