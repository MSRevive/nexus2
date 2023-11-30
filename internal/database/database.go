package database

import (

)

type Database interface {
	Connect(conn string) error
	Disconnect() error
}