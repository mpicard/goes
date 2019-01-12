## Getting started

```bash
$ docker run -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres
$ export DATABASE_URL="postgres://postgres:mysecretpassword@localhost/?sslmode=disable"
$ psql $DATABASE_URL -c "CREATE DATABASE goes"
$ export DATABASE_URL="postgres://postgres:mysecretpassword@localhost/goes?sslmode=disable"
$ go run main.go
```
