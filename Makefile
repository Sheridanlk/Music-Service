.SILENT:


migrate:
	migrate -path ./migrations -database 'postgres://postgres:postgres@localhost:5432/music_service?sslmode=disable' up