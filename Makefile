goose_up:
	goose -dir sql/schema postgres "postgres://postgres:test@localhost:5432/chirpy" up

goose_down:
	goose -dir sql/schema postgres "postgres://postgres:test@localhost:5432/chirpy" down

start: 
	go run cmd/api/main.go