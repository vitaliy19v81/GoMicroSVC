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
        "/api/message": {
            "post": {
                "description": "Создает новое сообщение и сохраняет его в базе данных",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Api"
                ],
                "summary": "Создание сообщения",
                "parameters": [
                    {
                        "description": "Сообщение",
                        "name": "message",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.CreateMessageRequest"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/models.CreateMessageResponse"
                        }
                    },
                    "400": {
                        "description": "Неверный формат данных",
                        "schema": {
                            "type": "string"
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/messages": {
            "get": {
                "description": "Возвращает список сообщений с учетом offset и limit",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Api"
                ],
                "summary": "Получение списка сообщений",
                "parameters": [
                    {
                        "type": "integer",
                        "default": 0,
                        "description": "Смещение",
                        "name": "offset",
                        "in": "query"
                    },
                    {
                        "type": "integer",
                        "default": 10,
                        "description": "Лимит",
                        "name": "limit",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/models.Message"
                            }
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/api/stats": {
            "get": {
                "description": "Получает количество обработанных сообщений",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "Api"
                ],
                "summary": "Получение статистики",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": {
                                "type": "integer"
                            }
                        }
                    },
                    "500": {
                        "description": "Ошибка сервера",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "gorm.DeletedAt": {
            "type": "object",
            "properties": {
                "time": {
                    "type": "string"
                },
                "valid": {
                    "description": "Valid is true if Time is not NULL",
                    "type": "boolean"
                }
            }
        },
        "models.CreateMessageRequest": {
            "type": "object",
            "properties": {
                "content": {
                    "description": "validate:\"required\"` + "`" + `",
                    "type": "string"
                }
            }
        },
        "models.CreateMessageResponse": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "id": {
                    "type": "integer"
                },
                "processed": {
                    "type": "boolean"
                }
            }
        },
        "models.Message": {
            "type": "object",
            "properties": {
                "content": {
                    "type": "string"
                },
                "createdAt": {
                    "type": "string"
                },
                "deletedAt": {
                    "$ref": "#/definitions/gorm.DeletedAt"
                },
                "id": {
                    "type": "integer"
                },
                "processed": {
                    "description": "Устанавливаем значение по умолчанию для processed",
                    "type": "boolean"
                },
                "updatedAt": {
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
