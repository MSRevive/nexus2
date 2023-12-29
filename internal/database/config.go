package database

import (

)

type Config struct {
	MongoDB struct {
		Connection string
	}
	Clover struct {
		Connection string
	}
}