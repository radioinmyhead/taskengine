package main

import (
	"fmt"
	"haha/db"
	"haha/job"
	"haha/plugin"

	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	db.Init()

	router := gin.Default()

	router.POST("/machine/create", func(c *gin.Context) {
		var err error
		/*
			var order *model.CreateMachine
			err = c.Bind(order)
			if err != nil {
				fmt.Println(err)
				c.String(400, "err=%s", err.Error())
				return
			}
			order.Init()
		*/

		order, _ := plugin.NewMachineCreate("admin", "planA", 3)
		err = order.Insert()
		if err != nil {
			fmt.Println(err)
			c.String(400, "err=%s", err.Error())
			return
		}

		err = job.NewDbjob("machine_create", []string{"machine_create_callapi",
			"machine_create_cloudinit",
			"machine_create_check"}, string(plugin.DBMachineCreate), order.ID).Start()
		if err != nil {
			fmt.Println(err)
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "order=%s", order)
	})

	router.GET("/all", func(c *gin.Context) {
		err := job.ContinueJobs()
		if err != nil {
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "list=%s", "succ")
	})

	router.Run(":8080")
}
