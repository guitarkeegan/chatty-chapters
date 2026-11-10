# Music Tutor

## Description

This application will serve as a supplement to my online music courses, starting Fall 2024.

## MVP requirements

- api deployed to EC2
- authentication/authorization will be handled by Cognito
- Postgres setup with RDS
- Possibly Lambdas used from triggers between Cognito and the API or RDS

- students must be able to...
- login
- find their course textbook
- find and select the current chapter
- ask questions about the current chapter, for their understanding
- have ai generate and grade quizzes based on the chapter reading material and/or music
- students can access and play musical examples from the app
- ai in the background to assess the student/ai interactions to insure grading fairness
approprateness of the chat, and quality of student questions and interactions

### Other wishlist requirements

- integration with instructure api, to automate grade updates
- other ideas that can allow students to experience/learn the material more activly

# Getting Started

Make sure to turn off postgres if you have it running locally or else change the 
port info below. I believe it is ```hostport:containerport```

1. Clone the project
2. Install Docker
3. Pull the postgres/pgvector image from Dockerhub
```docker pull pgvector/pgvector:pg16```
4. In the terminal, list all images. You should see only one. 
```docker image ls```
5. Copy the id onto your clip board. It should be on the left and look something like 
ceb82277bed1
6. This command will create and run the 'image' and create a new 'container' running in the
background on your computer. The container is basically a mini Linux OS running Postgres. 
```
docker run --name music-postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=music
-p 5432:5432 -d <the-id-from-your-clipboard>
``` 
7. You can list the running containers with 
```docker ps```
8. If you want to trash it ever, do ```docker ps```, then
```
docker stop <id-of-container>
docker rm <id-of-container>
```
9. The database is now setup. If for any reason you want to interact with the db directly,
the command below will sign you into the container and into the database.
You can type ```exit``` to exit the container or ```\q```
```psql -h localhost -U postgres -d music```
or use the Makefile
```
make db/connect
```

## Install Go Dependancies
1. I believe ```go get``` will install everything you need.

## Fire up the server

1. in the root directory, this command will show you the flags that I've setup.
```go run ./cmd/api -help```
2. Seed the db
```go run ./cmd/api -seed_embeddings=true```
3. Run migrations (you'll need to have 'migrate' installed. You can get it via Homebrew.
```https://formulae.brew.sh/formula/golang-migrate```
These commands do the same thing...
```migrate -seq -ext=.sql -dir=./migrations postgres://postgres:password@localhost:5432/music\?sslmode=disable up```
or use the Makefile script
```make db/up```
4. Start the server
```go run ./cmd/api/``` or use ```go run/front``` if you are using the NextJS frontend
The logs should tell you the rest!

## Create a test user

When the server is running, send a curl request with the new user info. 

```
curl -X POST http://localhost:8080/v1/register/users \
-H "Content-Type: application/json" \
-d '[
  {
    "cognito_sub": "",
    "first_name": "",
    "last_name": "",
    "email": "",
    "course_id": 1
  }
]'
```
### .env file

Copy `.env.example` to `.env` and fill in your own values.
Optionally, you can also set ```DEBUG=1``` if you want to see some debug logs from the server.

## Sample data

The only book included is *Alice's Adventures in Wonderland* by Lewis Carroll, from
[Project Gutenberg](https://www.gutenberg.org/ebooks/11) (public domain). Each chapter is a
text file in `books/`, named `title-author-edition-chapter.txt`.

```
go run ./cmd/api -seed_all=true         # demo users, the book, chapters, and a course
go run ./cmd/api -seed_embeddings=true  # chunk and embed books/ into pgvector (uses OpenAI)
```
