.PHONY: all up send_coin_test buy_item_test

# Убедитесь, что у вас есть docker-compose и go

all: up

# Запуск основного проекта
up:
	docker-compose up

# Запуск e2e теста для send_coin
send_coin_test:
	cd e2e/send_coin && docker-compose up -d && go test

# Запуск e2e теста для buy_item
buy_item_test:
	cd e2e/buy_item && docker-compose up -d && go test

# Остановка всех контейнеров
down:
	docker-compose down
