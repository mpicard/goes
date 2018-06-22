package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/vektah/gqlgen/graphql"
	"github.com/vektah/gqlgen/handler"
	"github.com/z0mbie42/goes"
	"github.com/z0mbie42/goes/_examples/api/domain"
	"github.com/z0mbie42/goes/_examples/api/graph"
)

func init() {
	godotenv.Load()
	err := goes.InitDB(os.Getenv("DATABASE"), false)
	if err != nil {
		panic(err)
	}
	goes.MigrateEventsTable()

	todo := &domain.Todo{}
	goes.DB.DropTable(todo)
	goes.DB.AutoMigrate(todo)
	goes.RegisterEvents(domain.CreatedV1{}, domain.TextUpdatedV1{})
}

type middleware func(next http.HandlerFunc) http.HandlerFunc

func chainMiddleware(mw ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last(w, r)
		}
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "mycustomToken", "z0mbie42")
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func main() {
	app := &graph.Api{}
	graphQLChain := chainMiddleware(authMiddleware)

	http.Handle("/", handler.Playground("Api", "/graphql"))
	//http.Handle("/graphql", handler.GraphQL(graph.MakeExecutableSchema(app)))
	http.Handle("/graphql", graphQLChain(
		handler.GraphQL(graph.MakeExecutableSchema(app),
			handler.ResolverMiddleware(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
				token := ctx.Value("mycustomToken")
				_, ok := token.(string)
				if ok != true {
					return res, errors.New("Auth error")
				}
				res, err = next(ctx)
				return res, err
			}),
		)))

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
