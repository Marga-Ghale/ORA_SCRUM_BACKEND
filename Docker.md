docker system prune -a --volumes -f

migrate -path ./internal/db/migrations -database "postgresql://postgres:postgres@localhost:5432/ora_scrum?sslmode=disable" force 3

docker exec -it ora_scrum_db psql -U postgres -d ora_scrum

docker image prune -a && docker builder prune -a

docker exec -it ora_scrum_db psql -U postgres -d ora_scrum -c "
DELETE FROM tasks;
DELETE FROM sprints;
DELETE FROM labels;
DELETE FROM project_members;
DELETE FROM projects;
DELETE FROM folder_members;
DELETE FROM folders;
DELETE FROM space_members;
DELETE FROM spaces;
DELETE FROM workspace_members;
DELETE FROM workspaces;
DELETE FROM users;
"

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

POST http://localhost:8080/api/members/project/b44744d1-1d29-434e-a2d2-64ff9eb08dcd
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyMjI4MjQsImlhdCI6MTc2NjEzNjQyNCwic3ViIjoiMTA3MjY1NzEtOGQxYi00ZTlhLTk2ZDMtNmZhZTBiZmYyODVhIn0.IC5mLDwED_2M6TmEvxKlBOCXhnI5TCQ9QWTZF3UOmX8
Content-Type: application/json

{
"userId": "5df3f708-ece9-4ada-9636-1609e67dd5ad",
"role": "member"
}

curl -X POST http://localhost:8080/api/members/project/b44744d1-1d29-434e-a2d2-64ff9eb08dcd \
 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyMjI4MjQsImlhdCI6MTc2NjEzNjQyNCwic3ViIjoiMTA3MjY1NzEtOGQxYi00ZTlhLTk2ZDMtNmZhZTBiZmYyODVhIn0.IC5mLDwED_2M6TmEvxKlBOCXhnI5TCQ9QWTZF3UOmX8" \
 -H "Content-Type: application/json" \
 -d '{
"userId": "5df3f708-ece9-4ada-9636-1609e67dd5ad",
"role": "member"
}'

curl -X GET http://localhost:8080/api/members/project/e6ce2752-403f-4b7d-9433-3702246d4951/direct \
 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyNDEwNTAsImlhdCI6MTc2NjE1NDY1MCwic3ViIjoiMTA3MjY1NzEtOGQxYi00ZTlhLTk2ZDMtNmZhZTBiZmYyODVhIn0.10namnxwbQ05Qnlbuq_r0RiWkxwnycVKOPRUH0M6D_4"

curl -X GET "http://localhost:8080/api/members/project/e6ce2752-403f-4b7d-9433-3702246d4951/access?userId=10726571-8d1b-4e9a-96d3-6fae0bff285a" \
 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyNDEwNTAsImlhdCI6MTc2NjE1NDY1MCwic3ViIjoiMTA3MjY1NzEtOGQxYi00ZTlhLTk2ZDMtNmZhZTBiZmYyODVhIn0.10namnxwbQ05Qnlbuq_r0RiWkxwnycVKOPRUH0M6D_4"
"

curl -X POST http://localhost:8080/api/members/project/e6ce2752-403f-4b7d-9433-3702246d4951 \
 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyNDEwNTAsImlhdCI6MTc2NjE1NDY1MCwic3ViIjoiMTA3MjY1NzEtOGQxYi00ZTlhLTk2ZDMtNmZhZTBiZmYyODVhIn0.10namnxwbQ05Qnlbuq_r0RiWkxwnycVKOPRUH0M6D_4" \
 -H "Content-Type: application/json" \
 -d '{
"userId": "5df3f708-ece9-4ada-9636-1609e67dd5ad",
"role": "member"
}'

curl -X POST http://localhost:8080/api/tasks/67d7f3fa-8926-4d3b-8e23-87668d0b0d18/comments \
 -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjYyNDUwNTAsImlhdCI6MTc2NjE1ODY1MCwic3ViIjoiOTM5YzQ0ODYtOTM2OS00NDhjLWJiNjEtOGJlNjc4MjBlYzQ0In0.U8x7z_m8Hd2rmkQ8WMpLEE_Qy0YfDuVTO4aQBHiG_7Y" \
 -H "Content-Type: application/json" \
 -d '{
"content": "I can help with this!",
"mentionedUsers": ["10726571-8d1b-4e9a-96d3-6fae0bff285a"]
}'
