#ifndef MEETING_DB_H
#define MEETING_DB_H

void open_db();
void close_db();

typedef enum {false, true} bool;

void create_meeting(long meeting_id, int num_attendees, const char* attendees[]);
int get_meetings(const char* personal_number, int* ids, int max_size);
bool check_meeting(long meeting_id);

#endif
