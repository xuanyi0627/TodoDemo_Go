package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"database/sql"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"time"
)

type Thing struct {
	Uid     	string    `json:"uid"`
	Title   	string 	  `json:"title"`
	Content 	string    `json:"content"`
	Finish  	bool      `json:"finish"`
	CreateTime 	int       `json:"createTime"`
}

func newUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

func openDB() (*sql.DB , error) {
	db, err := sql.Open("mysql", "root:123123@tcp(localhost:3306)/db_todos?charset=utf8")
	return db, err
}

func fetchTodos(c *gin.Context) {

	db, err := openDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库打开失败",
		})
	}

	rows, err := db.Query("SELECT uid, title, content, isdone, createTime FROM things ORDER BY id ASC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库查询失败",
		})
	}

	things := make([]Thing, 0)

	for rows.Next() {
		var uid string
		var title string
		var content string
		var finish bool
		var createTime int
		rows.Scan(&uid, &title, &content, &finish, &createTime)
		thing := Thing{Uid:uid, Title:title, Content:content, Finish:finish, CreateTime:createTime}
		things = append(things, thing)
	}

	for _, t := range things {
		fmt.Println(t)
	}

	if err = rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库读取失败",
		})
	}

	content := gin.H{"things": things,}
	c.JSON(http.StatusOK, content)
}

func addTodo(c *gin.Context) {

	db, err := openDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库打开失败",
		})
	}

	uid := newUUID()
	title := c.Request.FormValue("title")
	content := c.Request.FormValue("content")
	createTime := time.Now().Unix()

	rs, err := db.Exec("INSERT INTO things (uid, title, content, createTime, isdone) VALUE (?,?,?,?, false)", uid, title, content, createTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据写入失败",
		})
	}

	_, err = rs.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据写入失败",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"uid": uid,
			"title": title,
		})
	}
}

func fetchTodo(c *gin.Context) {

	db, err := openDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库打开失败",
		})
	}

	uid := c.Param("uid")
	rows := db.QueryRow("SELECT * FROM things WHERE uid = ?", uid)

	var title string
	var content string
	var finish bool
	var createTime int
	rows.Scan(&uid, &title, &content, &finish, &createTime)
	thing := Thing{Uid:uid, Title:title, Content:content, Finish:finish, CreateTime:createTime}

	c.JSON(http.StatusOK, thing)
}

func setTodoIsFinish(c *gin.Context)  {
	db, err := openDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库打开失败",
		})
		return
	}

	uid := c.Param("uid")
	stmt, err := db.Prepare("UPDATE things SET isdone = ? WHERE uid = ?")
	defer stmt.Close()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库修改失败",
		})
		return
	}

	rs, err := stmt.Exec(true, uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库修改失败",
		})
		return
	}

	ra, err := rs.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库修改失败",
		})
		return
	}

	msg := fmt.Sprintf("Update things %s successful %d", uid, ra)

	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})

}

func deleteTodo(c *gin.Context)  {
	db, err := openDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库打开失败",
		})
		return
	}

	uid := c.Param("uid")

	rs, err := db.Exec("DELETE FROM things WHERE uid = ?", uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库删除失败",
		})
		return
	}

	ra, err := rs.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "数据库删除失败",
		})
		return
	}

	msg := fmt.Sprintf("delete things %s successful %d", uid, ra)

	c.JSON(http.StatusOK, gin.H{
		"msg": msg,
	})

}

func main()  {

	gin.SetMode(gin.DebugMode)

	route := gin.Default()

	route.GET("/fetchTodos", fetchTodos)
	route.POST("/addTodo", addTodo)
	route.GET("/fetchTodo/:uid", fetchTodo)
	route.PUT("/setTodoIsFinish/:uid", setTodoIsFinish)
	route.DELETE("/deleteTodo/:uid", deleteTodo)

	route.Run(":9090")
}

