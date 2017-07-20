package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type server struct {
	Db *sqlx.DB
}

var (
	configFile  = flag.String("Config", "conf.json", "Where to read the Config from")
	servicePort = flag.Int("Port", 4005, "Application port")
)

var config struct {
	MysqlLogin    string `json:"mysqlLogin"`
	MysqlPassword string `json:"mysqlPassword"`
	MysqlHost     string `json:"mysqlHost"`
	MysqlDb       string `json:"mysqlDb"`
}

type UserInfo struct {
	Id       int `db:"id" json:"id"`
	FullName string `db:"full_name" json:"full_name"`
	Position string `db:"position" json:"position"`
	IsStart  string `db:"is_start" json:"is_start"`
}
type Statistics struct {
	FullName string `db:"fullName"`
	Time     *string `db:"time"`
}

type UserStatus struct {
	Id int `db:"id"`
}

type StatisticsPage struct {
	Statistics []Statistics
	NextDate   string
	PrevDate   string
}

type SubmitData struct {
}

func loadConfig(path string) error {
	jsonData, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, &config)
}

func (s *server) insertTime(userId int) {
	if _, err := s.Db.Exec("INSERT INTO schedule (userId) VALUES (?)", userId); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) updateTime(userId int) {
	if _, err := s.Db.Exec("UPDATE schedule SET endTime = CURRENT_TIMESTAMP WHERE userId = ? AND endTime IS NULL", userId); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) submitHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	tagId := r.URL.Path[len("/submit/"):]
	user := UserInfo{}
	if err := s.Db.Get(&user, "SELECT id, CONCAT(firstName, ' ', lastName) as full_name, position FROM users WHERE `tagId` = ?", tagId); err != nil {
		log.Println(err)
		return
	}

	userTime := UserStatus{}
	if err := s.Db.Get(&userTime, "SELECT id FROM schedule WHERE `userId` = ? AND `endTime` IS NULL", user.Id); err != nil {
		if err == sql.ErrNoRows {
			s.insertTime(user.Id)
			user.IsStart = "true"
		} else {
			log.Println(err)
			return
		}
	} else {
		s.updateTime(user.Id)
		user.IsStart = "false"
	}

	if result, err := json.Marshal(user); err != nil {
		log.Println(err)
		return
	} else {
		fmt.Fprintf(w, string(result))
	}
}

func (s *server) statisticsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	date := r.URL.Path[len("/statistics/"):]
	var err error
	var pageDate time.Time
	if date == "" {
		pageDate = time.Now()
	} else {
		pageDate, err = time.Parse("2006-01-02", date)
		if err != nil {
			log.Println(err)
			return
		}
	}

	for pageDate.Weekday() != time.Monday {
		pageDate = pageDate.Add(-time.Hour * 24)
	}
	stat := make([]Statistics, 0)
	if err := s.Db.Select(&stat, " SELECT (SELECT CONCAT(firstName, ' ', lastName) FROM users WHERE id = schedule.userId) as fullName , sum(TO_SECONDS(`endTime`) - TO_SECONDS(`startTime`)) / 3600 as time FROM schedule WHERE DATE(startTime) >= ? AND DATE(startTime) <= ? GROUP BY userId;", pageDate.Format("2006-01-02"), pageDate.Add(time.Hour * 24 * 6).Format("2006-01-02")); err != nil {
		log.Println(err)
		return
	}
	testTemplate, err := template.ParseFiles("templates/statistics.html")
	if err != nil {
		log.Println(err)
		return
	}

	if err := testTemplate.Execute(w, StatisticsPage{
		Statistics: stat,
		PrevDate:   pageDate.Add(-time.Hour * 24 * 7).Format("2006-01-02"),
		NextDate:   pageDate.Add(time.Hour * 24 * 7).Format("2006-01-02"),
	}); err != nil {
		log.Println(err)
		return
	}
}

func main() {
	flag.Parse()
	err := loadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Config loaded from " + *configFile)

	s := server{
		Db: sqlx.MustConnect("mysql", config.MysqlLogin+":"+config.MysqlPassword+"@tcp("+config.MysqlHost+")/"+config.MysqlDb+"?charset=utf8"),
	}
	defer s.Db.Close()
	log.Printf("Connected to database on %s", config.MysqlHost)

	http.HandleFunc("/submit/", s.submitHandler)
	http.HandleFunc("/statistics/", s.statisticsHandler)

	log.Print("Server started at port " + strconv.Itoa(*servicePort))
	err = http.ListenAndServe(":"+strconv.Itoa(*servicePort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
