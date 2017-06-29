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
	Id       int `db:"id"`
	FullName string `db:"full_name"`
	Position string `db:"position"`
	IsStart  string `db:"is_start"`
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

func main() {
	flag.Parse()
	loadConfig(*configFile)
	log.Println("Config loaded from " + *configFile)

	s := server{
		Db: sqlx.MustConnect("mysql", config.MysqlLogin+":"+config.MysqlPassword+"@tcp("+config.MysqlHost+")/"+config.MysqlDb+"?charset=utf8"),
	}
	defer s.Db.Close()
	log.Printf("Connected to database on %s", config.MysqlHost)

	http.HandleFunc("/submit/", s.submitHandler)

	log.Print("Server started at port " + strconv.Itoa(*servicePort))
	err := http.ListenAndServe(":"+strconv.Itoa(*servicePort), nil)
	if err != nil {
		log.Fatal(err)
	}
}
