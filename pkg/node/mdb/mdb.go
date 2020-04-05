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
    "github.com/gorilla/mux" // TODO: review license
)

// The protocol consists of two endpoints /get/{personNumber},
// and /create. These structs serialize to json specify the protocol.
type Room struct {
	RoomName  string       `json:"roomName"`
}

type GetResponse struct {
	Rooms []Room           `json:"rooms"`
}

type CreateRequest struct {
	RoomName        string `json:"roomName"`
	PersonNumbers []string `json:"personNumbers"`
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

	if (len(create.PersonNumbers) == 0) {
		http.Error(w, "No attendees specified", http.StatusBadRequest)
		return
	}

	if (! createRoomInDatabase(create.PersonNumbers, create.RoomName) ) {
		http.Error(w, "Failed to create room", http.StatusBadRequest)
		return
	}
}

func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key := mux.Vars(r)["id"]

	rooms := GetResponse{getRoomsFromDatabase(key)}
	res, err := json.Marshal(rooms)

	if (err != nil) {
		http.Error(w, "Failed to serializer request data", http.StatusBadRequest)
		panic(err)
	}

	fmt.Fprintf(w, string(res))
}

func createRoomInDatabase(personalNumbers []string, roomName string) bool {
	cArray := C.create_array(C.size_t(len(personalNumbers)))
	defer C.delete_array(cArray, C.size_t(len(personalNumbers)))

	for i := 0; i < len(personalNumbers); i++ {
		C.set_array(cArray, C.CString(personalNumbers[i]), C.size_t(i))
	}

	return int(C.create_room(C.CString(roomName), C.size_t(len(personalNumbers)), cArray)) != 0
}

func getRoomsFromDatabase(personalNumber string) []Room {
	const MAX_ROOMS = 1024

	cArray := C.create_array(C.size_t(MAX_ROOMS))

	count := int(C.get_rooms(C.CString(personalNumber), cArray, C.size_t(MAX_ROOMS)))

	rooms := make([]Room, count)
	for i := 0; i < count; i++ {
		rooms[i].RoomName = C.GoString(C.get_array(cArray, C.size_t(i)))
	}
	C.delete_array(cArray, C.size_t(count));

	return rooms
}
