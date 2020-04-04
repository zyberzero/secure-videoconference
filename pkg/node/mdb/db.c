#include "db.h"
#include "sqlite3.h"
#include <assert.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>
#include <pthread.h>

#define SECRET_KEY_DEFAULT "theveryhemliganyckleln"
#define SECRET_KEY_ENV     "MDB_SQLITE_KEY"
#define DB_FILENAME        "meetings.sqlite"

#define QUERY_CREATE       "CREATE TABLE IF NOT EXISTS invites ("   \
                           " id INTEGER PRIMARY KEY AUTOINCREMENT," \
                           " personal_number TEXT,"                 \
                           " meeting_id INTEGER) ;"

#define QUERY_INSERT       "INSERT INTO invites (personal_number, meeting_id)" \
                           " values (?, ?)"

#define QUERY_SELECT       "SELECT meeting_id FROM invites" \
                           " WHERE personal_number = ?"

#define QUERY_NEXT         "SELECT MAX(meeting_id) + 1 from invites;"

#define CHECK_ERROR(actual, expected) \
if (actual != expected)               \
{                                     \
    puts(sqlite3_errstr(rc));         \
    assert(0);                        \
}

static pthread_mutex_t create_lock = PTHREAD_MUTEX_INITIALIZER;
static sqlite3* db                 = NULL;

void open_db()
{
	int rc;
	rc = sqlite3_open(DB_FILENAME, &db);
	CHECK_ERROR(rc, SQLITE_OK);

	const char* key = getenv(SECRET_KEY_ENV);
	if (!key)
	{
		fputs("ERROR: using default encryption key", stderr);
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

static long create_meeting_id()
{
	int rc;
	sqlite3_stmt* statement;
	rc = sqlite3_prepare_v2(db, QUERY_NEXT, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	rc = sqlite3_step(statement);
	CHECK_ERROR(rc, SQLITE_ROW);

	long id = sqlite3_column_int(statement, 0);
	sqlite3_finalize(statement);

	return id;
}

long create_meeting(int num_attendees, const char* attendees[])
{
	if (num_attendees <= 0)
		return -1;

	assert(db);

	// Syncronize here so that the id cannot be allocated again before
	// it is inserted into the DB.
	pthread_mutex_lock(&create_lock);
	long meeting_id = create_meeting_id();

	sqlite3_stmt* statement;
	int rc;
	rc = sqlite3_prepare_v2(db, QUERY_INSERT, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);

	for (int i = 0; i < num_attendees; ++i)
	{
		rc = sqlite3_bind_text(statement, 1, attendees[0], -1, NULL);
		CHECK_ERROR(rc, SQLITE_OK);

		rc = sqlite3_bind_int(statement, 2, meeting_id);
		CHECK_ERROR(rc, SQLITE_OK);

		rc = sqlite3_step(statement);
		CHECK_ERROR(rc, SQLITE_DONE);

		rc = sqlite3_reset(statement);
		CHECK_ERROR(rc, SQLITE_OK);
	}

	pthread_mutex_unlock(&create_lock);

	sqlite3_finalize(statement);
	return meeting_id;
}

int get_meetings(const char* personal_number, int* ids, int max_size)
{
	assert(db);

	sqlite3_stmt* statement;
	int rc;
	rc = sqlite3_prepare_v2(db, QUERY_SELECT, -1, &statement, NULL);
	CHECK_ERROR(rc, SQLITE_OK);
	rc = sqlite3_bind_text(statement, 1, personal_number, -1, NULL);

	int count = 0;
	while (sqlite3_step(statement) == SQLITE_ROW)
	{
		ids[count] = sqlite3_column_int(statement, 0);

		if(++count == max_size)
			break;
	}

	sqlite3_finalize(statement);
	return count;
}
