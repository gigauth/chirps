goose_up:
	goose postgres "postgres://postgres:test@localhost:5432/chirpy" up

goose_down:
	goose postgres "postgres://postgres:test@localhost:5432/chirpy" down