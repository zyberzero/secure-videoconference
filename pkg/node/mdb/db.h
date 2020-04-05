#ifndef HOLD_SPACE_DB_H
#define HOLD_SPACE_DB_H

#include <stddef.h>

typedef enum {false, true} bool;

// Call at application start / exit.
void open_db();
void close_db();

// Create a room. Returns room id. Attendees is an array of person
// numbers of length num_attendees as strings.
bool create_room(const char* name, size_t num_attendees, char** attendees);

// Get all room for one person number stored in names.
// This array is allocated by the called and its size is max size.
// The return value is the number of entries written.
size_t get_rooms(const char* personal_number, char** names, size_t max_size);

#endif
