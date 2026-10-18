package server

import (
	"time"

	"dev/internal/admin"
	"dev/internal/auth"
	"dev/internal/chip"
	"dev/internal/contact"
	"dev/internal/health"
	"dev/internal/hipparcos"
	"dev/internal/icon"
	"dev/internal/middlewares"
	"dev/internal/ormi"
	"dev/internal/slides"
	"dev/internal/user"
	"dev/internal/utils"
	"dev/internal/wordcloud"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	// The reverse proxy reaches the API from a private network, clients cannot spoof X-Forwarded-For
	r.SetTrustedProxies([]string{"127.0.0.0/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", "::1/128", "fc00::/7"})
	r.Use(middlewares.RequestIdMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: utils.AllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	r.Use(middlewares.CORPMiddleware())

	adminController := new(admin.Controller)
	authController := new(auth.Controller)
	contactController := new(contact.Controller)
	healthController := new(health.Controller)
	hipparcosController := new(hipparcos.Controller)
	iconController := new(icon.Controller)
	wordCloud := new(wordcloud.Controller)
	chipController := new(chip.Controller)
	ormiController := new(ormi.Controller)
	slidesController := new(slides.Controller)
	userController := new(user.Controller)

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		apiGroup.GET("/health", healthController.GetHealth)
		apiGroup.POST("/contact", middlewares.RateLimitByIP(5, time.Hour), contactController.PostContact)
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.Use(middlewares.RateLimitByIP(30, time.Minute))
			authGroup.POST("/login", middlewares.RateLimitByIP(10, time.Hour), authController.PostLogin)
			authGroup.POST("/verify", authController.PostVerify)
			authGroup.POST("/refresh", authController.PostRefresh)
			authGroup.GET("/providers", authController.GetOAuthProviders)
			authGroup.POST("/:provider", authController.PostOAuthLogin)
			authGroup.POST("/passkey/begin", authController.PostPasskeyLoginBegin)
			authGroup.POST("/passkey/finish", authController.PostPasskeyLoginFinish)
		}
		hipparcosGroup := apiGroup.Group("/hipparcos")
		{
			hipparcosGroup.GET("", hipparcosController.GetHipparcosHR)
			hipparcosGroup.GET("/:hip", hipparcosController.GetHipparcosHRByHIP)
		}
		wordCloudGroup := apiGroup.Group("/word-cloud")
		{
			wordCloudGroup.GET("", wordCloud.GetWordCloud)
			wordCloudGroup.GET("/:id", wordCloud.GetWordCloudById)
			wordCloudGroup.GET("/:id/ws", wordCloud.WSWordCloud)
			// High ceiling: a whole audience can share one IP
			wordCloudGroup.POST("/:id/participant", middlewares.RateLimitByIP(300, time.Minute), wordCloud.PostWordCloudParticipant)
			wordCloudGroup.POST("/:id/word", middlewares.RateLimitByIP(300, time.Minute), wordCloud.PostWordCloudWord)
			wordCloudGroup.POST("", middlewares.JwtAuthMiddleware(), wordCloud.PostWordCloud)
			wordCloudGroup.PATCH("/:id", middlewares.JwtAuthMiddleware(), wordCloud.PatchWordCloud)
			wordCloudGroup.POST("/:id/duplicate", middlewares.JwtAuthMiddleware(), wordCloud.PostWordCloudDuplicate)
			wordCloudGroup.DELETE("/:id", middlewares.JwtAuthMiddleware(), wordCloud.DeleteWordCloud)
			wordCloudGroup.PATCH("/:id/words/:wordId", middlewares.JwtAuthMiddleware(), wordCloud.PatchWord)
			wordCloudGroup.POST("/:id/words/merge", middlewares.JwtAuthMiddleware(), wordCloud.PostMergeWords)
		}
		iconGroup := apiGroup.Group("/icons")
		{
			iconGroup.GET("", iconController.GetIcons)
			iconGroup.GET("/search", iconController.SearchIcons)
		}
		chipGroup := apiGroup.Group("/chips")
		{
			chipGroup.GET("", chipController.GetMicroprocessor)
			chipGroup.GET("/:id", chipController.GetMicroprocessorById)
		}
		ormiGroup := apiGroup.Group("/ormi")
		{
			ormiGroup.Use(middlewares.JwtAuthMiddleware())
			ormiGroup.GET("", ormiController.GetTodos)
			ormiGroup.GET("/stats", ormiController.GetTodoStats)
			ormiGroup.GET("/stats/:year", ormiController.GetTodoYearStats)
			ormiGroup.GET("/labels", ormiController.GetTodoLabels)
			ormiGroup.PUT("/order", ormiController.PutTodoOrder)
			ormiGroup.GET("/:id", ormiController.GetTodo)
			ormiGroup.POST("", ormiController.PostTodo)
			ormiGroup.PATCH("/:id", ormiController.PatchTodo)
			ormiGroup.DELETE("/:id", ormiController.DeleteTodo)
		}
		apiGroup.GET("/slides/:id/ws", slidesController.WSDeck)
		slidesGroup := apiGroup.Group("/slides")
		{
			slidesGroup.Use(middlewares.JwtAuthMiddleware())
			slidesGroup.GET("", slidesController.GetDecks)
			slidesGroup.POST("", slidesController.PostDeck)
			slidesGroup.GET("/:id", slidesController.GetDeck)
			slidesGroup.POST("/:id/duplicate", slidesController.PostDeckDuplicate)
			slidesGroup.DELETE("/:id", slidesController.DeleteDeck)
		}
		userGroup := apiGroup.Group("/users")
		{
			userGroup.Use(middlewares.JwtAuthMiddleware())
			userGroup.GET("/:id", userController.GetUser)
			userGroup.PATCH("/:id", userController.PatchUser)
			userGroup.DELETE("/:id", userController.DeleteUser)
			userGroup.GET("/:id/export", userController.GetExport)
			userGroup.DELETE("/:id/identities/:provider", userController.DeleteIdentity)
			userGroup.POST("/:id/passkeys/begin", authController.PostPasskeyRegisterBegin)
			userGroup.POST("/:id/passkeys/finish", authController.PostPasskeyRegisterFinish)
			userGroup.DELETE("/:id/passkeys/:passkeyId", authController.DeletePasskey)
		}
		adminGroup := apiGroup.Group("/admin")
		{
			adminGroup.Use(middlewares.JwtAuthMiddleware(), admin.RequireAdmin())
			adminGroup.GET("/users", adminController.GetUsers)
			adminGroup.GET("/users/:id", adminController.GetUser)
			adminGroup.PATCH("/users/:id/role", adminController.PatchUserRole)
			adminGroup.DELETE("/users/:id", adminController.DeleteUser)
			adminGroup.GET("/audit", adminController.GetAudit)
		}
	}

	// Static files
	r.Use(static.Serve("/", static.LocalFile("./build", false)))
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	return r
}
