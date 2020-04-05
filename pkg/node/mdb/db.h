#ifndef HOLD_SPACE_DB_H
#define HOLD_SPACE_DB_H

#include <stddef.h>

// Call at application start / exit.
void open_db();
void close_db();

// Create a room. Returns room id. Attendees is an array of person
// numbers of length num_attendees as strings.
long create_room(const char* name, size_t num_attendees, char** attendees);

// Get all room for one person number stored into ids and names.
// These arrays must be of the same size, is allocated by the caller and 
// their size pased in max_size.
// The return value is the number of entries written.
size_t get_rooms(const char* personal_number, long* ids, char** names, size_t max_size);

#endif
