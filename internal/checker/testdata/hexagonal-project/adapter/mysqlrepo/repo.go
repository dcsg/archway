package mysqlrepo

// FindByID has SQL concatenation — should trigger sql_concatenation.
func FindByID(id string) string {
	return "SELECT * FROM orders WHERE id=" + id
}
