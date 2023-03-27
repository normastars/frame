package main

import (
	"fmt"

	"github.com/nomainc/frame"
	"github.com/nomainc/frame/version"
	"gopkg.in/yaml.v3"
)

func main() {
	fmt.Printf("commit: %20s\n", version.GitCommit)
	fmt.Printf("built on %20s\n", version.BuildGoVersion)
	fmt.Printf("built on %20s\n", version.BuildSystem)
	// router := frame.Default()
	// router.GET("/hello", func(c *frame.Context) {
	// 	logrus.Info("哈哈")
	// 	logrus.Infof("hello world %s", "v")
	// 	c.Success(nil)
	// })

	// router.Run(":8080")
	config := frame.GetConfig()
	b, _ := yaml.Marshal(config)
	fmt.Println(string(b))
}
