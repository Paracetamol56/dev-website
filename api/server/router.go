package server

import (
	"github.com/Paracetamol56/dev-website/api/controllers"
	"github.com/Paracetamol56/dev-website/api/middlewares"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Init controllers
	// auth := new(controllers.AuthController)
	contact := new(controllers.ContactController)
	heatlh := new(controllers.HealthController)
	user := new(controllers.UserController)

	apiGroup := r.Group("/api")
	{
		apiGroup.GET("/health", heatlh.GetHealth)
		apiGroup.POST("/contact", contact.PostContact)
		userGroup := apiGroup.Group("/user")
		{
			userGroup.Use(middlewares.JwtAuthMiddleware())
			userGroup.GET("/:id", user.GetUser)
			userGroup.PATCH("/:id", user.PatchUser)
			userGroup.DELETE("/:id", user.DeleteUser)
		}
	}

	// Static files
	r.Use(static.Serve("/", static.LocalFile("./build", true)))
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	return r
}
