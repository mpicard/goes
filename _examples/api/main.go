package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/gin-gonic/gin"
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

/*
TOREMOVE: old middlewares
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
		ctx := r.Context()

		// To remove... for mocking purpose
		r.Header.Set("authorization", "secrettoken")
		authHeader := r.Header.Get("authorization")

		if authHeader == "secrettoken" {
			ctx = context.WithValue(ctx, "authenticated_user", "z0mbie42")
		}
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

*/

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "authenticated_user", "z0mbie42")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func main() {
	app := &graph.Api{}
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(Auth())

	r.GET("/", gin.WrapH(handler.Playground("Api", "/graphql")))

	r.POST("/graphql", gin.WrapH(handler.GraphQL(graph.MakeExecutableSchema(app),
		handler.ResolverMiddleware(func(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
			token := ctx.Value("authenticated_user")
			_, ok := token.(string)
			if ok != true {
				return res, errors.New("Auth error")
			}
			res, err = next(ctx)
			return res, err
		}),
	)))

	log.Println("Listening on :8080")

	r.Run()
}
