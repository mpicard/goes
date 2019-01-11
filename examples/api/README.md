This project use [goes](https://github.com/z0mbie42/goes) as event sourcing framework and [gqlgen](https://github.com/vektah/gqlgen) as a typed GraphQL API layer.

## Getting started

```bash
$ docker run -p 5432:5432 -e POSTGRES_PASSWORD=mysecretpassword -d postgres
$ export DATABASE="postgres://postgres:mysecretpassword@localhost/?sslmode=disable"
$ psql $DATABASE -c "CREATE DATABASE goes"
$ export DATABASE="postgres://postgres:mysecretpassword@localhost/goes?sslmode=disable"
$ go get -u
$ go run main.go
```

then go to [localhost:8080](http://localhost:8080) and run
```graphql
mutation {
  create_todo(input: {text: "todo", author: "z0mbie42"}) {
    id
  }
}
```
then
```graphql
query {
  todos {
    id
    text
    author {name}
    events {
      id
      data {
        ... on TodoTextUpdatedV1 {
          text
        }
        ... on TodoCreatedV1 {
          id
          text
          author_name
        }
      }
    }
  }
}
```


## How

The project is divided in 2 subpackages: `domain` and `graph` which are 2 layers.

The package `domain` contains all the domain logic (Aggregates, commands, events...). It's our domain layer and wrap all our application logic.

The package `graph` contains all the GraphQL API related things. It's our query layer.


## Resources:

* https://github.com/z0mbie42/goes
* https://github.com/vektah/gqlgen
* https://outcrawl.com/go-graphql-gateway-microservices/
* https://kickstarter.engineering/event-sourcing-made-simple-4a2625113224
