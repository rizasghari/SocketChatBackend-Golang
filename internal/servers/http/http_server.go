package http

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
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
	router        *gin.Engine
	restHandler   *handlers.RestHandler
	htmlHandler   *handlers.HtmlHandler
	socketHandler *handlers.SocketHandler
	redis         *redis.Client
	ctx           context.Context
}

func NewHttpServer(
	ctx context.Context,
	redis *redis.Client,
	restHandler *handlers.RestHandler,
	socketHandler *handlers.SocketHandler,
	htmlHandler *handlers.HtmlHandler,
) *HttpServer {
	once.Do(func() {
		httpServer = &HttpServer{
			restHandler:   restHandler,
			redis:         redis,
			ctx:           ctx,
			socketHandler: socketHandler,
			htmlHandler:   htmlHandler,
		}
	})
	return httpServer
}

func (hs *HttpServer) Run() {
	hs.initializeGin()
	hs.setupWebSocketRoutes()
	hs.setupRestfulRoutes()
	hs.socketHandler.StartSocket()
	server := hs.startServer()
	// Wait for interrupt signal to gracefully shut down the server
	hs.socketHandler.WaitForShutdown(server)
}

func (hs *HttpServer) initializeGin() {
	hs.router = gin.Default()

	hs.router.Static("/web/static", "./web/static")

	ginHtmlRenderer := hs.router.HTMLRender
	hs.router.HTMLRender = &handlers.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Disable trusted proxy warning.
	err := hs.router.SetTrustedProxies(nil)
	if err != nil {
		return
	}
}

func (hs *HttpServer) setupRestfulRoutes() {
	// Handle no route found
	hs.router.NoRoute(hs.htmlHandler.NotFound)

	// Apply the CORS middleware to the router
    hs.router.Use(handlers.CORSMiddleware())

	web := hs.router.Group("/")
	{
		web.GET("/", hs.htmlHandler.Index)
		web.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	v1 := hs.router.Group("/api/v1")
	{
		v1.POST("/login", hs.restHandler.Login)
		v1.POST("/register", hs.restHandler.Register)
	}

	authenticated := v1.Group("/")
	authenticated.Use(handlers.MustAuthenticateMiddleware())
	{
		authenticated.GET("/users", hs.restHandler.GetAllUsersWithPagination)
		authenticated.GET("/users/:id", hs.restHandler.GetSingleUser)
		authenticated.POST("/users/upload-profile-photo", hs.restHandler.UploadUserProfilePhoto)
		authenticated.PUT("/users", hs.restHandler.UpdateUser)
		authenticated.GET("/users/discover", hs.restHandler.DiscoverUsers)

		authenticated.POST("/conversations", hs.restHandler.CreateConversation)
		authenticated.GET("/conversations/user/:id", hs.restHandler.GetUserConversations)
		authenticated.GET("/conversations/my", hs.restHandler.GetUserConversationsByToken)

		authenticated.POST("/messages", hs.restHandler.SaveMessage)
		authenticated.GET("/messages/conversation/:id", hs.restHandler.GetMessagesByConversationID)
	}
}

func (hs *HttpServer) setupWebSocketRoutes() {
	hs.router.GET("/ws", hs.socketHandler.HandleSocketRoute)
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
