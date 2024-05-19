run:
	docker-compose up -d --build

stop:
	docker-compose down

start:
	go run main.go