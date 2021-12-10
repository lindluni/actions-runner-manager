// Package docs GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "GitHub Professional Services",
            "url": "https://github.com/lindluni/actions-runner-manager",
            "email": "lindluni@github.com"
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
        "/group-create": {
            "post": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Creates a new GitHub Action organization runner group named with the team slug",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Create a new GitHub Action organization Runner Group",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Authorization token",
                        "name": "Authorization",
                        "in": "header",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/group-delete": {
            "delete": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Deletes an existing GitHub Action organization runner group named with the team slug",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "Deletes an existing GitHub Action organization Runner Group",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/group-list": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "List all repositories and runners assigned to a GitHub Action organization runner group named with the team slug",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Groups"
                ],
                "summary": "List all resources associated with a GitHub Action organization Runner Group",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "$ref": "#/definitions/apis.listResponse"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/repos-add": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Adds new repositories to an existing GitHub Actions organization named with the team slug",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Repos"
                ],
                "summary": "Add new repositories to an existing GitHub Actions organization runner group",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "description": "Comma-seperated list of repository slugs",
                        "name": "repos",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/repos-remove": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Removes existing repositories to an existing GitHub Actions organization named with the team slug",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Repos"
                ],
                "summary": "Remove existing repositories from an existing GitHub Actions organization runner group",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "description": "Comma-seperated list of repository slugs",
                        "name": "repos",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/repos-set": {
            "patch": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Replaces all existing repositories in an existing GitHub Actions organization named with the team slug with a new set of repositories",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Repos"
                ],
                "summary": "Replaces all existing repositories in an existing GitHub Actions organization runner group with a new set of repositories",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "array",
                        "items": {
                            "type": "string"
                        },
                        "description": "Comma-seperated list of repository slugs",
                        "name": "repos",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "type": "string"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/token-register": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Creates a new GitHub Action organization runner removal token that can be used to configure GitHub Action runners at the organization level",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tokens"
                ],
                "summary": "Create a new GitHub Action organization runner registration token",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "$ref": "#/definitions/github.RegistrationToken"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        },
        "/token-remove": {
            "get": {
                "security": [
                    {
                        "ApiKeyAuth": []
                    }
                ],
                "description": "Creates a new GitHub Action organization runner removal token that can be used remove a GitHub Action runners at the organization level",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Tokens"
                ],
                "summary": "Create a new GitHub Action organization runner removal token",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Canonical **slug** of the GitHub team",
                        "name": "team",
                        "in": "query",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "allOf": [
                                {
                                    "$ref": "#/definitions/apis.JSONResultSuccess"
                                },
                                {
                                    "type": "object",
                                    "properties": {
                                        "Code": {
                                            "type": "integer"
                                        },
                                        "Response": {
                                            "$ref": "#/definitions/github.RegistrationToken"
                                        }
                                    }
                                }
                            ]
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "apis.JSONResultSuccess": {
            "type": "object",
            "properties": {
                "Code": {
                    "type": "integer"
                },
                "Response": {}
            }
        },
        "apis.listResponse": {
            "type": "object",
            "properties": {
                "repos": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "runners": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "github.RegistrationToken": {
            "type": "object",
            "properties": {
                "expires_at": {
                    "$ref": "#/definitions/github.Timestamp"
                },
                "token": {
                    "type": "string"
                }
            }
        },
        "github.Timestamp": {
            "type": "object",
            "properties": {
                "time.Time": {
                    "type": "string"
                }
            }
        }
    },
    "securityDefinitions": {
        "APIKeyAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header"
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "0.1.0",
	Host:        "localhost",
	BasePath:    "/api/v1",
	Schemes:     []string{},
	Title:       "Action Runner Manager API",
	Description: "API for managing GitHub organization Runner Groups",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
		"escape": func(v interface{}) string {
			// escape tabs
			str := strings.Replace(v.(string), "\t", "\\t", -1)
			// replace " with \", and if that results in \\", replace that with \\\"
			str = strings.Replace(str, "\"", "\\\"", -1)
			return strings.Replace(str, "\\\\\"", "\\\\\\\"", -1)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register("swagger", &s{})
}
