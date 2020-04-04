#ifndef MEETING_DB_H
#define MEETING_DB_H

void open_db();
void close_db();

long create_meeting(int num_attendees, const char* attendees[]);
int get_meetings(const char* personal_number, int* ids, int max_size);

#endif
