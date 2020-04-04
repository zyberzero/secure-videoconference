package main

// #cgo LDFLAGS: -ldl -pthread -lssl -lcrypto
// #cgo CFLAGS: -DSQLITE_HAS_CODEC
// #include "db.h"
// #include "str.h"
import "C"

import (
    "fmt"
    "log"
    "strconv"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "github.com/gorilla/mux"
)

type GetResult struct {
	Meetings []int
}

type CreateRequest struct {
	PersonNumbers []string
}


func main() {
	C.open_db()
	defer C.close_db()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/create/{id}", create)
	router.HandleFunc("/get/{id}", get)
    log.Fatal(http.ListenAndServe(":8081", router))
}

func create(w http.ResponseWriter, r *http.Request) {
	key, err1 := strconv.Atoi(mux.Vars(r)["id"])
	if err1 != nil {
		fmt.Println(err1)
	}

	defer r.Body.Close()
    body, err2 := ioutil.ReadAll(r.Body)
	if (err2 != nil) {
		fmt.Println(err2)
	}

	var create CreateRequest
	err3 := json.Unmarshal(body, &create)
	if (err3 != nil) {
		fmt.Println(err3)
	}

	if (! meetingExists(key)) {
		createMeeting(key, create.PersonNumbers)
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	 w.Header().Set("Content-Type", "application/json")

	key := mux.Vars(r)["id"]

	meetings := GetResult{getMeetings(key)}
	res, err3 := json.Marshal(meetings)

	if (err3 != nil) {
		fmt.Println(err3)
	}

	fmt.Fprintf(w, string(res))
}

func meetingExists(id int) bool {
	// TODO:
	return false
}

func createMeeting(id int, personalNumbers []string) {
	cArray := C.create_array(C.int(len(personalNumbers)))
	defer C.delete_array(cArray, C.int(len(personalNumbers)))

	for i := 0; i < len(personalNumbers); i++ {
		C.set_array(cArray, C.CString(personalNumbers[i]), C.int(i))
	}

	C.create_meeting(C.long(id), C.int(len(personalNumbers)), cArray)
}

func getMeetings(personalNumber string) []int {
	const MAX_MEETINGS = 1024
	cInts := [MAX_MEETINGS]C.int{0}
	count := int(C.get_meetings(C.CString(personalNumber), &cInts[0], C.int(len(cInts))))

	meeting_ids := make([]int, count)
	for i := 0; i < count; i++ {
		meeting_ids[i] = int(cInts[i])
	}
	return meeting_ids
}
