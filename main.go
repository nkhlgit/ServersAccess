package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	_ "github.com/mattn/go-sqlite3"
)

//sever struct contains the server properties
type server struct {
	SrvId                string
	Name                 string
	IP                   string
	Hostname             string
	OsUser               string
	OsPassword           string
	OsPort               string
	WebPort              string
	Product              string
	Datacenter           string
	WebPrefix            string
	WebSuffix            string
	Fav                  string
	DateTimeLastAccessed string
}

// template for index page
var templates = template.Must(template.ParseFiles("template\\index.html"))
var templates1 = template.Must(template.ParseFiles("template\\addPage.html"))
var templates2 = template.Must(template.ParseFiles("template\\editPage.html"))

// chkErr is common function for any error
func chkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// index function habled first index function
func index(w http.ResponseWriter, r *http.Request) {
	if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func addPage(w http.ResponseWriter, r *http.Request) {
	if err := templates1.ExecuteTemplate(w, "addPage.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// search function Query SQL and upload data
func search(w http.ResponseWriter, r *http.Request) {
	var result server
	var results []server
	filterStr := ""
	firstArg := true
	// get the form data entered in search-form with name "search"
	r.ParseForm()
	for key, value := range r.Form {
		//fmt.Printf("%s --%v", key, value)
		if len(value[0]) > 0 {
			if firstArg {
				filterStr = filterStr + " WHERE "
				firstArg = false
			} else {
				filterStr = filterStr + " AND "
			}
			filterStr = filterStr + " " + key + " like '%" + value[0] + "%'"
		}

	}
	//columns refered from sql server
	selectColumns := "srvId,name,ip,hostname,product,datacenter,dateTimeLastAccessed"
	queryString := "SELECT " + selectColumns + " FROM servers " +
		filterStr + " ORDER BY dateTimeLastAccessed DESC"
	//fmt.Printf("\n%s\n", queryString)
	// Open sqlite connection for dc.db. The table the data should be cretaed using csv_to_sql.go tool
	db, _ := sql.Open("sqlite3", "dc.db")
	rows, err := db.Query(queryString)
	chkErr(err)
	var dateTime time.Time
	for rows.Next() {
		err = rows.Scan(&result.SrvId, &result.Name, &result.IP, &result.Hostname, &result.Product,
			&result.Datacenter, &result.DateTimeLastAccessed)

		dateTime, err = time.Parse(time.RFC3339, result.DateTimeLastAccessed)
		chkErr(err)
		result.DateTimeLastAccessed = dateTime.Format("2006-Jan-02 15:04:05")
		results = append(results, result)
	}
	encoder := json.NewEncoder(w)
	if err := encoder.Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	db.Close()
	return
}

// Connect function will acton ssh request
func connect(w http.ResponseWriter, r *http.Request) {
	var result server
	//var results []server
	type accessData struct {
		SID  string
		Type string
	}
	decoder := json.NewDecoder(r.Body)
	var t accessData
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	var selectColumns string
	selectColumns = "ip,osPort,osUser,osPassword,webPort, webPrefix"
	queryString := "SELECT " + selectColumns + " FROM servers where srvId =" + t.SID

	db, _ := sql.Open("sqlite3", "dc.db")
	rows, err := db.Query(queryString)
	chkErr(err)

	for rows.Next() {
		err = rows.Scan(&result.IP, &result.OsPort, &result.OsUser, &result.OsPassword, &result.WebPort, &result.WebPrefix)
		chkErr(err)
	}
	var prog, progArg string
	//Denive the windows command string based on connect type requested
	switch t.Type {
	case "ssh":
		prog = "cmd"
		progArg = " /c putty " + result.OsUser + "@" + result.IP + " -pw " + result.OsPassword + " -P " + result.OsPort
	case "ftp":
		prog = "c:\\Program Files (x86)\\WinSCP\\WinSCP.exe"
		progArg = "sftp://" + result.OsUser + ":" + result.OsPassword + "@" + result.IP + ":" + result.OsPort
	case "web":
		prog = "cmd"
		progArg = " /c start " + result.WebPrefix + "://" + result.IP + ":" + result.WebPort
	}
	//Preate to update last access in sql
	stmt, err := db.Prepare("update servers set DateTimeLastAccessed=? where SrvId=?")
	chkErr(err)
	timeNow := time.Now().Format(time.RFC3339)
	stmt.Exec(timeNow, t.SID)
	chkErr(err)
	//windows command executed with start
	cmdInstance := exec.Command(prog, progArg)
	cmdInstance.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = cmdInstance.Output()
	chkErr(err)
}

func upload(w http.ResponseWriter, r *http.Request) {
	// add http request file to memeory
	r.ParseMultipartForm(32 << 20)
	//open the form file
	csvFile, _, err := r.FormFile("uploadfile")
	chkErr(err)
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)
	reader.FieldsPerRecord = -1
	csvData, err := reader.ReadAll()
	chkErr(err)
	//strip header ;and  send csvDada as string [][].
	res := addServerDb(csvData[1:])
	//write response to hrrp response
	w.Write([]byte(res))
	return
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	//create struct to match the reciving data
	type deleteData struct {
		DelSrvId string
	}
	var t deleteData
	//decode the recived reeq body in json format
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&t)
	chkErr(err)
	db, _ := sql.Open("sqlite3", "./dc.db")
	statement, _ := db.Prepare("DELETE FROM servers WHERE SrvId = ?")
	_, err = statement.Exec(t.DelSrvId)
	var res string
	if err != nil {
		res = err.Error()
	} else {
		res = "Server " + t.DelSrvId + " is deleted."
	}
	w.Write([]byte(res))
	db.Close()
	return
}

func editPage(w http.ResponseWriter, r *http.Request) {
	type editData struct {
		EdtSrvId string
	}
	var t editData
	//decode the recived reeq body in json format
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&t)
	chkErr(err)
	//fmt.Printf("%s\n", t.EdtSrvId)
	db, _ := sql.Open("sqlite3", "./dc.db")
	var result server
	selectColumns := "srvId,name,ip,hostname,osUser,osPassword,osPort," +
		"webPort,product,datacenter,webPrefix,webSuffix,fav"
	queryString := "SELECT " + selectColumns + " FROM servers where srvId =" + t.EdtSrvId
	rows, err := db.Query(queryString)
	chkErr(err)
	for rows.Next() {
		err = rows.Scan(&result.SrvId, &result.Name, &result.IP, &result.Hostname, &result.OsUser,
			&result.OsPassword, &result.OsPort, &result.WebPort, &result.Product, &result.Datacenter, &result.WebPrefix, &result.WebSuffix, &result.Fav)
		chkErr(err)
		if result.Fav == "true" {
			result.Fav = "checked"
		} else {
			result.Fav = ""
		}
	}
	if err := templates2.Execute(w, result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// function addServerDb add string to database used in addSubmit and upload dunctions
func addServerDb(ss [][]string) (res string) {
	res = "Server added: "
	db, _ := sql.Open("sqlite3", "./dc.db")
	statement, _ := db.Prepare("CREATE TABLE IF NOT EXISTS servers (srvId INTEGER PRIMARY KEY, name TEXT," +
		"ip TEXT,hostname TEXT,osUser TEXT,osPassword TEXT,osPort TEXT,webPort TEXT,product TEXT,datacenter TEXT," +
		"webPrefix TEXT,webSuffix TEXT, fav TEXT, dateTimeCreated TEXT, dateTimeModified TEXT, dateTimeLastAccessed TEXT)")
	statement.Exec()
	statement, _ = db.Prepare("INSERT INTO servers (name,ip,hostname,osUser,osPassword,osPort," +
		"webPort,product,datacenter,webPrefix,webSuffix,fav, dateTimeCreated, dateTimeModified," +
		"dateTimeLastAccessed ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	var timeNow string
	for _, s := range ss {
		timeNow = time.Now().Format(time.RFC3339)
		/*
			s[0] = Name, s[1] = IP, s[2] = Hostname, s[3] = OsUser, s[4] = OsPassword, s[5] = OsPort, s[6] = WebPort,
			s[7]=Product, s[8]=Datacenter, s[9]=WebPrefix, s[10]=WebSuffix, s[11]=Fav, timeNow = dateTimeCreated,
			 timeNow = dateTimeModified, timeNow = dateTimeLastAccessed
		*/
		_, err := statement.Exec(s[0], s[1], s[2], s[3], s[4], s[5], s[6],
			s[7], s[8], s[9], s[10], s[11], timeNow, timeNow, timeNow)
		if err != nil {
			res = res + " " + err.Error()
		} else {
			res = res + " " + s[0]
		}
	}
	db.Close()
	return res
}

// function addServerDb add string to database used in addSubmit and upload dunctions
func editServerDb(ss [][]string, srvId string) (res string) {
	res = "Server data modified: "
	db, _ := sql.Open("sqlite3", "./dc.db")
	dbQueryString := "UPDATE servers SET name=?,ip=?,hostname=?,osUser=?,osPassword=?,osPort=?," +
		"webPort=?,product=?,datacenter=?,webPrefix=?,webSuffix=?,fav=? where srvId =?"
	statement, _ := db.Prepare(dbQueryString)
	fmt.Printf("\nsrvID == %s\n", srvId)
	for _, s := range ss {
		/*
			srvId = SrvId,s[0] = Name, s[1] = IP, s[2] = Hostname, s[3] = OsUser, s[4] = OsPassword, s[6] = OsPort, s[7] = WebPort,
			s[8]=Product, s[9]=Datacenter, s[10]=WebPrefix, s[11]=WebSuffix, s[12]=Fav
		*/
		_, err := statement.Exec(s[0], s[1], s[2], s[3], s[4], s[5], s[6], s[7], s[8], s[9], s[10], s[11], srvId)
		if err != nil {
			res = res + " " + err.Error()
		} else {
			res = res + " " + s[0]
		}
	}
	db.Close()
	return res
}

func addEditSubmit(w http.ResponseWriter, r *http.Request) {
	// get the form data entered in add-form with name as in form
	r.ParseForm()
	//Converint checked input to /true/false
	favBool := "false"
	if _, ok := r.Form["fav"]; ok {
		if strings.EqualFold(r.Form["fav"][0], "true") || r.Form["fav"][0] == "on" {
			favBool = "true"
		}
	}
	fmt.Print("%v", r.Form)
	s := []string{r.Form["name"][0],
		r.Form["ip"][0], r.Form["hostname"][0], r.Form["osUser"][0], r.Form["osPassword"][0], r.Form["osPort"][0], r.Form["webPort"][0], r.Form["product"][0],
		r.Form["datacenter"][0], r.Form["webPrefix"][0], r.Form["webSuffix"][0], favBool}

	// define two dimention array 'ss'
	ss := make([][]string, 1)
	ss[0] = s
	// send ss to add to db
	res := ""
	switch r.Form["reqType"][0] {
	case "add":
		res = addServerDb(ss)
	case "edit":
		fmt.Printf("\nFrom Value== %s\n", r.Form["srvId"][0])
		res = editServerDb(ss, r.Form["srvId"][0])
	}
	//write response
	w.Write([]byte(res))
	return
}
func killMe(w http.ResponseWriter, r *http.Request) {
	os.Exit(3)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/search", search)
	r.HandleFunc("/connect", connect)
	r.HandleFunc("/addPage", addPage)
	r.HandleFunc("/deleteServer", deleteServer)
	r.HandleFunc("/upload", upload)
	r.HandleFunc("/editPage", editPage)
	r.HandleFunc("/addEditSubmit", addEditSubmit)
	r.HandleFunc("/killMe", killMe)
	//Specifying the http file location for CSS
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./template/")))
	http.Handle("/", r)
	fmt.Println(http.ListenAndServe(":8080", nil))
}
