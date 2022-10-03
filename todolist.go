package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

func Healthz(w http.ResponseWriter, t *http.Request) {
	log.Info("API is doing well")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive": "true"}`)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	log.WithFields(log.Fields{"description": description}).Info("Add new TodoItem. Saving to database...")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Value)
}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Check if the TodoITem is already in the database
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record not Found"}`)
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		todo.Completed = completed
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Check if the TodoITem is already in the database
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "Record not Found"}`)
	} else {
		log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
	}
}

func GetItemByID(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoItem not found in the database")
		return false
	}
	return true
}

func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Getting completed TodoItems")
	completedTodoItems := GetTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func GetIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Getting Incomplete TodoItems")
	completedTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func main() {
	defer db.Close()

	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting TodoList API Server...")
	router := mux.NewRouter()
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedItems).Methods("GET")
	router.HandleFunc("/todo-incomplete", GetIncompleteItems).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")
	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)
	http.ListenAndServe(":8080", handler)
}
