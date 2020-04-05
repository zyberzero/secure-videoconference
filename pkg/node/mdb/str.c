#include "str.h"
#include <stdlib.h>

char** create_array(size_t size)
{
	return calloc(1, sizeof(char*) * size);
}

char* get_array(char** array, size_t i)
{
	return array[i];
}

void set_array(char** array, char* string, size_t i)
{
	array[i] = string;
}

void delete_array(char** array, size_t size)
{
	for (int i = 0; i < size; ++i)
		free(array[i]);
	free(array);
}
