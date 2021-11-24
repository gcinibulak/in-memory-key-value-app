package main

import (
	"github.com/gin-gonic/gin"

	"encoding/json"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"inMemoryKeyValueApp/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	ginSwagger "github.com/swaggo/gin-swagger"
)

const fileName string = "tmp/TIMESTAMP-data.json"
const interval float64 = 100000

var kVItems []models.KVItem

func checkFile(filename string) error {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		_, err := os.Create(filename)
		if err != nil {
			return err
		}
	}
	return nil
}

func readFile() {
	err := checkFile(fileName)
	if err != nil {
		log.Println("Error when reading file ", err)
	}

	file, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Error when reading file ", err)
	}

	json.Unmarshal(file, &kVItems)
}

func writeFile() {
	err := checkFile(fileName)
	if err != nil {
		log.Println("Error when reading file ", err)
	}

	dataBytes, err := json.Marshal(kVItems)
	if err != nil {
		log.Println("Error when marshal ", err)
	}

	err = ioutil.WriteFile(fileName, dataBytes, 0644)
	if err != nil {
		log.Println("Error when writing ", err)
	}
}

func flushData(ctx *gin.Context) {
	err := checkFile(fileName)
	if err != nil {
		log.Println("Error when reading file ", err)
	}

	err = os.Truncate(fileName, 0)
	if err != nil {
		log.Println("Error when flush data ", err)
	}

	ctx.IndentedJSON(http.StatusOK, "Data is successfully flushed")
}

func checkItemIsExists(key string) bool {
	for _, item := range kVItems {
		if item.Key == key {
			return true
		}
	}
	return false
}

func allKVItems(ctx *gin.Context) {
	if len(kVItems) > 0 {
		ctx.IndentedJSON(http.StatusOK, kVItems)
	} else {
		ctx.IndentedJSON(http.StatusNotFound, "Data not found")
	}
}

func addItem(ctx *gin.Context) {
	var item models.KVItem
	if err := ctx.BindJSON(&item); err != nil {
		ctx.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	result := checkItemIsExists(item.Key)
	if result {
		ctx.IndentedJSON(http.StatusInternalServerError, "Key is exists. Please try different key value")
		return
	}
	kVItems = append(kVItems, item)
	ctx.IndentedJSON(http.StatusOK, kVItems)
}

func findByKey(ctx *gin.Context) {
	key := ctx.Param("key")
	for _, item := range kVItems {
		if item.Key == key {
			ctx.IndentedJSON(http.StatusOK, item)
		}
	}
}

func autoSaveData() {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	log.Println("Automatically file saved started")

	go func() {
		for {
			select {
			case t := <-ticker.C:
				writeFile()
				log.Println("Automatically file saved processed ", t)
			}
		}
	}()
	log.Println("Automatically file saved completed")
}

func handleRequest() {
	router := gin.Default()

	router.Use(gin.Logger())

	// swagger:route GET /items  get all key-value
	router.GET("/items", func(ctx *gin.Context) {
		allKVItems(ctx)
	})

	// swagger:route GET /items  get value by key
	router.GET("/items/{key}", func(ctx *gin.Context) {
		findByKey(ctx)
	})

	// swagger:route POST /items add new key-value
	router.POST("/items", func(ctx *gin.Context) {
		addItem(ctx)
	})

	// flushData ... Flush all existing data
	// swagger:route GET /flush flush all existing data
	router.GET("/flush", func(ctx *gin.Context) {
		flushData(ctx)
	})

	// swagger:route GET /swagger/*any api documentation
	router.GET("/swagger/*any", func(ctx *gin.Context) {
		ginSwagger.WrapHandler(swaggerFiles.Handler)
	})

	log.Fatal(router.Run("localhost:8080"))
}

//@title In Memory Key-Value Application Documentation
//@version 1.0.0

//@host localhost:8080
//@BasePath /items

func main() {
	readFile()
	autoSaveData()
	handleRequest()
}
