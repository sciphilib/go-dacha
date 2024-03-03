// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/ads": {
            "get": {
                "description": "Retrieves a list of all advertisements with detailed information",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "ads"
                ],
                "summary": "Get all ads",
                "responses": {
                    "200": {
                        "description": "An array of advertisement objects",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.AdResponse"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error"
                    }
                }
            }
        }
    },
    "definitions": {
        "models.AdResponse": {
            "type": "object",
            "properties": {
                "datetime": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "location": {
                    "description": "Предполагается, что Location - это структура с полями type и coordinates",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.LocationAd"
                        }
                    ]
                },
                "pictures": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "price": {
                    "type": "string"
                },
                "subcategory": {
                    "description": "Предполагается, что Subcategory - это структура с полями id, name и category",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.SubcategoryAd"
                        }
                    ]
                },
                "title": {
                    "type": "string"
                },
                "user": {
                    "description": "Предполагается, что User - это структура с полями id, name, email, phone_number, и location",
                    "allOf": [
                        {
                            "$ref": "#/definitions/models.UserAd"
                        }
                    ]
                }
            }
        },
        "models.LocationAd": {
            "type": "object",
            "properties": {
                "coordinates": {
                    "description": "Coordinates is an array of two float numbers.\nExample: [123.45, 67.89]",
                    "type": "array",
                    "items": {
                        "type": "number"
                    }
                },
                "type": {
                    "type": "string"
                }
            }
        },
        "models.SubcategoryAd": {
            "type": "object",
            "properties": {
                "category": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "models.UserAd": {
            "type": "object",
            "properties": {
                "email": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "location": {
                    "$ref": "#/definitions/models.LocationAd"
                },
                "name": {
                    "type": "string"
                },
                "phone_number": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}