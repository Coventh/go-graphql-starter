package main

import (
	"github.com/OscarYuen/go-graphql-starter/config"
	"github.com/OscarYuen/go-graphql-starter/handler"
	"github.com/OscarYuen/go-graphql-starter/model"
	"github.com/OscarYuen/go-graphql-starter/resolver"
	"github.com/OscarYuen/go-graphql-starter/schema"
	"github.com/OscarYuen/go-graphql-starter/service"
	"log"
	"net/http"

	graphql "github.com/neelance/graphql-go"
	relay "github.com/neelance/graphql-go/relay"
	"golang.org/x/net/context"
)

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/home" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "notification.html")
}

func main() {
	db, err := config.OpenDB("test.db")
	if err != nil {
		log.Fatal("Unable to connect to db:")
		log.Fatal(err)
	}
	notificationHub := model.NewNotificationHub()
	go notificationHub.Run()
	ctx := context.WithValue(context.Background(), "userService", service.NewUserService(db))
	ctx = context.WithValue(ctx, "authService", service.NewAuthService())
	ctx = context.WithValue(ctx, "notificationHub", notificationHub)

	graphqlSchema := graphql.MustParseSchema(schema.GetRootSchema(), &resolver.Resolver{})

	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "graphiql.html")
	}))

	http.Handle("/login", handler.Login(ctx))

	http.Handle("/ws", handler.SimpleMiddleware(ctx, handler.WebSocket(notificationHub)))

	http.HandleFunc("/home", serveHome)

	http.Handle("/query", handler.Authenticate(ctx, &relay.Handler{Schema: graphqlSchema}))

	log.Fatal(http.ListenAndServe(":3000", nil))
}
