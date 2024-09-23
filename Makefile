run:
	docker-compose up --build

redoc:
	swag init -g api.go --parseDependency --parseInternal