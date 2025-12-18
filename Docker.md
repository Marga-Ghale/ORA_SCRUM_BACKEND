docker system prune -a --volumes -f

migrate -path ./internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ora_scrum?sslmode=disable" force 3

docker exec -it ora_scrum_db psql -U postgres -d ora_scrum
