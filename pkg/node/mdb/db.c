#include "db.h"
#include "sqlite3.h"
#include <assert.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <pthread.h>

#define SECRET_KEY_DEFAULT  "theveryhemliganyckleln"
#define SECRET_KEY_ENV      "MDB_SQLITE_KEY"
#define DB_FILENAME         "hold_space.sqlite"

#define QUERY_CREATE        "CREATE TABLE IF NOT EXISTS invites ("   \
                            " id INTEGER PRIMARY KEY AUTOINCREMENT," \
                            " personal_number TEXT,"                 \
                            " room_id INTEGER); "                 \
                            "CREATE TABLE IF NOT EXISTS rooms ("  \
                            " id INTEGER PRIMARY KEY,"                \
                            " name TEXT);"

#define QUERY_INSERT_INVITE "INSERT INTO invites (personal_number, room_id)" \
                            " values (?, ?)"

#define QUERY_INSERT_ROOM   "INSERT INTO rooms (id, name)" \
                            " values (?, ?)"

#define QUERY_SELECT        "SELECT rooms.id, rooms.name FROM rooms" \
                            " INNER JOIN invites"                             \
                            " ON rooms.id = invites.room_id"            \
                            " WHERE invites.personal_number = ?;"

#define QUERY_NEXT_ID       "SELECT MAX(id) + 1 from rooms;"

#define CHECK_ERROR(actual, expected) \
if (actual != expected)               \
{                                     \
    puts(sqlite3_errstr(rc));         \
    assert(actual == expected);       \
}

static sqlite3* db                 = NULL;

void open_db()
{
	int rc;
	rc = sqlite3_open(DB_FILENAME, &db);
	CHECK_ERROR(rc, SQLITE_OK);

	const char* key = getenv(SECRET_KEY_ENV);
	if (!key)
	{
		fputs("ERROR: using default encryption key\n", stderr);
		key = SECRET_KEY_DEFAULT;
	}
	rc = sqlite3_key(db, key, strlen(key));
	CHECK_ERROR(rc, SQLITE_OK);

	char* err;
	rc = sqlite3_exec(db, QUERY_CREATE, NULL, NULL, &err);
	CHECK_ERROR(rc, SQLITE_OK);
}

void close_db()
{
	sqlite3_close(db);
	db = NULL;
}

static long create_room_id(const char* name)
{
	static pthread_mutex_t create_lock = PTHREAD_MUTEX_INITIALIZER;
	// Syncronize here so that the id cannot be allocated again before
	// it is inserted into the DB.

	// TODO: should names be unique?

	pthread_mutex_lock(&create_lock);

	int rc;
	sqlite3_stmt* statement;
	rc = sqlite3_prepare_v2(db, QUERY_NEXT_ID, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	rc = sqlite3_step(statement);
	CHECK_ERROR(rc, SQLITE_ROW);

	long id = sqlite3_column_int64(statement, 0);
	sqlite3_finalize(statement);

	assert(id >= 0);

	rc = sqlite3_prepare_v2(db, QUERY_INSERT_ROOM, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	rc = sqlite3_bind_int64(statement, 1, id);
	CHECK_ERROR(rc, SQLITE_OK);

	rc = sqlite3_bind_text(statement, 2, name, -1, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	rc = sqlite3_step(statement);
	CHECK_ERROR(rc, SQLITE_DONE);

	sqlite3_finalize(statement);

	pthread_mutex_unlock(&create_lock);

	return id;
}

static void insert_attendees(long id, size_t num_attendees, char** attendees)
{
	sqlite3_stmt* statement;
	int rc;
	rc = sqlite3_prepare_v2(db, QUERY_INSERT_INVITE, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	for (int i = 0; i < num_attendees; ++i)
	{
		rc = sqlite3_bind_text(statement, 1, attendees[i], -1, NULL);
		CHECK_ERROR(rc, SQLITE_OK);

		rc = sqlite3_bind_int64(statement, 2, id);
		CHECK_ERROR(rc, SQLITE_OK);

		rc = sqlite3_step(statement);
		CHECK_ERROR(rc, SQLITE_DONE);

		rc = sqlite3_reset(statement);
		CHECK_ERROR(rc, SQLITE_OK);
	}

	sqlite3_finalize(statement);
}

long create_room(const char* name, size_t num_attendees, char** attendees)
{
	if (num_attendees <= 0)
		return -1;

	assert(db);

	long id = create_room_id(name);
	insert_attendees(id, num_attendees, attendees);
	return id;
}

size_t get_rooms(const char* personal_number, long* ids, char** names, size_t max_size)
{
	assert(db);

	sqlite3_stmt* statement;
	int rc;
	rc = sqlite3_prepare_v2(db, QUERY_SELECT, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);
	rc = sqlite3_bind_text(statement, 1, personal_number, -1, NULL);

	size_t count = 0;
	while (sqlite3_step(statement) == SQLITE_ROW)
	{
		ids[count] = sqlite3_column_int64(statement, 0);
		names[count] = strdup(sqlite3_column_text(statement, 1));

		if(++count == max_size)
			break;
	}

	sqlite3_finalize(statement);
	return count;
}
