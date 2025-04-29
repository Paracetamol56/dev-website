package controllers

import "dev/internal/repositories"

type CodeCarbonController struct {
	projectRepo *repositories.CodeCarbonProjectRepository
	// experimentRepo *repositories.CodeCarbonExperimentRepository
	// tokenRepo *repositories.CodeCarbonTokenRepository
}

func NewCodeCarbonController() *CodeCarbonController {
	return &CodeCarbonController{
		projectRepo: repositories.NewCodeCarbonProjectRepository(),
		// TODO: experimentRepo
		// TODO: tokenRepo
	}
}
