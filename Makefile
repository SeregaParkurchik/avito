.PHONY: all up send_coin_test buy_item_test down

# Убедитесь, что у вас есть docker-compose и go

all: up

# Запуск основного проекта
up:
	docker-compose up

# Запуск e2e теста для send_coin (поднимает базу данных)
send_coin_test:
	cd e2e/send_coin && docker-compose up -d db && go test

# Запуск e2e теста для buy_item (поднимает базу данных)
buy_item_test:
	cd e2e/buy_item && docker-compose up -d db && go test

# Остановка всех контейнеров
down:
	docker-compose down
