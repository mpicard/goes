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
	"github.com/bloom42/goes"
	"github.com/bloom42/goes/examples/api/domain"
	"github.com/bloom42/goes/examples/api/graph"
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

func auth() gin.HandlerFunc {
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
	r.Use(auth())

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
