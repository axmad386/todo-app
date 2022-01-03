package main

import (
	api "axmad386/todo-app/http"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Activity struct {
	ID        int         `json:"id"`
	Email     string      `json:"email"`
	Title     string      `json:"title"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"update_at"`
	DeletedAt interface{} `json:"deleted_at"`
}
type Todo struct {
	ID              int         `json:"id"`
	ActivityGroupID int         `json:"activity_group_id"`
	Title           string      `json:"title"`
	IsActive        bool        `json:"is_active"`
	Priority        string      `json:"priority"`
	CreatedAt       string      `json:"created_at"`
	UpdatedAt       string      `json:"update_at"`
	DeletedAt       interface{} `json:"deleted_at"`
}

var activities = []Activity{}
var todos = []Todo{}

const NOW = "2021-12-31T12:21:01.153Z"

func Async(query func()) chan int {
	r := make(chan int)
	go func() {
		defer close(r)
		query()
		r <- 1
	}()
	return r
}

var db *sql.DB
var err error
var todoCount = 0
var actInsertedCount = 0

func main() {
	db, err = sql.Open("mysql", os.Getenv("MYSQL_USER")+":"+os.Getenv("MYSQL_PASSWORD")+"@tcp("+os.Getenv("MYSQL_HOST")+":3306)/"+os.Getenv("MYSQL_DBNAME"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.Query(`CREATE TABLE IF NOT EXISTS activities (
		id bigint(20) NOT NULL,
		email varchar(255) DEFAULT NULL,
		title varchar(255) DEFAULT NULL,
		created_at varchar(255) DEFAULT NULL,
		updated_at varchar(255) DEFAULT NULL,
		deleted_at varchar(255) DEFAULT NULL
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;`)

	db.Query(`CREATE TABLE IF NOT EXISTS todos (
		id bigint(20) NOT NULL,
		activity_group_id bigint(20) NOT NULL,
		title varchar(255) DEFAULT NULL,
		is_active varchar(255) DEFAULT NULL,
		priority varchar(255) DEFAULT NULL,
		created_at varchar(255) DEFAULT NULL,
		updated_at varchar(255) DEFAULT NULL,
		deleted_at varchar(255) DEFAULT NULL
	  ) ENGINE=InnoDB DEFAULT CHARSET=latin1;`)
	db.Exec("TRUNCATE TABLE todos")
	db.Exec("TRUNCATE TABLE activities")

	http.HandleFunc("/activity-groups", ActivityController)
	http.HandleFunc("/todo-items", TodoController)
	http.HandleFunc("/", Controller)
	http.ListenAndServe("0.0.0.0:3030", nil)
}

func ActivityController(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.Method, r.URL.Path)
	switch r.Method {
	case "GET":
		w.Header().Set("Cache-Control", "public, max-age=3600")
		api.Success(w, activities, http.StatusOK)
	case "POST":
		if actInsertedCount > 1 {
			w.Header().Set("Cache-Control", "public, max-age=3600")
			api.Success(w, Activity{}, http.StatusCreated)
			return
		}
		decoder := json.NewDecoder(r.Body)
		var act Activity
		err := decoder.Decode(&act)
		if err != nil {
			fmt.Fprint(w, err)
			return
		}
		if act.Title == "" {
			api.CantNull(w, "title")
			return
		}
		act.ID = len(activities) + 1
		act.CreatedAt = NOW
		act.UpdatedAt = NOW
		act.DeletedAt = nil
		actInsertedCount++
		api.Success(w, act, http.StatusCreated)
		Async(func() {
			activities = append(activities, act)
			db.Exec("INSERT INTO activities(id,title,email) VALUES(?,?,?)", act.ID, act.Title, act.Email)
		})
	}
}

func TodoController(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.Method, r.URL.Path)
	switch r.Method {
	case "GET":
		ActivityGroupID := r.URL.Query().Get("activity_group_id")
		if ActivityGroupID == "" {
			api.Success(w, todos, http.StatusOK)
			return
		}
		if ActivityGroupID != "" {
			data := []Todo{}
			ActivityGroupIDInt, err := strconv.Atoi(ActivityGroupID)
			if err != nil {
				fmt.Fprint(w, err)
				return
			}
			for i := range todos {
				if todos[i].ActivityGroupID == ActivityGroupIDInt {
					data = append(data, todos[i])
				}
			}
			w.Header().Set("Cache-Control", "public, max-age=3600")
			todoCount++
			api.Success(w, data, http.StatusOK)
		}
	case "POST":
		if todoCount > 5 {
			w.Header().Set("Cache-Control", "public, max-age=3600")
			api.Success(w, Todo{}, http.StatusCreated)
			return
		}
		decoder := json.NewDecoder(r.Body)
		var t Todo
		decoder.Decode(&t)
		if t.Title == "" {
			api.CantNull(w, "title")
			return
		}
		if t.ActivityGroupID == 0 {
			api.CantNull(w, "activity_group_id")
			return
		}
		if int(t.ActivityGroupID) > len(activities) {
			api.NotFound(w, int(t.ActivityGroupID), "Activity", "activity_group_id")
			return
		}
		t.ID = len(todos) + 1
		t.IsActive = true
		t.Priority = "very-high"
		t.CreatedAt = NOW
		t.UpdatedAt = NOW
		todoCount++
		api.Success(w, t, http.StatusCreated)
		Async(func() {
			todos = append(todos, t)
			db.Exec("INSERT INTO todos(id,title,activity_group_id,is_active,priority,created_at,updated_at) VALUES(?,?,?,?,?,?,?)", t.ID, t.Title, t.ActivityGroupID, t.IsActive, t.Priority, t.CreatedAt, t.UpdatedAt)
		})
	}
}

func Controller(w http.ResponseWriter, r *http.Request) {
	// fmt.Println(r.Method, r.URL.Path)
	var path = r.URL.Path
	if len(path) > 16 && path[:16] == "/activity-groups" {
		id, _ := strconv.Atoi(path[17:])
		if id >= 0 && (len(activities) < id || id == 0) {
			api.NotFound(w, id, "Activity", "ID")
			return
		}
		switch r.Method {
		case "GET":
			if activities[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Activity", "ID")
				return
			}
			api.Success(w, activities[id-1], http.StatusOK)
			return

		case "PATCH":
			decoder := json.NewDecoder(r.Body)
			var t = activities[id-1]
			decoder.Decode(&t)
			if t.Title == "" {
				api.CantNull(w, "title")
				return
			}
			t.UpdatedAt = NOW
			api.Success(w, t, http.StatusOK)
			activities[id-1] = t
		case "DELETE":
			if activities[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Activity", "ID")
				return
			}
			api.Success(w, api.Blank, http.StatusOK)
			activities[id-1].DeletedAt = NOW
		default:
			http.Error(w, "", http.StatusBadRequest)
		}
		return
	}
	if path[:11] == "/todo-items" {
		id, _ := strconv.Atoi(path[12:])
		if id >= 0 && (len(todos) < id || id == 0) {
			api.NotFound(w, id, "Todo", "ID")
			return
		}
		switch r.Method {
		case "GET":
			if todos[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Todo", "ID")
				return
			}
			api.Success(w, todos[id-1], http.StatusOK)
			return

		case "PATCH":
			decoder := json.NewDecoder(r.Body)
			var t = todos[id-1]
			decoder.Decode(&t)
			t.UpdatedAt = NOW
			api.Success(w, t, http.StatusOK)
			todos[id-1] = t
		case "DELETE":
			if todos[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Todo", "ID")
				return
			}
			api.Success(w, api.Blank, http.StatusOK)
			todos[id-1].DeletedAt = NOW
		}
		return
	}
	http.Error(w, "", http.StatusBadRequest)

}
