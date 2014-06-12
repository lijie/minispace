// +build: usemongo

package minispace

import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

type SpaceDB struct {
	session *mgo.Session
	ip string
}

func (s *SpaceDB) Connect() error {
	if s.session != nil {
		return nil
	}

	// TODO: read ip from config file
	var err error
	s.session, err = mgo.Dial(s.ip)
	if err != nil {
		return err
	}
	s.session.SetMode(mgo.Monotonic, true)
	return nil
}

func (s *SpaceDB) Load(name string, data interface{}) error {
	c := s.session.DB("minispace").C("user")
	err := c.Find(bson.M{"_id": name}).One(data)

	if err == mgo.ErrNotFound {
		return ErrUserNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (s *SpaceDB) Save(name string, data interface{}) error {
	c := s.session.DB("minispace").C("user")
	_, err := c.Upsert(bson.M{"_id":name}, data)

	if err != nil {
		return err
	}

	return nil
}

var sharedSpaceDB *SpaceDB

func NewSpaceDB(ip string) (*SpaceDB) {
	db := &SpaceDB{
		ip: ip,
	}
	return db
}

func SharedDB() *SpaceDB {
	return sharedSpaceDB
}

func ConnectSharedDB(ip string) error {
	db := NewSpaceDB(ip)
	if err := db.Connect(); err != nil {
		return err
	}
	sharedSpaceDB = db
	return nil
}
