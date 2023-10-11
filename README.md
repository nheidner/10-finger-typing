# 10 finger typing

`docker compose --profile dev up --build`: start dev environment
`docker compose --profile build up --build`: create and run build

`psql postgres://typing:password@localhost:54329/typing` to connect to database

curl <https://api.openai.com/v1/chat/completions> \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer sk-..." \
  -d '{
     "model": "gpt-3.5-turbo",
     "messages": [{"role": "user", "content": "Say this is a test!"}],
     "temperature": 0.7
   }'
