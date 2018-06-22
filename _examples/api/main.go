package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/vektah/gqlgen/handler"
	"github.com/z0mbie42/goes"
	"github.com/z0mbie42/goes/_examples/api/domain"
	"github.com/z0mbie42/goes/_examples/api/graph"
)

func init() {
	godotenv.Load()
	err := goes.InitDB(os.Getenv("DATABASE"), true)
	if err != nil {
		panic(err)
	}
	goes.MigrateEventsTable()

	todo := &domain.Todo{}
	goes.DB.DropTable(todo)
	goes.DB.AutoMigrate(todo)
	goes.RegisterEvents(domain.CreatedV1{}, domain.TextUpdatedV1{})
}

func main() {
	app := &graph.Api{}
	http.Handle("/", handler.Playground("Api", "/graphql"))
	http.Handle("/graphql", handler.GraphQL(graph.MakeExecutableSchema(app)))

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
