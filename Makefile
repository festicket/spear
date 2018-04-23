help:
	@echo "build - bild service from sources"
	@echo "install - install dependecies"
	@echo "up - run the service"
	@echo "deploy - deply the latest built image to the registry"

build: 
	GOOS=linux go build -o bin/spear src/spear/*.go

install:
	cd src/spear && dep ensure

up: 
	docker-compose up spear

bup: build
	docker-compose stop spear
	docker-compose build spear
	docker-compose up spear

deploy:
	docker tag $(shell docker images --format="{{.ID}}" | head -n1) zerc/spear
	docker push zerc/spear
