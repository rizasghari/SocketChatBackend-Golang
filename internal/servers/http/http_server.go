package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"socketChat/internal/handlers"
	"sync"
)

var (
	httpServer *HttpServer
	once       sync.Once
)

type HttpServer struct {
	router  *gin.Engine
	handler *handlers.Handler
	socket  *handlers.Socket
	redis   *redis.Client
	ctx     context.Context
}

func NewHttpServer(ctx context.Context, redis *redis.Client, handler *handlers.Handler) *HttpServer {
	once.Do(func() {
		httpServer = &HttpServer{
			handler: handler,
			redis:   redis,
			ctx:     ctx,
			socket:  handlers.NewSocket(redis, ctx),
		}
	})
	return httpServer
}

func (hs *HttpServer) Run() {
	hs.initializeGin()
	hs.setupWebSocketRoutes()
	hs.setupRestfulRoutes()
	hs.socket.StartSocket()
	server := hs.startServer()
	// Wait for interrupt signal to gracefully shut down the server
	hs.socket.WaitForShutdown(server)
}

func (hs *HttpServer) initializeGin() {
	hs.router = gin.Default()
	hs.router.LoadHTMLGlob("./*.html")
}

func (hs *HttpServer) setupRestfulRoutes() {
	web := hs.router.Group("/")
	{
		web.GET("/", hs.handler.Index)
	}

	v1 := hs.router.Group("/api/v1")
	{
		v1.POST("/login", hs.handler.Login)
		v1.POST("/register", hs.handler.Register)
	}

	v1_authenticated := v1.Group("/protected")
	v1_authenticated.Use(hs.handler.MustAuthenticateMiddleware())
	{
		v1_authenticated.POST("/conversations", hs.handler.CreateConversation)
	}
}

func (hs *HttpServer) setupWebSocketRoutes() {
	hs.router.GET("/ws", hs.socket.HandleSocketRoute)
}

func (hs *HttpServer) startServer() *http.Server {
	server := &http.Server{
		Addr:    ":8000",
		Handler: hs.router,
	}

	go func() {
		log.Println("HTTP server started on :8000")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	return server
}
