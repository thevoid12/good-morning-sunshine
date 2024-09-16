package dbpkg

const (
	SCHEMA       = "CREATE TABLE IF NOT EXISTS email_record (id TEXT NOT NULL,email_id TEXT NOT NULL,expiry_date TEXT NOT NULL,created_on TIME NOT NULL,is_deleted BOOL NOT NULL);"
	OWNER_SCHEMA = "CREATE TABLE IF NOT EXISTS owner (id TEXT NOT NULL,email_id TEXT NOT NULL,rate_limit INT NOT NULL,created_on TIME NOT NULL,updated_on TIME NOT NULL,is_deleted BOOL NOT NULL);"

	//QUERY
	CREATE_EMAIL_RECORD_QUERY        = "INSERT OR REPLACE INTO email_record(id,email_id,expiry_date,created_on,is_deleted) VALUES(?,?,?,?,?);"
	HARD_DELETE_RECORD_QUERY         = "DELETE FROM email_record WHERE expiry_date< ?;"
	SOFT_DELETE_EXPIRED_RECORD_QUERY = "UPDATE TABLE email_record SET is_deleted = true WHERE expiry_date< ?"
	LIST_ACTIVE_EMAIL_RECORD_QUERY   = "SELECT * FROM email_record WHERE expiry_date< ? AND is_deleted=false;"

	CREATE_OWNER_QUERY                 = "INSERT OR REPLACE INTO owner(id,email_id,rate_limit,created_on,updated_on,is_deleted) VALUES (?,?,?,?,?,?);"
	GET_OWNER_DETAILS_BY_EMAILID_QUERY = "SELECT * FROM owner WHERE email_id = ?;"
	UPDATE_OWNER_RATE_LIMIT_QUERY      = "UPDATE owner SET rate_limit = ?, updated_on = ? WHERE email_id=?;"
)
