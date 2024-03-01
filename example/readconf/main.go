package main

import (
	"fmt"

	"github.com/normastars/frame"
)

func main() {
	ymlPath := `E:\src\go\workspace\frame\example\readconf\conf\c.json`
	yml2Path := `E:\src\go\workspace\frame\example\conf\default.yaml`
	cm, err := frame.NewConfigManager(ymlPath)
	if err != nil {
		panic(err)
	}
	data := cm.Get("apiVersion")
	fmt.Println(data)
	rep := cm.Get("spec.selector")
	fmt.Println(rep)
	// name := cm.GetString("spec.template.spec.containers[0].name")
	// fmt.Println(name)
	fmt.Println("============================")
	cm2, err := frame.NewConfigManager(yml2Path)
	if err != nil {
		panic(err)
	}
	conf := &frame.Config{}
	if err := cm2.ReadConfigObject(conf); err != nil {
		panic(err)
	}
	fmt.Printf("%+v", conf)

}
