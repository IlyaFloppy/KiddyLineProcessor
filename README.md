# Kiddy Line Processor

Все параметры настраиваются через файл .env:
```.env
DB_HOST=kiddy-postgres
DB_USER=kiddy_postgres_user
DB_PASSWORD=kiddy_postgres_password
DB_NAME=kiddy
DB_PORT=5432
KLP_HTTP_ADDRESS="" # empty for all ips
KLP_HTTP_PORT=8080
KLP_GRPC_ADDRESS="" # empty for all ips
KLP_GRPC_PORT=9090
KLP_LOG_LEVEL=info # debug | info | warn | error | panic | fatal
GIN_MODE=debug # debug | release
FETCH_ADDRESS=kiddy-lines-provider
FETCH_PORT=8000
FETCH_SPORTS="{\"baseball\": 3, \"football\": 4, \"soccer\": 5}" # sports and update intervals(seconds) as json
```