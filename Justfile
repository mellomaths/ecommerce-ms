
run:
    go run cmd/*.go

build:
    go build -o ecomm ./cmd

exec:
    ./ecomm

sqlc:
    sqlc generate

pg-migration-up:
    cd internal/adapters/postgresql/migrations && goose postgres postgres://postgres:postgres@192.168.1.100:5432/ecomm up

pg-migration-down:
    cd internal/adapters/postgresql/migrations && goose postgres postgres://postgres:postgres@192.168.1.100:5432/ecomm down

create-migration NAME:
    goose -s create {{NAME}} sql
