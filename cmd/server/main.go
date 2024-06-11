package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/mrangelba/go-exp-clean-arch/configs"
	"github.com/mrangelba/go-exp-clean-arch/internal/di"
	"github.com/mrangelba/go-exp-clean-arch/internal/event/handler"
	"github.com/mrangelba/go-exp-clean-arch/internal/infra/graph"
	"github.com/mrangelba/go-exp-clean-arch/internal/infra/grpc/pb"
	"github.com/mrangelba/go-exp-clean-arch/internal/infra/grpc/service"
	"github.com/mrangelba/go-exp-clean-arch/pkg/events"
	"github.com/streadway/amqp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	defer db.Close()

	driver, _ := mysql.WithInstance(db, &mysql.Config{})
	migrations, err := migrate.NewWithDatabaseInstance("file://migrations", "mysql", driver)

	println("Migration: start")
	if err != nil {
		panic(err)
	}
	err = migrations.Up()

	if err == migrate.ErrNoChange {
		println("Migration: no change detected")
	} else {
		println("Migration: done")
	}

	rabbitMQChannel := getRabbitMQChannel(configs)

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated",
		&handler.OrderCreatedHandler{
			RabbitMQChannel: rabbitMQChannel,
		})

	createOrderUseCase := di.NewCreateOrderUseCase(db, eventDispatcher)
	listOrderUseCase := di.NewListOrderUseCase(db, eventDispatcher)

	listener, err := net.Listen("tcp", configs.WebServerPort)
	if err != nil {
		log.Fatalf("Error create listen tcp %v", err)
	}

	m := cmux.New(listener)
	httpL := m.Match(cmux.HTTP1Fast())
	grpcL := m.Match(cmux.HTTP2())

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	srv := graphql_handler.NewDefaultServer(
		graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
			CreateOrderUseCase: *createOrderUseCase,
			ListOrderUseCase:   *listOrderUseCase,
		}}),
	)

	// GraphQL
	router.Get("/graphql", playground.Handler("GraphQL playground", "/graphql/query"))
	router.Handle("/graphql/query", srv)

	// REST
	webOrderHandler := di.NewWebOrderHandler(db, eventDispatcher)
	router.Post("/order", webOrderHandler.Create)
	router.Get("/order", webOrderHandler.List)

	// gRPC
	grpcServer := grpc.NewServer()
	orderService := service.NewOrderService(*createOrderUseCase, *listOrderUseCase)
	pb.RegisterOrderServiceServer(grpcServer, orderService)
	reflection.Register(grpcServer)

	g := new(errgroup.Group)

	g.Go(func() error {
		return http.Serve(httpL, router)
	})

	g.Go(func() error {
		return grpcServer.Serve(grpcL)
	})

	g.Go(func() error {
		println("Server running on port 8080")
		return m.Serve()
	})

	log.Fatalf("Run server: %v", g.Wait())
}

func getRabbitMQChannel(conf *configs.Conf) *amqp.Channel {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", conf.RabbitMQUser, conf.RabbitMQPassword, conf.RabbitMQHost, conf.RabbitMQPort))
	if err != nil {
		panic(err)
	}
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}
