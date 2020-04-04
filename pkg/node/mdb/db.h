#ifndef MEETING_DB_H
#define MEETING_DB_H

// Call at application start / exit.
void open_db();
void close_db();

// Create a meeting. Returns meeting id. Attendees is an array of person
// numbers of length num_attendees as strings.
long create_meeting(int num_attendees, const char* attendees[]);

// Get all meetings for one person number stored into ids.
// This array is allocated by the caller and its size pased in max_size.
// The return value is the number of entries written.
int get_meetings(const char* personal_number, int* ids, int max_size);

#endif
