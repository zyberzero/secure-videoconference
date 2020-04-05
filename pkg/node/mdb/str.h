#ifndef HOLD_SPACE_STR_H
#define HOLD_SPACE_STR_H

#include <stddef.h>

// Helpers to create and manage a char** from golang
char** create_array(size_t size);
char* get_array(char** array, size_t i);
void set_array(char** array, char* string, size_t i);
void delete_array(char** array, size_t size);

#endif
