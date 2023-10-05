package mongodb

import (
	//"time"
	//"context"

	//"go.mongodb.org/mongo-driver/mongo"
)

func (d *mongoDB) ProcessQueue() error {
	return nil
}

func (d *mongoDB) GetQueueSize() int {
	return len(d.userQueue) + len(d.serverQueue)
}