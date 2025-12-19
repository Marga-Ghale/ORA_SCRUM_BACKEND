docker system prune -a --volumes -f

migrate -path ./internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ora_scrum?sslmode=disable" force 3

docker exec -it ora_scrum_db psql -U postgres -d ora_scrum

docker image prune -a && docker builder prune -a


curl -X POST http://localhost:8080/api/spaces/c5712068-db10-4087-84e1-8bbb9fa16288/projects \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyMTc3NTEsImlhdCI6MTc2NjEzMTM1MSwic3ViIjoiZWFmYmYzZjUtMjBhYS00YTgxLTlhNzEtMzA3ODQ4MmVmZDEyIn0.MHuaXw0NIOUw05hst_mQ4mndJ8A5zSYY5oFIdjONEy0" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ORA Mobile App",
    "key": "OMH",
    "description": "Mobile app for ORA Scrum",
    "folderId": "7e9b031-a6ee-4be0-baa2-60859b55c476",
    "icon": "📱",
    "color": "#EC4899"
  }'


curl -X POST http://localhost:8080/api/spaces/c5712068-db10-4087-84e1-8bbb9fa16288/projects \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyMTc3NTEsImlhdCI6MTc2NjEzMTM1MSwic3ViIjoiZWFmYmYzZjUtMjBhYS00YTgxLTlhNzEtMzA3ODQ4MmVmZDEyIn0.MHuaXw0NIOUw05hst_mQ4mndJ8A5zSYY5oFIdjONEy0" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ORA Mobile App",
    "key": "OMH",
    "folderId": "97e9b031-a6ee-4be0-baa2-60859b55c476",
    "description": "Mobile app for ORA Scrum",
    "icon": "📱",
    "color": "#EC4899"
  }'
