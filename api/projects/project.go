package projects

import (
	"database/sql"

	database "github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gin-gonic/gin"
	"github.com/masterminds/squirrel"
)

func ProjectMiddleware(c *gin.Context) {
	user := c.MustGet("user").(*models.User)

	projectID, err := util.GetIntParam("project_id", c)
	if err != nil {
		return
	}

	query, args, _ := squirrel.Select("p.*").
		From("project as p").
		Join("project__user as pu on pu.project_id=p.id").
		Where("p.id=?", projectID).
		Where("pu.user_id=?", user.ID).
		ToSql()

	var project models.Project
	if err := database.Mysql.SelectOne(&project, query, args...); err != nil {
		if err == sql.ErrNoRows {
			c.AbortWithStatus(404)
			return
		}

		panic(err)
	}

	c.Set("project", project)
	c.Next()
}

func GetProject(c *gin.Context) {
	c.JSON(200, c.MustGet("project"))
}

func MustBeAdmin(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	user := c.MustGet("user").(*models.User)

	userC, err := database.Mysql.SelectInt("select count(1) from project__user as pu join user as u on pu.user_id=u.id where pu.user_id=? and pu.project_id=? and pu.admin=1", user.ID, project.ID)
	if err != nil {
		panic(err)
	}

	if userC == 0 {
		c.AbortWithStatus(403)
		return
	}
}

func UpdateProject(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	var body struct {
		Name  string `json:"name"`
		Alert bool   `json:"alert"`
	}

	if err := c.Bind(&body); err != nil {
		return
	}

	if _, err := database.Mysql.Exec("update project set name=?, alert=? where id=?", body.Name, body.Alert, project.ID); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}

func DeleteProject(c *gin.Context) {
	project := c.MustGet("project").(models.Project)

	tx, err := database.Mysql.Begin()
	if err != nil {
		panic(err)
	}

	statements := []string{
		"delete tao from task__output as tao join task as t on t.id=tao.task_id join project__template as pt on pt.id=t.template_id where pt.project_id=?",
		"delete t from task as t join project__template as pt on pt.id=t.template_id where pt.project_id=?",
		"delete from project__template where project_id=?",
		"delete from project__user where project_id=?",
		"delete from project__repository where project_id=?",
		"delete from project__inventory where project_id=?",
		"delete from access_key where project_id=?",
		"delete from project where id=?",
	}

	for _, statement := range statements {
		_, err := tx.Exec(statement, project.ID)

		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	if err := tx.Commit(); err != nil {
		panic(err)
	}

	c.AbortWithStatus(204)
}
