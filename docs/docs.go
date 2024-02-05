// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "Matheo Galuba",
            "url": "https://dev.matheo-galuba.com/contact",
            "email": "matheo.galu56@gmail.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/auth/github/login": {
            "post": {
                "description": "Login or register a user with GitHub by code",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login or register a user with GitHub",
                "parameters": [
                    {
                        "description": "Code",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.GithubLoginBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    }
                }
            }
        },
        "/auth/login": {
            "post": {
                "description": "Login or register a user by email",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Login or register a user",
                "parameters": [
                    {
                        "description": "Email",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.LoginBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.FullUser"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    }
                }
            }
        },
        "/auth/refresh": {
            "post": {
                "description": "Refresh a user's access token by refresh token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Refresh a user's access token",
                "parameters": [
                    {
                        "description": "Refresh token",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.VerifyBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "string"
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    }
                }
            }
        },
        "/auth/verify": {
            "post": {
                "description": "Verify a user's email by token",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "auth"
                ],
                "summary": "Verify a user's email",
                "parameters": [
                    {
                        "description": "Token",
                        "name": "body",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.VerifyBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.FullUser"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    }
                }
            }
        },
        "/contact": {
            "post": {
                "description": "Send a contact message",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "contact"
                ],
                "summary": "Send a contact message",
                "parameters": [
                    {
                        "description": "Contact message",
                        "name": "contact",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.ContactBody"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    },
                    "400": {
                        "description": "Bad Request"
                    }
                }
            }
        },
        "/health": {
            "get": {
                "description": "Get health status",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Get health status",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/controllers.HealthResponse"
                        }
                    }
                }
            }
        },
        "/hipparcos/hr": {
            "get": {
                "description": "Get all Hipparcos HR stars",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "hipparcos"
                ],
                "summary": "Get all Hipparcos HR stars",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.HipparcosHR"
                            }
                        }
                    }
                }
            }
        },
        "/hipparcos/hr/{hip}": {
            "get": {
                "description": "Get one Hipparcos HR star by HIP number",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "hipparcos"
                ],
                "summary": "Get one Hipparcos HR star",
                "parameters": [
                    {
                        "type": "string",
                        "description": "HIP number",
                        "name": "hip",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.HipparcosHR"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    }
                }
            }
        },
        "/user/{id}": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Get one user by id",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get one user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.FullUser"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "404": {
                        "description": "Not Found"
                    }
                }
            },
            "delete": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Delete a user by id",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Delete a user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "404": {
                        "description": "Not Found"
                    }
                }
            },
            "patch": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Update a user by id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Update a user",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "User data",
                        "name": "patchUser",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/controllers.PatchUserBody"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.FullUser"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "404": {
                        "description": "Not Found"
                    }
                }
            }
        },
        "/user/{id}/export": {
            "get": {
                "security": [
                    {
                        "Bearer": []
                    }
                ],
                "description": "Get user export by id",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "user"
                ],
                "summary": "Get user export",
                "parameters": [
                    {
                        "type": "string",
                        "description": "User ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/models.FullUser"
                        }
                    },
                    "400": {
                        "description": "Bad Request"
                    },
                    "403": {
                        "description": "Forbidden"
                    },
                    "404": {
                        "description": "Not Found"
                    }
                }
            }
        }
    },
    "definitions": {
        "controllers.ContactBody": {
            "type": "object",
            "required": [
                "email",
                "message",
                "name"
            ],
            "properties": {
                "email": {
                    "type": "string"
                },
                "message": {
                    "type": "string",
                    "maxLength": 1000,
                    "minLength": 10
                },
                "name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 2
                },
                "userId": {
                    "type": "string"
                }
            }
        },
        "controllers.GithubLoginBody": {
            "type": "object",
            "required": [
                "code"
            ],
            "properties": {
                "code": {
                    "type": "string"
                }
            }
        },
        "controllers.HealthResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string",
                    "example": "ok"
                }
            }
        },
        "controllers.LoginBody": {
            "type": "object",
            "required": [
                "email"
            ],
            "properties": {
                "email": {
                    "type": "string"
                }
            }
        },
        "controllers.PatchUserBody": {
            "type": "object",
            "properties": {
                "flavour": {
                    "type": "string",
                    "enum": [
                        "latte",
                        "frappe",
                        "macchiato",
                        "mocha"
                    ]
                },
                "name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 2
                },
                "profilePicture": {
                    "type": "string"
                }
            }
        },
        "controllers.VerifyBody": {
            "type": "object",
            "required": [
                "token"
            ],
            "properties": {
                "token": {
                    "type": "string"
                }
            }
        },
        "models.FullUser": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "flavour": {
                    "type": "string"
                },
                "github": {
                    "$ref": "#/definitions/models.GitHubUser"
                },
                "githubAccessToken": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "lastLogin": {
                    "type": "string"
                },
                "lastRefresh": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "profilePicture": {
                    "type": "string"
                }
            }
        },
        "models.GitHubUser": {
            "type": "object",
            "properties": {
                "avatar_url": {
                    "type": "string"
                },
                "bio": {
                    "type": "string"
                },
                "blog": {
                    "type": "string"
                },
                "company": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "email": {
                    "type": "string"
                },
                "events_url": {
                    "type": "string"
                },
                "followers": {
                    "type": "integer"
                },
                "followers_url": {
                    "type": "string"
                },
                "following": {
                    "type": "integer"
                },
                "following_url": {
                    "type": "string"
                },
                "gists_url": {
                    "type": "string"
                },
                "gravatar_id": {
                    "type": "string"
                },
                "hireable": {
                    "type": "boolean"
                },
                "html_url": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "location": {
                    "type": "string"
                },
                "login": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "node_id": {
                    "type": "string"
                },
                "organizations_url": {
                    "type": "string"
                },
                "public_gists": {
                    "type": "integer"
                },
                "public_repos": {
                    "type": "integer"
                },
                "received_events_url": {
                    "type": "string"
                },
                "repos_url": {
                    "type": "string"
                },
                "site_admin": {
                    "type": "boolean"
                },
                "starred_url": {
                    "type": "string"
                },
                "subscriptions_url": {
                    "type": "string"
                },
                "twitter_username": {
                    "type": "string"
                },
                "type": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "models.HipparcosHR": {
            "type": "object",
            "properties": {
                "Amag": {
                    "type": "number"
                },
                "BV": {
                    "type": "number"
                },
                "hip": {
                    "type": "integer"
                }
            }
        }
    },
    "securityDefinitions": {
        "Bearer": {
            "description": "Bearer token",
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/api",
	Schemes:          []string{},
	Title:            "Dev Website API",
	Description:      "This is the API for my personal dev website",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
