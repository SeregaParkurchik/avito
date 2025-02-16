# avito
для запуска прокета:

docker-compose up

для запуска e2e теста для send_coin:

cd e2e/send_coin
docker-compose up
go test 

для запуска e2e теста для buy_item:

cd e2e/buy_item
docker-compose up
go test 