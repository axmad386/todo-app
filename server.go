package main

import (
	api "axmad386/todo-app/http"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type Activity struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Email     string     `gorm:"size:255" json:"email"`
	Title     string     `gorm:"size:255" json:"title"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"update_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}
type Todo struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	ActivityGroupID uint       `gorm:"index" json:"activity_group_id"`
	Title           string     `gorm:"size:255" json:"title"`
	IsActive        bool       `gorm:"boolean" json:"is_active"`
	Priority        string     `gorm:"size:255" json:"priority"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"update_at"`
	DeletedAt       *time.Time `json:"deleted_at"`
}

var activities = []Activity{}
var todos = []Todo{}

// var db *gorm.DB = nil

type dbStatus struct {
	// created bool
	// updated bool
	// deleted bool
	cached bool
}
type cached struct {
	activities []Activity
	todos      []Todo
}

var activityStatus = dbStatus{}
var todoStatus = dbStatus{}
var cachedData = cached{}

// func Async(query func()) chan int {
// 	r := make(chan int)
// 	go func() {
// 		defer close(r)
// 		query()
// 		r <- 1
// 	}()
// 	return r
// }

func main() {
	// dbstring := os.Getenv("MYSQL_USER") + ":" + os.Getenv("MYSQL_PASSWORD") + "@tcp(" + os.Getenv("MYSQL_HOST") + ":3306)/" + os.Getenv("MYSQL_DBNAME")
	// DB, err := gorm.Open(mysql.Open(dbstring), &gorm.Config{})
	// if err == nil {
	// 	db = DB
	// 	db.AutoMigrate(&Activity{}, &Todo{})
	// 	db.Exec("TRUNCATE TABLE todos")
	// 	db.Exec("TRUNCATE TABLE activities")
	// } else {
	// 	fmt.Println(err.Error(), "error")
	// 	return
	// }

	http.HandleFunc("/", Controller)
	http.ListenAndServe("0.0.0.0:3030", nil)
}

func Controller(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	if path == "/activity-groups" {
		switch r.Method {
		case "GET":
			if activityStatus.cached {
				api.Success(w, cachedData.activities, http.StatusOK)
				return
			}
			data := []Activity{}
			for i := range activities {
				if activities[i].DeletedAt == nil {
					data = append(data, activities[i])
				}
			}
			api.Success(w, data, http.StatusOK)
			cachedData.activities = data
			activityStatus.cached = true
		case "POST":
			decoder := json.NewDecoder(r.Body)
			var t Activity
			decoder.Decode(&t)
			if t.Title == "" {
				api.CantNull(w, "title")
				return
			}
			t.ID = uint(len(activities)) + 1
			t.CreatedAt = time.Now().UTC()
			t.UpdatedAt = time.Now().UTC()
			api.Success(w, t, http.StatusCreated)
			activities = append(activities, t)
			activityStatus.cached = false
			// if activityStatus.created {
			// 	return
			// }
			// Async(func() {
			// 	db.Create(t)
			// })
			// activityStatus.created = true
		}
		return
	}
	if path == "/todo-items" {
		switch r.Method {
		case "GET":
			data := []Todo{}
			ActivityGroupID := r.URL.Query().Get("activity_group_id")
			if todoStatus.cached && ActivityGroupID == "" {
				api.Success(w, cachedData.todos, http.StatusOK)
				return
			}
			for i := range todos {
				if todos[i].DeletedAt == nil {
					if ActivityGroupID != "" {
						ActivityGroupIDInt, _ := strconv.ParseUint(ActivityGroupID, 10, 32)
						if todos[i].ActivityGroupID == uint(ActivityGroupIDInt) {
							data = append(data, todos[i])
						}
					} else {
						data = append(data, todos[i])
						cachedData.todos = data
						todoStatus.cached = true
					}
				}
			}
			api.Success(w, data, http.StatusOK)
		case "POST":
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
			t.ID = uint(len(todos)) + 1
			t.IsActive = true
			t.Priority = "very-high"
			t.CreatedAt = time.Now().UTC()
			t.UpdatedAt = time.Now().UTC()
			api.Success(w, t, http.StatusCreated)
			todos = append(todos, t)
			todoStatus.cached = false
			// if todoStatus.created {
			// 	return
			// }
			// Async(func() {
			// 	db.Create(t)
			// })
			// todoStatus.created = true
		}
		return
	}

	if len(path) > 16 && path[:16] == "/activity-groups" {
		id, _ := strconv.Atoi(path[17:])
		if id >= 0 && (len(activities) < id || id == 0) {
			api.NotFound(w, id, "Activity", "")
			return
		}
		switch r.Method {
		case "GET":
			// if id > 0 {
			if activities[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Activity", "")
				return
			}
			api.Success(w, activities[id-1], http.StatusOK)
			return
			// }

		case "PATCH":
			decoder := json.NewDecoder(r.Body)
			var t = activities[id-1]
			decoder.Decode(&t)
			if t.Title == "" {
				api.CantNull(w, "title")
				return
			}
			t.UpdatedAt = time.Now().UTC()
			api.Success(w, t, http.StatusOK)
			activities[id-1] = t
			activityStatus.cached = false

			// if activityStatus.updated {
			// 	return
			// }
			// Async(func() {
			// 	db.Where(Activity{ID: t.ID}).Updates(t)
			// })
			// activityStatus.updated = true
		case "DELETE":
			if activities[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Activity", "")
				return
			}
			api.Success(w, api.Blank, http.StatusOK)
			var now = time.Now().UTC()
			activities[id-1].DeletedAt = &now
			activityStatus.cached = false

			// if activityStatus.deleted {
			// 	return
			// }
			// Async(func() {
			// 	db.Delete(Activity{ID: uint(id)})
			// })
			// activityStatus.deleted = true
		default:
			http.Error(w, "", http.StatusBadRequest)
		}
		return
	}
	if path[:11] == "/todo-items" {
		id, _ := strconv.Atoi(path[12:])
		if id >= 0 && (len(todos) < id || id == 0) {
			api.NotFound(w, id, "Todo", "")
			return
		}
		switch r.Method {
		case "GET":
			// if id > 0 {
			if todos[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Todo", "")
				return
			}
			api.Success(w, todos[id-1], http.StatusOK)
			return
			// }

		case "PATCH":
			decoder := json.NewDecoder(r.Body)
			var t = todos[id-1]
			decoder.Decode(&t)
			t.UpdatedAt = time.Now().UTC()
			api.Success(w, t, http.StatusOK)
			todos[id-1] = t
			todoStatus.cached = false
			// if todoStatus.updated {
			// 	return
			// }
			// Async(func() {
			// 	db.Where(Todo{ID: t.ID}).Updates(t)
			// })
			// todoStatus.updated = true
		case "DELETE":
			if todos[id-1].DeletedAt != nil {
				api.NotFound(w, id, "Todo", "")
				return
			}
			api.Success(w, api.Blank, http.StatusOK)
			var now = time.Now().UTC()
			todos[id-1].DeletedAt = &now
			todoStatus.cached = false
			// if todoStatus.deleted {
			// 	return
			// }
			// Async(func() {
			// 	db.Delete(Todo{ID: uint(id)})
			// })
			// todoStatus.deleted = true
		}
		return
	}
	http.Error(w, "", http.StatusBadRequest)

}
