package main

// #cgo LDFLAGS: -ldl -pthread -lssl -lcrypto
// #cgo CFLAGS: -DSQLITE_HAS_CODEC
// #include "db.h"
// #include "str.h"
import "C"

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "encoding/json"
    "github.com/gorilla/mux"
)

type GetResponse struct {
	Meetings []int
}

type CreateRequest struct {
	PersonNumbers []string
}

type CreateResponse struct {
	Meeting int
}


func main() {
	C.open_db()
	defer C.close_db()

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/create", create)
	router.HandleFunc("/get/{id}", get)
	log.Fatal(http.ListenAndServe(":8081", router))
}

func create(w http.ResponseWriter, r *http.Request) {
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

	response := CreateResponse {createMeetingInDatabase(create.PersonNumbers)}
	res, err := json.Marshal(response)

	if (err != nil) {
		http.Error(w, "Failed to serializer request data", http.StatusBadRequest)
		panic(err)
	}

	fmt.Fprintf(w, string(res))

}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key := mux.Vars(r)["id"]

	meetings := GetResponse{getMeetingsFromDatabase(key)}
	res, err := json.Marshal(meetings)

	if (err != nil) {
		http.Error(w, "Failed to serializer request data", http.StatusBadRequest)
		panic(err)
	}

	fmt.Fprintf(w, string(res))
}

func createMeetingInDatabase(personalNumbers []string) int {
	cArray := C.create_array(C.int(len(personalNumbers)))
	defer C.delete_array(cArray, C.int(len(personalNumbers)))

	for i := 0; i < len(personalNumbers); i++ {
		C.set_array(cArray, C.CString(personalNumbers[i]), C.int(i))
	}

	return int(C.create_meeting(C.int(len(personalNumbers)), cArray))
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
