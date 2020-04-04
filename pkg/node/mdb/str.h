#ifndef MEETING_STR_H
#define MEETING_STR_H

// Helpers to create and manage a char** from golang
const char** create_array(int size);
void set_array(char** array, char* string, int i);
void delete_array(char** array, int size);

#endif
