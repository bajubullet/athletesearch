package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	DB_HOST    = "localhost"
	DB_NAME    = "athletesearch"
	COLLECTION = "athlete"
	SHORT_FORM = "January 02, 2006"
)

type Athlete struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	Name          string
	Birthday      time.Time
	Exp           int
	Skills        []string
	Championships []string
}

func (a *Athlete) Save(session *mgo.Session) error {
	return session.DB(DB_NAME).C(COLLECTION).Insert(a)
}

func CreateAthlete(name, birthday, skills, championships string, exp int) (Athlete, error) {
	athlete := Athlete{}
	bday, err := time.Parse(SHORT_FORM, birthday)
	skillsArr := strings.Split(skills, ", ")
	championshipsArr := strings.Split(championships, ", ")
	if err == nil {
		athlete.Name = name
		athlete.Birthday = bday
		athlete.Exp = exp
		athlete.Skills = skillsArr
		athlete.Championships = championshipsArr
	} else {
		log.Panic(err)
	}
	return athlete, err
}

func ReadCSV(file string, session *mgo.Session) {
	f, err := os.Open(file)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Panic(err)
	}
	for _, line := range lines {
		name := line[0]
		birthday := line[1]
		exp, err := strconv.Atoi(line[2])
		if err != nil {
			log.Panic(err)
		}
		skills := line[3]
		championships := line[4]
		fmt.Println(name)
		athelete, err := CreateAthlete(name, birthday, skills, championships, exp)
		if err != nil {
			log.Panic(err)
		}
		athelete.Save(session)
	}
}

func PrintResults(results []Athlete) {
	for _, athlete := range results {
		fmt.Println(athlete.Name)
	}
}

func FindByName(q string, c *mgo.Collection) []Athlete {
	// TODO: Add createIndex(text) and use $text instead of $regex
	result := []Athlete{}
	c.Find(bson.M{
		"name": bson.M{"$regex": ".*" + q + ".*"},
	}).All(&result)
	return result
}

func FindBySkill(skill string, c *mgo.Collection) []Athlete {
	result := []Athlete{}
	c.Find(bson.M{
		"skills": skill,
	}).All(&result)
	return result
}

func FindByAge(min, max int, c *mgo.Collection) []Athlete {
	result := []Athlete{}
	now := time.Now()
	dobGreaterThan := now.AddDate(-1*max, 0, 0)
	dobSmallerThan := now.AddDate(-1*min, 0, 0)
	c.Find(bson.M{
		"birthday": bson.M{
			"$gt": dobGreaterThan,
			"$lt": dobSmallerThan,
		},
	}).All(&result)
	return result
}

func main() {
	session, err := mgo.Dial(DB_HOST)
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	c := session.DB("athletesearch").C("athlete")
	// PrintResults(FindByName("ss", c))
	// PrintResults(FindBySkill("snowboarding", c))
	PrintResults(FindByAge(20, 40, c))
}
