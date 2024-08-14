DB_URL = postgresql://buddhima:12345@localhost:5432/eduApp?sslmode=disable

postgres:
	docker run -d --name eduApp -p 5432:5432 -e POSTGRES_USER=buddhima -e POSTGRES_PASSWORD=12345 postgres:16-alpine

createdb:
	docker exec -it eduApp createdb --username=buddhima --owner=buddhima eduApp

migrate:
	migrate create -ext sql -dir db/migration -seq init_mg

dropdb:
	docker exec -it eduApp dropdb --username=buddhima eduApp

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

sqlc:
	sqlc generate

dbdocs:
	dbdocs build docs/db.dbml

redis:
	docker run --name redis -p 6379:6379 -d redis:alpine3.19

dbschema:
	dbml2sql --postgres -o docs/schema.sql docs/db.dbml

server:
	go run main.go

.PHONY: postgres createdb migrate dropdb migrateup migratedown sqlc server dbdocs dbschema