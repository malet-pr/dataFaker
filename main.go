package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
	"math/rand"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
	"github.com/gin-contrib/cors"

)

type Shift struct {
	Date      	time.Time `faker:"-"`
	Capacity  	string    `faker:"oneof:AB, BB, CB, AD"`
	Skill     	[3]Skill  `faker:"unique"`
	RouteArea 	string    `faker:"oneof:123der, 873mlh, 545kje, 990res"`
}

type Skill struct {
	Code 		string 	`faker:"oneof:bec,EIF,LDr,Ndr,Adg,AFb,yJg,DBa,MxG"`
}

type Superior struct {
	SuperiorID  int    	`faker:"boundary_start=10,boundary_end=100,unique"`
	Name 		string 	`faker:"name,unique"`
	Phone       string 	`faker:"phone_number,unique"`
}

type Technic struct {
	TechnicID	int    	`faker:"boundary_start=100,boundary_end=1000,unique"`
	Name  		string 	`faker:"name,unique"`
	SuperiorID 	int     `faker:"-"`
	Shift 		Shift
}

type DataSaver interface {
    ID() int
	Serialize() ([]byte, error)
}

var superiors []Superior
var technics []Technic
var superiorsToSave []DataSaver
var technicsToSave []DataSaver

func putDate() time.Time {
	startDate := time.Now()
    endDate := startDate.Add(time.Hour * 24 * 7)
    duration := endDate.Sub(startDate)
    randomDuration := time.Duration(rand.Int63n(int64(duration)))
    return startDate.Add(randomDuration)
}

func (t *Technic) ID() int {
    return t.TechnicID
}

func (t *Technic) Serialize() ([]byte, error) {
    return json.Marshal(t)
}

func (t *Technic) Deserialize(data []byte) error {
    return json.Unmarshal(data, t)
}

func (s *Superior) ID() int {
    return s.SuperiorID
}

func (s *Superior) Serialize() ([]byte, error) {
    return json.Marshal(s)
}

func (s *Superior) Deserialize(data []byte) error {
    return json.Unmarshal(data, s)
}

func SaveData(db *bolt.DB, dataSlice []DataSaver, bucketName string) error {
    return db.Update(func(tx *bolt.Tx) error {
		bucket, createErr := tx.CreateBucketIfNotExists([]byte(bucketName))
		if createErr != nil {
			return createErr
		}
		for _, data := range dataSlice {
            serializedData, err := data.Serialize()
            if err != nil {
                return err
            }
            if err := bucket.Put([]byte(fmt.Sprintf("%d", data.ID())), serializedData); err != nil {
                return err
            }
        }
		log.Println("Data saved successfully")
        return nil
    })
}

func GetTechnics(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, technics)
}

func GetSuperiors(c *gin.Context) {
    c.IndentedJSON(http.StatusOK, superiors)
}

func main() {
    for i := 1; i <= 15; i++ {
		superior := Superior{}
		err := faker.FakeData(&superior)
		superiors = append(superiors, superior)
		superiorsToSave = append(superiorsToSave,&superior)
        if err != nil {
            fmt.Println(err)
        }
    }
	for i :=1; i <= 100; i++ {
		technic := Technic{}
		err := faker.FakeData(&technic)
		randomIndex := rand.Intn(len(superiors))
		technic.SuperiorID = superiors[randomIndex].SuperiorID
		technic.Shift.Date = putDate();
		technics = append(technics, technic)
		technicsToSave = append(technicsToSave,&technic)
        if err != nil {
            fmt.Println(err)
        }
	}
	faker.ResetUnique()
	db, err := bolt.Open("mydb.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = SaveData(db,superiorsToSave,"superiors")
	if err != nil {
		log.Fatal(err)
	}
	err = SaveData(db,technicsToSave,"technics")
	if err != nil {
		log.Fatal(err)
	}
	router := gin.Default()
	config := cors.DefaultConfig()
    config.AllowOrigins = []string{"*"}  
    router.Use(cors.New(config))
	router.GET("/technics", GetTechnics)
	router.GET("/superiors", GetSuperiors)
	httpServer := &http.Server{
		Addr:    ":9090",
		Handler: router,
		ReadTimeout: 0,
		WriteTimeout: 0,
	}
	httpServer.ListenAndServe()
}

