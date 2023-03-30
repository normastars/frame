package main

import (
	"fmt"
	"net/http"

	"github.com/nomainc/frame"
	"github.com/nomainc/frame/example/version"
)

type User struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

func main() {
	fmt.Printf("commit: %20s\n", version.GitCommit)
	fmt.Printf("built on %20s\n", version.BuildGoVersion)
	fmt.Printf("built on %20s\n", version.BuildSystem)
	router := frame.Default()
	router.GET("/hello", HelloWorld)

	router.Run(":8080")

}

// HelloWorld hell world handler
func HelloWorld(c *frame.Context) {
	db := c.GetDB("user")
	// table auto migrate
	// err := db.AutoMigrate(&User{})
	// if err != nil {
	// 	c.Fatalf("failed to migrate table: %v", err)
	// }
	// 创建用户
	user := User{Name: "test_user"}
	result := db.Create(&user)
	if result.Error != nil {
		c.Fatalf("failed to create user: %v", result.Error)
	}
	c.Infof("created user: %v\n", user)

	// 查询用户
	var foundUser User
	result = db.First(&foundUser, user.ID)
	if result.Error != nil {
		c.Fatalf("failed to find user: %v", result.Error)
	}
	c.Infof("found user: %v\n", foundUser)

	// 更新用户
	result = db.Model(&foundUser).Update("name", "updated_user")
	if result.Error != nil {
		c.Fatalf("failed to update user: %v", result.Error)
	}
	c.Infof("updated user: %v\n", foundUser)

	// 删除用户
	result = db.Delete(&foundUser)
	if result.Error != nil {
		c.Fatalf("failed to delete user: %v", result.Error)
	}
	c.HTTPError2(http.StatusOK, "X0111", "普通错误", fmt.Errorf("系统致命错误"))
}
