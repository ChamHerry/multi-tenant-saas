.PHONY: dev web build

dev:
	go run .

web:
	cd web && npm run dev

build:
	cd web && npm run build
	gf build -ew
