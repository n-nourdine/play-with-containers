package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/n-nourdine/play-with-containers/api-gateway/rabbitmq"
)

type Handler struct {
	Logger    *log.Logger
	Publisher *rabbitmq.Publisher
}

type BillingRequest struct {
	UserID        string `json:"user_id"`
	NumberOfItems string `json:"number_of_items"`
	TotalAmount   string `json:"total_amount"`
}

func NewHandler(logger *log.Logger) (*Handler, error) {
	// Create RabbitMQ publisher
	publisher, err := rabbitmq.NewPublisher(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
	}

	return &Handler{
		Logger:    logger,
		Publisher: publisher,
	}, nil
}

func (h *Handler) Close() {
	if h.Publisher != nil {
		h.Publisher.Close()
	}
}

// ProxyToInventory forwards all requests to the inventory service
func (h *Handler) ProxyToInventory(w http.ResponseWriter, r *http.Request) {
	// Get inventory service URL
	inventoryURL := fmt.Sprintf("http://%s:%s",
		os.Getenv("INVENTORY_SERVICE_HOST"),
		os.Getenv("INVENTORY_SERVICE_PORT"))

	// Create the target URL
	targetURL := inventoryURL + r.URL.Path
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	h.Logger.Printf("Proxying %s request to: %s", r.Method, targetURL)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Read request body
	var body io.Reader
	if r.Body != nil {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			h.Logger.Printf("Error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}
		body = bytes.NewReader(bodyBytes)
	}

	// Create new request
	req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, body)
	if err != nil {
		h.Logger.Printf("Error creating request: %v", err)
		http.Error(w, "Error creating request", http.StatusInternalServerError)
		return
	}

	// Copy headers
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Make the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		h.Logger.Printf("Error making request to inventory service: %v", err)
		http.Error(w, "Inventory service unavailable", http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set status code
	w.WriteHeader(resp.StatusCode)

	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		h.Logger.Printf("Error copying response body: %v", err)
	}

	h.Logger.Printf("Proxied request completed with status: %d", resp.StatusCode)
}

// HandleBilling processes billing requests and sends them to RabbitMQ
func (h *Handler) HandleBilling(w http.ResponseWriter, r *http.Request) {
	h.Logger.Println("Received billing request")

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.Logger.Printf("Error reading billing request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	// Validate JSON structure
	var billingReq BillingRequest
	if err := json.Unmarshal(body, &billingReq); err != nil {
		h.Logger.Printf("Error parsing billing JSON: %v", err)
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if billingReq.UserID == "" || billingReq.NumberOfItems == "" || billingReq.TotalAmount == "" {
		h.Logger.Printf("Missing required fields in billing request: %+v", billingReq)
		http.Error(w, "Missing required fields: user_id, number_of_items, total_amount", http.StatusBadRequest)
		return
	}

	// Send message to RabbitMQ
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	err = h.Publisher.PublishBillingMessage(ctx, string(body))
	if err != nil {
		h.Logger.Printf("Error publishing billing message: %v", err)
		http.Error(w, "Error processing billing request", http.StatusInternalServerError)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"message": "Message posted successfully",
		"status":  "accepted",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.Logger.Printf("Error encoding response: %v", err)
	}

	h.Logger.Printf("Billing message published successfully for user: %s", billingReq.UserID)
}

// ServeOpenAPIDoc serves the OpenAPI documentation
func (h *Handler) ServeOpenAPIDoc(w http.ResponseWriter, r *http.Request) {
	openAPISpec := `{
  "openapi": "3.0.0",
  "info": {
    "title": "Movie Streaming Platform API Gateway",
    "description": "API Gateway for a microservices-based movie streaming platform. Routes requests to inventory and billing services.",
    "version": "1.0.0",
    "contact": {
      "name": "API Support",
      "email": "support@movieplatform.com"
    }
  },
  "servers": [
    {
      "url": "http://localhost:3000",
      "description": "Development server"
    }
  ],
  "paths": {
    "/api/health": {
      "get": {
        "summary": "Health check endpoint",
        "description": "Returns the health status of the API Gateway",
        "responses": {
          "200": {
            "description": "Service is healthy",
            "content": {
              "text/plain": {
                "schema": {
                  "type": "string",
                  "example": "API Gateway is healthy"
                }
              }
            }
          }
        }
      }
    },
    "/api/movies": {
      "get": {
        "summary": "Get all movies",
        "description": "Retrieve all movies from the inventory. Supports filtering by title.",
        "parameters": [
          {
            "name": "title",
            "in": "query",
            "description": "Filter movies by title",
            "required": false,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "List of movies",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "$ref": "#/components/schemas/Movie"
                  }
                }
              }
            }
          }
        }
      },
      "post": {
        "summary": "Create a new movie",
        "description": "Add a new movie to the inventory",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "required": ["title"],
                "properties": {
                  "title": {
                    "type": "string",
                    "description": "Movie title"
                  },
                  "description": {
                    "type": "string",
                    "description": "Movie description"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Movie created successfully",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Movie"
                }
              }
            }
          }
        }
      },
      "delete": {
        "summary": "Delete all movies",
        "description": "Delete all movies from the inventory. Requires confirmation header.",
        "parameters": [
          {
            "name": "Confirm-Delete",
            "in": "header",
            "required": true,
            "schema": {
              "type": "string",
              "enum": ["yes"]
            },
            "description": "Confirmation header required to delete all movies"
          }
        ],
        "responses": {
          "204": {
            "description": "All movies deleted successfully"
          }
        }
      }
    },
    "/api/movies/{id}": {
      "get": {
        "summary": "Get movie by ID",
        "description": "Retrieve a specific movie by its ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Movie details",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/Movie"
                }
              }
            }
          }
        }
      },
      "put": {
        "summary": "Update movie by ID",
        "description": "Update an existing movie",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "title": {
                    "type": "string"
                  },
                  "description": {
                    "type": "string"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Movie updated successfully"
          }
        }
      },
      "delete": {
        "summary": "Delete movie by ID",
        "description": "Delete a specific movie by its ID",
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "schema": {
              "type": "string"
            }
          }
        ],
        "responses": {
          "204": {
            "description": "Movie deleted successfully"
          }
        }
      }
    },
    "/api/billing": {
      "post": {
        "summary": "Process billing request",
        "description": "Submit a billing order that will be processed asynchronously via RabbitMQ",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/BillingRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Billing request accepted",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "message": {
                      "type": "string",
                      "example": "Message posted successfully"
                    },
                    "status": {
                      "type": "string",
                      "example": "accepted"
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "Movie": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "description": "Unique movie identifier"
          },
          "title": {
            "type": "string",
            "description": "Movie title"
          },
          "description": {
            "type": "string",
            "description": "Movie description"
          }
        }
      },
      "BillingRequest": {
        "type": "object",
        "required": ["user_id", "number_of_items", "total_amount"],
        "properties": {
          "user_id": {
            "type": "string",
            "description": "ID of the user making the order"
          },
          "number_of_items": {
            "type": "string",
            "description": "Number of items in the order"
          },
          "total_amount": {
            "type": "string",
            "description": "Total cost of the order"
          }
        }
      }
    }
  }
}`

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(openAPISpec))
}

// ServeSwaggerUI serves a simple Swagger UI interface
func (h *Handler) ServeSwaggerUI(w http.ResponseWriter, r *http.Request) {
	swaggerHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Movie Platform API Gateway</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/docs',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(swaggerHTML))
}
