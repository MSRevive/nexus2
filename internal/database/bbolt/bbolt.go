package bbolt

import (
	"fmt"
	"time"
	"errors"
	
	"github.com/msrevive/nexus2/internal/database"
	"github.com/msrevive/nexus2/internal/database/schema"
	"github.com/msrevive/nexus2/internal/database/bsoncoder"

	"github.com/google/uuid"
	"go.etcd.io/bbolt"
	//"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrNoDocument = errors.New("no document")
)

type bboltDB struct {
	db *bbolt.DB
}

func New() *bboltDB {
	return &bboltDB{}
}

func (d *bboltDB) Connect(cfg database.Config) error {
	timeout := cfg.BBolt.Timeout * time.Second
	db, err := bbolt.Open(cfg.BBolt.File, 0755, &bbolt.Options{Timeout: timeout})
	if err != nil {
		return err
	}

	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("characters"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	}); err != nil {
		return err
	}

	d.db = db
	return nil
}

func (d *bboltDB) Disconnect() error {
	return d.db.Close()
}

func (d *bboltDB) NewCharacter(steamid string, slot int, size int, data string) (uuid.UUID, error) {
	var user *schema.User
	var err error

	charID := uuid.New()
	char := schema.Character{
		ID: charID,
		SteamID: steamid,
		Slot: slot,
		CreatedAt: time.Now(),
		Data: schema.CharacterData{
			CreatedAt: time.Now(),
			Size: size,
			Data: data,
		},
	}

	//Create new user and insert new character.
	if err = d.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("users"))

		data := b.Get([]byte(steamid))
		if len(data) == 0 {
			return ErrNoDocument
		}

		if err := bsoncoder.Decode(data, &user); err != nil {
			return fmt.Errorf("bson: failed to unmarshal %v", err)
		}

		return nil
	}); err == ErrNoDocument {
		if err = d.db.Update(func(tx *bbolt.Tx) error {
			fmt.Println("NEW USER")
			user = &schema.User{
				ID: steamid,
				Characters: make(map[int]uuid.UUID),
			}
			user.Characters[slot] = charID

			userData, err := bsoncoder.Encode(&user)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal user %v", err)
			}

			charData, err := bsoncoder.Encode(&char)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal character %v", err)
			}

			bUser := tx.Bucket([]byte("users"))
			bChar := tx.Bucket([]byte("characters"))

			if err := bUser.Put([]byte(steamid), userData); err != nil {
				return fmt.Errorf("bbolt: failed to put in users", err)
			}

			if err := bChar.Put([]byte(charID.String()), charData); err != nil {
				return fmt.Errorf("bbolt: failed to put in characters", err)
			}
	
			return nil
		}); err != nil {
			return uuid.Nil, err
		}
	} else if err != nil {
		return uuid.Nil, err
	} else {
		if err = d.db.Update(func(tx *bbolt.Tx) error {
			fmt.Println("EXISTING USER")
			user.Characters[slot] = charID
	
			userData, err := bsoncoder.Encode(&user)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal user %v", err)
			}
	
			charData, err := bsoncoder.Encode(&char)
			if err != nil {
				return fmt.Errorf("bson: failed to marshal character %v", err)
			}
	
			bUser := tx.Bucket([]byte("users"))
			bChar := tx.Bucket([]byte("characters"))
	
			if err := bUser.Put([]byte(steamid), userData); err != nil {
				return fmt.Errorf("bbolt: failed to put in users", err)
			}
	
			if err := bChar.Put([]byte(charID.String()), charData); err != nil {
				return fmt.Errorf("bbolt: failed to put in characters", err)
			}
	
			return nil
		}); err != nil {
			return uuid.Nil, err
		}
	}

	return charID, nil
}

func (d *bboltDB) UpdateCharacter(id uuid.UUID, size int, data string, backupMax int, backupTime time.Duration) error {
	return nil
}

func (d *bboltDB) GetUser(steamid string) (user *schema.User, err error) {
	return &schema.User{}, nil
}

func (d *bboltDB) GetCharacter(id uuid.UUID) (char *schema.Character, err error) {
	return &schema.Character{}, nil
}

func (d *bboltDB) GetCharacters(steamid string) (map[int]schema.Character, error) {
	return make(map[int]schema.Character), nil
}

func (d *bboltDB) LookUpCharacterID(steamid string, slot int) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *bboltDB) SoftDeleteCharacter(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) DeleteCharacter(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) DeleteCharacterReference(steamid string, slot int) error {
	return nil
}

func (d *bboltDB) MoveCharacter(id uuid.UUID, steamid string, slot int) error {
	return nil
}

func (d *bboltDB) CopyCharacter(id uuid.UUID, steamid string, slot int) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (d *bboltDB) RestoreCharacter(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) RollbackCharacter(id uuid.UUID, ver int) error {
	return nil
}

func (d *bboltDB) RollbackCharacterToLatest(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) DeleteCharacterVersions(id uuid.UUID) error {
	return nil
}

func (d *bboltDB) SaveToDatabase() error {
	return nil
}

func (d *bboltDB) ClearCache() {
	
}