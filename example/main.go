package main

import (
	"fmt"
	"net/http"

	"github.com/nomastars/frame"
	"github.com/nomastars/frame/example/version"
)

type User struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func main() {
	fmt.Printf("commit: %20s\n", version.GitCommit)
	fmt.Printf("built on %20s\n", version.BuildGoVersion)
	fmt.Printf("built on %20s\n", version.BuildSystem)
	app := frame.New()
	app.GET("/hello", HelloWorld)

	app.Run()

}

// HelloWorld hell world handler
func HelloWorld(c *frame.Context) {
	db := c.GetDB()
	fmt.Println("db", db.Config)
	// create user
	user := User{Name: "test_user"}
	result := db.Create(&user)
	if result.Error != nil {
		c.Fatalf("failed to create user: %v", result.Error)
	}
	c.Infof("created user: %v\n", user)
	c.HTTPError2(http.StatusOK, "X0111", "normal error", fmt.Errorf("system panic"))
}

func Add(a, b int) int {
	return a + b
}
