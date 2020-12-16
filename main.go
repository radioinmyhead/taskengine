package main

import (
	"fmt"
	"haha/job"
	"net/http"

	"github.com/gin-gonic/gin"
)

var dic = map[string][]string{
	"create": {"callapi", "machinewait", "cloudinit"},
}

func main() {
	router := gin.Default()

	router.GET("/machine/:name", func(c *gin.Context) {
		jobname := c.Param("name")
		list := dic[jobname]
		err := job.NewDbjob(jobname, list).Start()
		if err != nil {
			fmt.Println(err)
			c.String(400, "err=%s", err.Error())
			return
		}
		c.String(http.StatusOK, "jobname=%s", jobname)
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
