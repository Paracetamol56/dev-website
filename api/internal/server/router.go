package server

import (
	"time"

	"dev/internal/controllers"
	"dev/internal/middlewares"
	"dev/internal/utils"

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
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: utils.AllowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	r.Use(middlewares.CORPMiddleware())

	// Init controllers
	auth := new(controllers.AuthController)
	contact := new(controllers.ContactController)
	heatlh := new(controllers.HealthController)
	hipparcos := new(controllers.HipparcosController)
	icon := new(controllers.IconController)
	wordCloud := new(controllers.WordCloudController)
	microprocessor := new(controllers.MicroprocessorController)
	ormi := new(controllers.OrmiController)
	passkey := new(controllers.PasskeyController)
	user := new(controllers.UserController)

	apiGroup := r.Group("/api")
	{
		apiGroup.Use(middlewares.RequestIdMiddleware())
		apiGroup.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		apiGroup.GET("/health", heatlh.GetHealth)
		apiGroup.POST("/contact", middlewares.RateLimitByIP(5, time.Hour), contact.PostContact)
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.Use(middlewares.RateLimitByIP(30, time.Minute))
			authGroup.POST("/login", middlewares.RateLimitByIP(10, time.Hour), auth.PostLogin)
			authGroup.POST("/verify", auth.PostVerify)
			authGroup.POST("/refresh", auth.PostRefresh)
			authGroup.GET("/providers", auth.GetOAuthProviders)
			authGroup.POST("/:provider", auth.PostOAuthLogin)
			authGroup.POST("/passkey/begin", passkey.PostLoginBegin)
			authGroup.POST("/passkey/finish", passkey.PostLoginFinish)
		}
		hipparcosGroup := apiGroup.Group("/hipparcos")
		{
			hipparcosGroup.GET("", hipparcos.GetHipparcosHR)
			hipparcosGroup.GET("/:hip", hipparcos.GetHipparcosHRByHIP)
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
			wordCloudGroup.DELETE("/:id", middlewares.JwtAuthMiddleware(), wordCloud.DeleteWordCloud)
		}
		iconGroup := apiGroup.Group("/icons")
		{
			iconGroup.GET("", icon.GetIcons)
			iconGroup.GET("/search", icon.SearchIcons)
		}
		microprocessorGroup := apiGroup.Group("/chips")
		{
			microprocessorGroup.GET("", microprocessor.GetMicroprocessor)
			microprocessorGroup.GET("/:id", microprocessor.GetMicroprocessorById)
		}
		ormiGroup := apiGroup.Group("/ormi")
		{
			ormiGroup.Use(middlewares.JwtAuthMiddleware())
			ormiGroup.GET("", ormi.GetTodos)
			ormiGroup.GET("/stats", ormi.GetTodoStats)
			ormiGroup.GET("/stats/:year", ormi.GetTodoYearStats)
			ormiGroup.GET("/labels", ormi.GetTodoLabels)
			ormiGroup.PUT("/order", ormi.PutTodoOrder)
			ormiGroup.GET("/:id", ormi.GetTodo)
			ormiGroup.POST("", ormi.PostTodo)
			ormiGroup.PATCH("/:id", ormi.PatchTodo)
			ormiGroup.DELETE("/:id", ormi.DeleteTodo)
		}
		userGroup := apiGroup.Group("/users")
		{
			userGroup.Use(middlewares.JwtAuthMiddleware())
			userGroup.GET("/:id", user.GetUser)
			userGroup.PATCH("/:id", user.PatchUser)
			userGroup.DELETE("/:id", user.DeleteUser)
			userGroup.GET("/:id/export", user.GetExport)
			userGroup.DELETE("/:id/identities/:provider", user.DeleteIdentity)
			userGroup.POST("/:id/passkeys/begin", passkey.PostRegisterBegin)
			userGroup.POST("/:id/passkeys/finish", passkey.PostRegisterFinish)
			userGroup.DELETE("/:id/passkeys/:passkeyId", passkey.DeletePasskey)
		}
	}

	// Static files
	r.Use(static.Serve("/", static.LocalFile("./build", false)))
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	return r
}
