package main

import (
	"fmt"
	"net/http"

	"github.com/normastars/frame"
	"github.com/normastars/frame/example/version"
)

// User represents a user entity for CRUD examples.
type User struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Item represents a paginated item in the list example.
type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	app := frame.New()

	// ── Basic routes ──────────────────────────────────────────────
	app.GET("/hello", Hello)
	app.GET("/user/:id", GetUser)
	app.POST("/user", CreateUser)
	app.PUT("/user/:id", UpdateUser)
	app.DELETE("/user/:id", DeleteUser)
	app.PATCH("/user/:id", PatchUser)
	app.HEAD("/health", HealthCheck)

	// ── Route group ───────────────────────────────────────────────
	v1 := app.Group("/api/v1")
	{
		v1.GET("/items", ListItems)
		v1.POST("/items", CreateItem)
	}

	// Startup info
	fmt.Printf("commit: %20s\n", version.GitCommit)
	fmt.Printf("go:     %20s\n", version.BuildGoVersion)
	fmt.Printf("system: %20s\n", version.BuildSystem)

	app.Run()
}

// ── Handlers ─────────────────────────────────────────────────────

// Hello returns a simple success response.
// GET /hello
func Hello(c *frame.Context) {
	c.Success(map[string]string{"message": "hello, world"})
}

// GetUser demonstrates path params via Gtx.Param and error handling.
// GET /user/:id
func GetUser(c *frame.Context) {
	id := c.Gtx.Param("id")

	db := c.GetDB()
	if db != nil {
		var user User
		if err := db.First(&user, id).Error; err != nil {
			c.HTTPError2(http.StatusNotFound, "USER_NOT_FOUND", "user not found", err)
			return
		}
		c.Success(user)
		return
	}

	// Fallback when no database is configured
	c.Success(User{ID: 1, Name: "example_user"})
}

// CreateUser demonstrates request body binding via Gtx.ShouldBindJSON.
// POST /user  Body: {"name":"alice"}
func CreateUser(c *frame.Context) {
	var user User
	if err := c.Gtx.ShouldBindJSON(&user); err != nil {
		c.HTTPError2(http.StatusBadRequest, "INVALID_BODY", "invalid request body", err)
		return
	}

	db := c.GetDB()
	if db != nil {
		if err := db.Create(&user).Error; err != nil {
			c.HTTPError2(http.StatusInternalServerError, "CREATE_FAILED", "failed to create user", err)
			return
		}
		c.Success(user)
		return
	}

	c.Success(user)
}

// UpdateUser demonstrates PUT with body binding.
// PUT /user/:id  Body: {"name":"updated_name"}
func UpdateUser(c *frame.Context) {
	id := c.Gtx.Param("id")
	var user User
	if err := c.Gtx.ShouldBindJSON(&user); err != nil {
		c.HTTPError2(http.StatusBadRequest, "INVALID_BODY", "invalid request body", err)
		return
	}

	c.Infof("updating user %s: %+v", id, user)
	c.Success(map[string]interface{}{
		"id":      id,
		"updated": user,
	})
}

// DeleteUser demonstrates DELETE handler.
// DELETE /user/:id
func DeleteUser(c *frame.Context) {
	id := c.Gtx.Param("id")
	c.Infof("deleting user %s", id)
	c.Success(map[string]string{"deleted": id})
}

// PatchUser demonstrates PATCH handler.
// PATCH /user/:id  Body: {"name":"patched_name"}
func PatchUser(c *frame.Context) {
	id := c.Gtx.Param("id")
	var patch map[string]interface{}
	if err := c.Gtx.ShouldBindJSON(&patch); err != nil {
		c.HTTPError2(http.StatusBadRequest, "INVALID_BODY", "invalid request body", err)
		return
	}

	c.Infof("patching user %s: %+v", id, patch)
	c.Success(map[string]interface{}{
		"id":    id,
		"patch": patch,
	})
}

// HealthCheck returns OK when the service is healthy.
// HEAD /health
func HealthCheck(c *frame.Context) {
	c.Gtx.Status(http.StatusOK)
}

// ── Route group handlers ────────────────────────────────────────

// ListItems demonstrates paginated list response with HTTPListSuccess.
// GET /api/v1/items
func ListItems(c *frame.Context) {
	items := []Item{
		{ID: 1, Name: "item_a"},
		{ID: 2, Name: "item_b"},
	}

	c.HTTPListSuccess(&frame.PageResults{
		Total:    len(items),
		Page:     1,
		PageSize: 10,
		Results:  items,
	})
}

// CreateItem demonstrates external HTTP call via DoHTTP() and structured logging.
// POST /api/v1/items  Body: {"name":"new_item"}
func CreateItem(c *frame.Context) {
	var item Item
	if err := c.Gtx.ShouldBindJSON(&item); err != nil {
		c.HTTPError2(http.StatusBadRequest, "INVALID_BODY", "invalid request body", err)
		return
	}

	// Structured log with trace_id embedded
	c.WithField("item", item.Name).Infof("creating item")

	// Example: make an external HTTP request
	resp, err := c.DoHTTP().R().Get("https://httpbin.org/uuid")
	if err != nil {
		c.Errorf("external request failed: %v", err)
	} else {
		c.Infof("external response: %s", resp.String())
	}

	c.Success(item)
}

// Add is a simple utility function used by unit tests.
func Add(a, b int) int {
	return a + b
}
