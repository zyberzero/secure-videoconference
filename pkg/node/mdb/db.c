#include "db.h"
#include "sqlite3.h"
#include <assert.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <pthread.h>

#define SECRET_KEY_DEFAULT  "theveryhemliganyckleln"
#define SECRET_KEY_ENV      "MDB_SQLITE_KEY"
#define DB_FILENAME         "/var/run/hold_space/hold_space.sqlite"

#define QUERY_CREATE        "CREATE TABLE IF NOT EXISTS invites ("   \
                            " id INTEGER PRIMARY KEY AUTOINCREMENT," \
                            " personal_number TEXT NOT NULL,"        \
                            " room_id INTEGER NOT NULL); "           \
                            "CREATE TABLE IF NOT EXISTS rooms ("     \
                            " id INTEGER PRIMARY KEY AUTOINCREMENT," \
                            " name TEXT UNIQUE NOT NULL);"

#define QUERY_INSERT_INVITE "INSERT INTO invites (personal_number, room_id)" \
                            " values (?, ?);"

#define QUERY_INSERT_ROOM   "INSERT INTO rooms (name)"     \
                            " values (?);"

#define QUERY_SELECT        "SELECT rooms.name FROM rooms"           \
                            " INNER JOIN invites"                    \
                            " ON rooms.id = invites.room_id"         \
                            " WHERE invites.personal_number = ?;"

#define CHECK_ERROR(actual, expected) \
if (actual != expected)               \
{                                     \
    puts(sqlite3_errstr(rc));         \
    assert(actual == expected);       \
}

static sqlite3* db                 = NULL;
static pthread_mutex_t create_lock = PTHREAD_MUTEX_INITIALIZER;

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
	int rc;
	sqlite3_stmt* statement;

	rc = sqlite3_prepare_v2(db, QUERY_INSERT_ROOM, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	rc = sqlite3_bind_text(statement, 1, name, -1, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	// Syncronize because of last_insert_rowid()
	pthread_mutex_lock(&create_lock);
	rc = sqlite3_step(statement);
	long id = sqlite3_last_insert_rowid(db);
	pthread_mutex_unlock(&create_lock);

	if (rc == SQLITE_CONSTRAINT)
	{
		sqlite3_finalize(statement);
		return -1;
	}

	CHECK_ERROR(rc, SQLITE_DONE);

	sqlite3_finalize(statement);
	return id;
}

static void insert_attendees(long id, size_t num_attendees, char** attendees)
{
	sqlite3_stmt* statement;
	int rc;
	rc = sqlite3_prepare_v2(db, QUERY_INSERT_INVITE, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	// Syncronize because of last_insert_rowid() in create_room_id
	pthread_mutex_lock(&create_lock);

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

	pthread_mutex_unlock(&create_lock);

	sqlite3_finalize(statement);
}

bool create_room(const char* name, size_t num_attendees, char** attendees)
{
	if (num_attendees <= 0)
		return false;

	assert(db);

	long id = create_room_id(name);

	if (id == -1)
		return false;

	insert_attendees(id, num_attendees, attendees);
	return true;
}

size_t get_rooms(const char* personal_number, char** names, size_t max_size)
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
		names[count] = strdup(sqlite3_column_text(statement, 0));

		if(++count == max_size)
			break;
	}

	sqlite3_finalize(statement);
	return count;
}
