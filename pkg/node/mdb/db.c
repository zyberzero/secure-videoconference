#include "db.h"
#include "sqlite3.h"
#include <assert.h>
#include <string.h>
#include <stdlib.h>
#include <stdio.h>

// TODO: fetch from somewhere
const char* SECRET_KEY = "theveryhemliganyckleln";

const char* DB_FILENAME = "meetings.sqlite";

const char* QUERY_CREATE = "CREATE TABLE IF NOT EXISTS invites ("
                           " id INTEGER PRIMARY KEY AUTOINCREMENT,"
                           " personal_number TEXT,"
                           " meeting_id INTEGER) ;";

const char* QUERY_INSERT = "INSERT INTO invites (personal_number, meeting_id)"
                           " values (?, ?)";

const char* QUERY_SELECT = "SELECT meeting_id FROM invites"
                           " WHERE personal_number = ?";

#define CHECK_ERROR(actual, expected) \
if (actual != expected) \
{ \
		puts(sqlite3_errstr(rc)); \
		assert(0); \
}

static sqlite3* db = NULL;

void open_db()
{
	int rc;
	rc = sqlite3_open(DB_FILENAME, &db);
	CHECK_ERROR(rc, SQLITE_OK);

	const char* key = getenv("MDB_SQLITE_KEY");
	if (!key)
	{
		fprintf(stderr, "WARNING: using default encryption key\n");
		key = SECRET_KEY;
	}
	rc = sqlite3_key(db, key, strlen(SECRET_KEY));
	CHECK_ERROR(rc, SQLITE_OK);

	char* err;
	rc = sqlite3_exec(db, QUERY_CREATE, NULL, NULL, &err);
	CHECK_ERROR(rc, SQLITE_OK);
}

void close_db()
{
	sqlite3_close(db);
}

void create_meeting(long meeting_id, int num_attendees, const char* attendees[])
{
	assert(db);
	// TODO: check that meeting doesn't exist

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

	sqlite3_finalize(statement);
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
