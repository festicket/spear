help:
	@echo "build - bild service from sources"
	@echo "install - install dependecies"
	@echo "up - run the service"

build: 
	GOOS=linux go build -o bin/spear src/spear/main.go

install:
	cd src/spear && dep ensure

up: 
	docker-compose up spear
