all: build up

build:
		docker-compose build
		docker build -t bench ./benchmark

up:
		docker-compose up

down:
		docker-compose down

bench:
		docker run bench

.PHONY: build up down bench