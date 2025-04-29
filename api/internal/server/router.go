package server

import (
	"dev/internal/controllers"
	"dev/internal/middlewares"

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
	r.Use(middlewares.CORPMiddleware())

	// Init controllers
	auth := new(controllers.AuthController)
	codecarbon := controllers.NewCodeCarbonController()
	contact := new(controllers.ContactController)
	heatlh := new(controllers.HealthController)
	hipparcos := new(controllers.HipparcosController)
	wordCloud := new(controllers.WordCloudController)
	microprocessor := new(controllers.MicroprocessorController)
	user := new(controllers.UserController)

	apiGroup := r.Group("/api")
	{
		apiGroup.Use(middlewares.RequestIdMiddleware())
		apiGroup.GET("/doc/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		apiGroup.GET("/health", heatlh.GetHealth)
		codecarbonGroup := apiGroup.Group("/codecarbon")
		{
			codecarbonGroup.Use(middlewares.JwtAuthMiddleware())
			codecarbonGroup.GET("/projects", codecarbon.GetCodeCarbonProjects)
			codecarbonGroup.GET("/projects/:id", codecarbon.GetCodeCarbonProjectById)
			codecarbonGroup.POST("/projects", codecarbon.PostCodeCarbonProject)
			codecarbonGroup.PATCH("/projects/:id", codecarbon.PatchCodeCarbonProject)
			codecarbonGroup.DELETE("/projects/:id", codecarbon.DeleteCodeCarbonProject)
			tokenGroup := codecarbonGroup.Group("/projects/:id/tokens")
			{
				tokenGroup.GET("", codecarbon.GetCodeCarbonProjectTokens)
				tokenGroup.POST("", codecarbon.PostCodeCarbonProjectToken)
				tokenGroup.DELETE("/:token_id", codecarbon.DeleteCodeCarbonProjectToken)
			}
			experimentGroup := codecarbonGroup.Group("/projects/:id/experiments")
			{
				experimentGroup.GET("", codecarbon.GetCodeCarbonProjectExperiments)
				experimentGroup.POST("", codecarbon.PostCodeCarbonProjectExperiment)
				experimentGroup.PATCH("/:experiment_id", codecarbon.PatchCodeCarbonProjectExperiment)
				experimentGroup.DELETE("/:experiment_id", codecarbon.DeleteCodeCarbonProjectExperiment)
			}
			/*
				runGroup := codecarbonGroup.Group("/projects/:id/experiments/:experiment_id/runs")
				{
					runGroup.GET("", codecarbon.GetCodeCarbonProjectExperimentRuns)
					runGroup.GET("/:run_id", codecarbon.GetCodeCarbonProjectExperimentRunById)
					runGroup.POST("", codecarbon.PostCodeCarbonProjectExperimentRun)
					runGroup.DELETE("/:run_id", codecarbon.DeleteCodeCarbonProjectExperimentRun)
				}
				codecarbonGroup.GET("/experiments", codecarbon.GetCodeCarbonExperiments)
				codeCarbonGroup.GET("/experiments/:id", codecarbon.GetCodeCarbonExperimentById)
				codecarbonGroup.GET("/runs", controllers.GetCodeCarbonRuns)
				codecarbonGroup.GET("/runs/:id", controllers.GetCodeCarbonRunById)
			*/
		}
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
		wordCloudGroup := apiGroup.Group("/word-cloud")
		{
			wordCloudGroup.GET("", wordCloud.GetWordCloud)
			wordCloudGroup.GET("/:id", wordCloud.GetWordCloudById)
			wordCloudGroup.GET("/:id/ws", wordCloud.WSWordCloud)
			wordCloudGroup.POST("/:id/word", wordCloud.PostWordCloudWord)
			wordCloudGroup.POST("", middlewares.JwtAuthMiddleware(), wordCloud.PostWordCloud)
			wordCloudGroup.DELETE("/:id", middlewares.JwtAuthMiddleware(), wordCloud.DeleteWordCloud)
		}
		microprocessorGroup := apiGroup.Group("/microprocessors")
		{
			microprocessorGroup.GET("", microprocessor.GetMicroprocessor)
			microprocessorGroup.GET("/:id", microprocessor.GetMicroprocessorById)
		}
		userGroup := apiGroup.Group("/users")
		{
			userGroup.Use(middlewares.JwtAuthMiddleware())
			userGroup.GET("/:id", user.GetUser)
			userGroup.PATCH("/:id", user.PatchUser)
			userGroup.DELETE("/:id", user.DeleteUser)
			userGroup.GET("/:id/export", user.GetExport)
			// integrationGroup := userGroup.Group("/:id/integrations")
			/*
				{
					integrationGroup.GET("", user.GetUserIntegrations)
					integrationGroup.GET("/github", user.GetUserGithubIntegration)
					integrationGroup.DELETE("/github", user.DeleteUserGithubIntegration)
					integrationGroup.GET("/electricitymaps", user.GetUserElectricityMapsIntegration)
					integrationGroup.POST("/electricitymaps", user.PostUserElectricityMapsIntegration)
					integrationGroup.DELETE("/electricitymaps", user.DeleteUserElectricityMapsIntegration)
				}
			*/
		}
	}

	// Static files
	r.Use(static.Serve("/", static.LocalFile("./build", false)))
	r.NoRoute(func(c *gin.Context) {
		c.File("./build/index.html")
	})

	return r
}
