#include "str.h"
#include <stdlib.h>

const char** create_array(int size)
{
	return malloc(sizeof(char*) * size);
}

void set_array(char** array, char* string, int i)
{
	array[i] = string;
}

void delete_array(char** array, int size)
{
	for (int i = 0; i < size; ++i)
		free(array[i]);
	free(array);
}
