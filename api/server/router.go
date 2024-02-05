package server

import (
	"github.com/Paracetamol56/dev-website/api/controllers"
	"github.com/Paracetamol56/dev-website/api/middlewares"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:8000", "http://localhost:5173", "https://dev.matheo-galuba.com", "https://dev-uat.matheo-galuba.com"},
		AllowMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))
	r.Use(func(c *gin.Context) {
		c.Header("Cross-Origin-Resource-Policy", "cross-origin")
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Embedder-Policy", "require-corp")
		c.Next()
	})

	// Init controllers
	auth := new(controllers.AuthController)
	contact := new(controllers.ContactController)
	heatlh := new(controllers.HealthController)
	hipparcos := new(controllers.HipparcosController)
	user := new(controllers.UserController)

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		apiGroup.GET("/health", heatlh.GetHealth)
		apiGroup.POST("/contact", contact.PostContact)
		authGroup := apiGroup.Group("/auth")
		{
			authGroup.POST("/login", auth.PostLogin)
			authGroup.POST("/verify", auth.PostVerify)
			authGroup.POST("/refresh", auth.PostRefresh)
			// Third party auth
			authGroup.POST("/github", auth.PostGithubLogin)
		}
		hipparcosGroup := apiGroup.Group("/hipparcos")
		{
			hipparcosGroup.GET("", hipparcos.GetHipparcosHR)
			hipparcosGroup.GET("/:hip", hipparcos.GetHipparcosHRByHIP)
		}
		userGroup := apiGroup.Group("/users")
		{
			userGroup.Use(middlewares.JwtAuthMiddleware())
			userGroup.GET("/:id", user.GetUser)
			userGroup.PATCH("/:id", user.PatchUser)
			userGroup.DELETE("/:id", user.DeleteUser)
			userGroup.GET("/:id/export", user.GetExport)
		}
	}

	// Static files
	r.Use(static.Serve("/", static.LocalFile("./build", true)))
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	return r
}
