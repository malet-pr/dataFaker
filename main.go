package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"time"

	"github.com/boltdb/bolt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
)

type Shift struct {
	Date      time.Time `faker:"-"`
	Capacity  string    `faker:"oneof:AB, BB, CB, AD"`
	Skill     [3]Skill  `faker:"unique"`
	RouteArea string    `faker:"oneof:123der, 873mlh, 545kje, 990res"`
}

type Skill struct {
	Code string `faker:"oneof:bec,EIF,LDr,Ndr,Adg,AFb,yJg,DBa,MxG"`
}

type Superior struct {
	SuperiorID int    `faker:"boundary_start=10,boundary_end=100,unique"`
	Name       string `faker:"name,unique"`
	Phone      string `faker:"phone_number,unique"`
}

type Technic struct {
	TechnicID  int    `faker:"boundary_start=100,boundary_end=1000,unique"`
	Name       string `faker:"name,unique"`
	SuperiorID int    `faker:"-"`
	Shift      Shift
}

type DataSaver interface {
	ID() int
	Serialize() ([]byte, error)
}

type DataRetriever interface {
	ID() int
}

var db *bolt.DB

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
		return nil
	})
}

func GetAllData(db *bolt.DB, bucketName string, dataRetriever DataRetriever) ([]DataRetriever, error) {
	var data []DataRetriever
	objType := reflect.TypeOf(dataRetriever).Elem()
	obj := reflect.New(objType).Interface()
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("Bucket not found: %s", bucketName)
		}
		cursor := bucket.Cursor()
		for key, value := cursor.First(); key != nil; key, value = cursor.Next() {
			if err := json.Unmarshal(value, obj); err != nil {
				return err
			}
			data = append(data, obj.(DataRetriever))
		}
		return nil
	})
	return data, err
}

func GetTechnics(c *gin.Context) {
	allTechnics, err := GetAllData(db, "technics", &Technic{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
	c.IndentedJSON(http.StatusOK, allTechnics)
}

func GetSuperiors(c *gin.Context) {
	allSuperiors, err := GetAllData(db, "superiors", &Superior{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
	}
	c.IndentedJSON(http.StatusOK, allSuperiors)
}

func main() {
	var err error
	db, err = bolt.Open("fake.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	for i := 1; i <= 15; i++ {
		superior := Superior{}
		err := faker.FakeData(&superior)
		superiors = append(superiors, superior)
		superiorsToSave = append(superiorsToSave, &superior)
		if err != nil {
			fmt.Println(err)
		}
	}
	for i := 1; i <= 100; i++ {
		technic := Technic{}
		err := faker.FakeData(&technic)
		randomIndex := rand.Intn(len(superiors))
		technic.SuperiorID = superiors[randomIndex].SuperiorID
		technic.Shift.Date = putDate()
		technics = append(technics, technic)
		technicsToSave = append(technicsToSave, &technic)
		if err != nil {
			fmt.Println(err)
		}
	}
	faker.ResetUnique()
	err = SaveData(db, superiorsToSave, "superiors")
	if err != nil {
		log.Fatal(err)
	}
	err = SaveData(db, technicsToSave, "technics")
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
		Addr:         ":9090",
		Handler:      router,
		ReadTimeout:  0,
		WriteTimeout: 0,
	}
	httpServer.ListenAndServe()
}
