package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"html/template"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"strings"
)

type server struct {
	Db *sqlx.DB
}

var (
    datass,inct,dect int
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
	Id       int `db:"id"`
	FullName string `db:"full_name"`
	Position string `db:"position"`
	IsStart  string `db:"is_start"`
}
type Statistic struct {
	//Id       int `db:"id"`
	//UserId   int `db:"userId"`
	//StartTime string `db:"startTime"`
	//EndTime  *string `db:"endTime"`
	FullName  string `db:"fullName"`
	Time  *string `db:"time"`
}

type UserStatus struct {
	Id int `db:"id"`
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

	fmt.Fprint(w, "{\"full_name\": \""+user.FullName+"\",\"position\": \""+user.Position+"\",\"is_start\": \""+user.IsStart+"\"}")
}
//BigTuna
func (s *server) testhandler(w http.ResponseWriter, r *http.Request){
	statLen := r.URL.Path[len("/test/"):]
	r.ParseForm()
	post := r.PostForm
	inc := strings.Join(post["inc"], "")
	dec := strings.Join(post["dec"], "")
	log.Println(inc)
	log.Println(dec)
	inct,_ =  strconv.Atoi(inc)
	dect,_ =  strconv.Atoi(dec)
	datass = datass + inct - dect
	log.Println(datass)
	log.Println(r.PostForm)
	var temp time.Time
	if statLen == ""  {
		temp = time.Now()
		temp = temp.AddDate(0,0,datass)
		log.Println("hey", datass)
		log.Println("hee", temp)
		fmt.Println(temp)
	} else {
		temp,_ = time.Parse("2006-01-02", statLen)
		//fmt.Println(temp)
		//fmt.Println(statLen)
		//fmt.Println(time.Parse("2006-01-02", statLen))
	}

	for temp.Weekday() != time.Monday {
		temp = temp.Add(-time.Hour * 24)
	}
	//fmt.Println(temp)
	stat := make([]Statistic, 0)
	//kek := "2017-07-14"
	if err := s.Db.Select(&stat, " SELECT (SELECT CONCAT(firstName, ' ', lastName) FROM users WHERE id = schedule.userId) as fullName , sum(TO_SECONDS(`endTime`) - TO_SECONDS(`startTime`)) / 3600 as time FROM schedule WHERE DATE(startTime)>=? GROUP BY userId;", temp.Format("2006-01-2")); err != nil {
		log.Println(err)
		return
	}
	testTemplate, err := template.ParseFiles("templates/test.html")
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println()
	if err := testTemplate.Execute(w, stat); err != nil {
		log.Println(err)
		return
	}
}
//BigTunaEnd
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
	http.HandleFunc("/test/", s.testhandler)

	log.Print("Server started at port " + strconv.Itoa(*servicePort))
	err = http.ListenAndServe(":"+strconv.Itoa(*servicePort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
