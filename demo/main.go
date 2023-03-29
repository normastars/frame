package main

import (
	"fmt"

	"github.com/nomainc/frame"
	"github.com/nomainc/frame/version"
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
	router.GET("/hello", func(ctx *frame.Context) {
		db := ctx.GetDB("user")

		// 自动迁移表结构
		err := db.AutoMigrate(&User{})
		if err != nil {
			ctx.Fatalf("failed to migrate table: %v", err)
		}
		// 创建用户
		user := User{Name: "test_user"}
		result := db.Create(&user)
		if result.Error != nil {
			ctx.Fatalf("failed to create user: %v", result.Error)
		}
		fmt.Printf("created user: %v\n", user)

		// 查询用户
		var foundUser User
		result = db.First(&foundUser, user.ID)
		if result.Error != nil {
			ctx.Fatalf("failed to find user: %v", result.Error)
		}
		fmt.Printf("found user: %v\n", foundUser)

		// 更新用户
		result = db.Model(&foundUser).Update("name", "updated_user")
		if result.Error != nil {
			ctx.Fatalf("failed to update user: %v", result.Error)
		}
		fmt.Printf("updated user: %v\n", foundUser)

		// 删除用户
		result = db.Delete(&foundUser)
		if result.Error != nil {
			ctx.Fatalf("failed to delete user: %v", result.Error)
		}
		ctx.Infoln("哈哈哈")
		ctx.Success(nil)
	})

	router.Run(":8080")

}
