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
	key, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid meeting id", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if (err != nil) {
		fmt.Println(err)
		http.Error(w, "Failed to read request data", http.StatusBadRequest)
		return
	}

	var create CreateRequest
	err = json.Unmarshal(body, &create)
	if (err != nil) {
		fmt.Println(err)
		http.Error(w, "Failed to parse request data", http.StatusBadRequest)
		return
	}

	if (! meetingExists(key)) {
		createMeetingInDatabase(key, create.PersonNumbers)
		http.Error(w, "Meeting already exists", http.StatusBadRequest)
		return
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key := mux.Vars(r)["id"]

	meetings := GetResult{getMeetingsFromDatabase(key)}
	res, err := json.Marshal(meetings)

	if (err != nil) {
		http.Error(w, "Failed to serializer request data", http.StatusBadRequest)
		panic(err)
	}

	fmt.Fprintf(w, string(res))
}

func meetingExists(id int) bool {
	ret := int(C.check_meeting(C.long(id)))
	return ret != 0
}

func createMeetingInDatabase(id int, personalNumbers []string) {
	cArray := C.create_array(C.int(len(personalNumbers)))
	defer C.delete_array(cArray, C.int(len(personalNumbers)))

	for i := 0; i < len(personalNumbers); i++ {
		C.set_array(cArray, C.CString(personalNumbers[i]), C.int(i))
	}

	C.create_meeting(C.long(id), C.int(len(personalNumbers)), cArray)
}

func getMeetingsFromDatabase(personalNumber string) []int {
	const MAX_MEETINGS = 1024
	cInts := [MAX_MEETINGS]C.int{0}
	count := int(C.get_meetings(C.CString(personalNumber), &cInts[0], C.int(len(cInts))))

	meeting_ids := make([]int, count)
	for i := 0; i < count; i++ {
		meeting_ids[i] = int(cInts[i])
	}
	return meeting_ids
}
