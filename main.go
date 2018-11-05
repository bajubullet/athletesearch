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
	DB_HOST             = "localhost"
	DB_NAME             = "athletesearch"
	COLLECTION          = "athlete"
	CATEGORY_COLLECTION = "category"
	SHORT_FORM          = "January 02, 2006"
)

type Athlete struct {
	ID            bson.ObjectId `bson:"_id,omitempty"`
	Name          string
	Birthday      time.Time
	Exp           int
	Skills        []string
	Championships []string
}

type SportsCategory struct {
	ID     bson.ObjectId `bson:"_id,omitempty"`
	Name   string
	Sports []string
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
	if len(results) == 0 {
		fmt.Println("No results")
		return
	}
	for _, athlete := range results {
		fmt.Println(athlete.Name)
	}
}

func FindByName(q string, c *mgo.Collection) []Athlete {
	// TODO: Add createIndex(text) and use $text instead of $regex also make this case insensitive
	result := []Athlete{}
	c.Find(bson.M{
		"name": bson.M{"$regex": ".*" + q + ".*"},
	}).All(&result)
	return result
}

func FindBySkill(skill string, session *mgo.Session) []Athlete {
	result := []Athlete{}
	category := SportsCategory{}
	err := session.DB(DB_NAME).C(CATEGORY_COLLECTION).Find(bson.M{
		"name": skill,
	}).One(&category)
	if err != nil {
		session.DB(DB_NAME).C(COLLECTION).Find(bson.M{
			"skills": skill,
		}).All(&result)
	} else {
		session.DB(DB_NAME).C(COLLECTION).Find(bson.M{
			"skills": bson.M{
				"$in": category.Sports,
			},
		}).All(&result)
	}
	return result
}

func FindByChampionship(championship string, c *mgo.Collection) []Athlete {
	result := []Athlete{}
	c.Find(bson.M{})
	c.Find(bson.M{
		"championships": championship,
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
	if len(os.Args) < 3 {
		log.Fatal("Please pass an option(name/skill/age/championship/loadCSV) and a value")
	}
	session, err := mgo.Dial(DB_HOST)
	if err != nil {
		log.Panic(err)
	}
	defer session.Close()

	c := session.DB("athletesearch").C("athlete")
	switch os.Args[1] {
	case "name":
		PrintResults(FindByName(strings.Join(os.Args[2:], " "), c))
	case "skill":
		PrintResults(FindBySkill(strings.Join(os.Args[2:], " "), session))
	case "championship":
		PrintResults(FindByChampionship(strings.Join(os.Args[2:], " "), c))
	case "loadCSV":
		ReadCSV(os.Args[2], session)
	case "age":
		if len(os.Args) < 4 {
			log.Fatal("Please pass 2 params")
		}
		min, _ := strconv.Atoi(os.Args[2])
		max, _ := strconv.Atoi(os.Args[3])
		PrintResults(FindByAge(min, max, c))
	default:
		log.Fatal("Invalid params")
	}
}
