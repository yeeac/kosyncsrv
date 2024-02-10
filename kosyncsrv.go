package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

type requestUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type requestHeader struct {
	AuthUser string `header:"x-auth-user"`
	AuthKey  string `header:"x-auth-key"`
}

type requestPosition struct {
	DocumentID string  `json:"document"`
	Percentage float64 `json:"percentage"`
	Progress   string  `json:"progress"`
	Device     string  `json:"device"`
	DeviceID   string  `json:"device_id"`
}

type replayPosition struct {
	DocumentID string
	Timestamp  int64
}

type requestDocid struct {
	DocumentID string `uri:"document" bingding:"required"`
}

func register(c *gin.Context) {
	var rUser requestUser
	if err := c.ShouldBindJSON(&rUser); err != nil {
		c.String(http.StatusBadRequest, "Bind Json error")
		return
	}

	if rUser.Username == "" || rUser.Password == "" {
		c.String(http.StatusBadRequest, "Invalid request")
		return
	} else if !addDBUser(rUser.Username, rUser.Password) {
		c.String(http.StatusConflict, "Username is already registered")
		return
	} else {
		c.JSON(http.StatusCreated, gin.H{
			"username": rUser.Username,
		})
		return
	}
}

func authorizeRequest(c *gin.Context) (ruser string) {
	var rHeader requestHeader
	if err := c.ShouldBindHeader(&rHeader); err != nil {
		c.String(http.StatusBadRequest, "Bind Header error")
	}
	if rHeader.AuthUser == "" || rHeader.AuthKey == "" {
		c.String(http.StatusUnauthorized, "Wrong header or Blank Value")
		return ruser
	}
	dUser, norows := getDBUser(rHeader.AuthUser)
	if norows {
		c.String(http.StatusForbidden, "Forbidden")
		return ruser
	} else if rHeader.AuthKey != dUser.Password {
		c.String(http.StatusUnauthorized, "Unauthorized")
		return ruser
	} else {
		return rHeader.AuthUser
	}
}

func authorize(c *gin.Context) {
	_ = authorizeRequest(c)
	c.JSON(200, gin.H{
		"authorized": "OK",
	})
}

func getProgress(c *gin.Context) {
	ruser := authorizeRequest(c)
	var rDocid requestDocid
	if err := c.ShouldBindUri(&rDocid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err})
		return
	}
	position := getDBPosition(ruser, rDocid.DocumentID)
	c.JSON(http.StatusOK, position)
}

func updateProgress(c *gin.Context) {
	ruser := authorizeRequest(c)
	var rPosition requestPosition
	var reply replayPosition

	if err := c.ShouldBindJSON(&rPosition); err != nil {
		c.String(http.StatusBadRequest, "Bind Json error")
		return
	}
	updatetime := updateDBdocument(ruser, rPosition)
	reply.DocumentID = rPosition.DocumentID
	reply.Timestamp = updatetime
	c.JSON(http.StatusOK, reply)
}

func main() {
	dbfile := flag.String("d", "syncdata.db", "Sqlite3 DB file name")
	srvhost := flag.String("t", "127.0.0.1", "Server host")
	srvport := flag.Int("p", 8080, "Server port")
	sslswitch := flag.Bool("ssl", false, "Start with https")
	sslc := flag.String("c", "", "SSL Certificate file")
	sslk := flag.String("k", "", "SSL Private key file")
	
	flag.Usage = func() {
		fmt.Println(`Usage: kosyncsrv [-h] [-t 127.0.0.1] [-p 8080] [-ssl -c "./cert.pem" -k "./cert.key"]`)
		flag.PrintDefaults()
	}
	flag.Parse()
	
	bindsrv := *srvhost + ":" + fmt.Sprint(*srvport)
	
	dbname = *dbfile
	initDB()

	router := gin.Default()
	router.POST("/users/create", register)
	router.GET("/users/auth", authorize)
	router.GET("/syncs/progress/:document", getProgress)
	router.PUT("/syncs/progress", updateProgress)
	if *sslswitch {
		router.RunTLS(bindsrv, *sslc, *sslk)
	} else {
		router.Run(bindsrv)
	}
}
