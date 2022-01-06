//credits to https://github.com/xiaoqidun/entps/blob/main/entps.go
package sqlite3

import (
  "database/sql"
	"database/sql/driver"
  
  "modernc.org/sqlite"
)

type sqlite3Driver struct {
	*sqlite.Driver
}

func (d sqlite3Driver) Open(name string) (driver.Conn, error) {
	conn, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
  
  c := conn.(interface{Exec(stmt string, args []driver.Value) (driver.Result, error)})
  if _,err := c.Exec("PRAGMA foreign_keys = ON;", nil); err != nil {
    conn.Close()
    return nil, err
  }
	
	return conn, err
}

func init() {
	sql.Register("sqlite3", sqlite3Driver{Driver: &sqlite.Driver{}})
}