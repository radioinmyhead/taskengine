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
		if dberr := order.Upsert(); dberr != nil {
			fmt.Println(dberr)
			c.String(400, "err=%s", err.Error())
			return
		}

		err = job.NewDbjob("machine_create", []string{"machine_create_callapi",
			"machine_create_cloudinit",
			"machine_create_check"}, string(plugin.DBMachineCreate), order.ID).Start()
		if err != nil {
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "order=%v", order)
	})

	router.GET("/machine/init", func(c *gin.Context) {
		ip := "1.2.3.4"
		ci := plugin.NewCloudinit(ip)
		err := ci.Upsert()
		if err != nil {
			fmt.Println(err)
			c.String(400, "err=%s", err.Error())
			return
		}
		err = job.NewDbjob("cloudinit", []string{"machine_init_create", "machine_init_installpackage", "machine_init_reboot"},
			string(plugin.DBMachineInit), ci.ID).Start()
		if err != nil {
			fmt.Println(err)
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "order=%s", ci.ID)
	})

	router.GET("/continue", func(c *gin.Context) {
		err := job.ContinueJobs()
		if err != nil {
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "list=%s", "succ")
	})

	router.GET("/retry", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			c.String(400, "need id")
			return
		}
		err := job.RetryJob(id)
		if err != nil {
			err = fmt.Errorf("http retry job failed id=%v err=%v", id, err.Error())
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "succ=", id)
		return
	})

	router.Run(":8080")
}
