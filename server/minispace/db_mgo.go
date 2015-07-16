// +build: usemongo

package minispace

import "sync"
import "labix.org/v2/mgo"
import "labix.org/v2/mgo/bson"

type SpaceDB struct {
	session *mgo.Session
	ip      string
	eventch chan *Event
}

type DBEventData struct {
	key  string
	data interface{}
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

func (s *SpaceDB) SyncLoad(name string, data interface{}) error {
	return s.syncEvent(kEventDBLoad, name, data)
}

func (s *SpaceDB) SyncSave(name string, data interface{}) error {
	return s.syncEvent(kEventDBSave, name, data)
}

func (s *SpaceDB) syncEvent(cmd int, name string, data interface{}) error {
	dbe := &DBEventData{
		key:  name,
		data: data,
	}

	var lock sync.Mutex
	var dberr error

	event := &Event{
		cmd:  cmd,
		data: dbe,
		callback: func(e *Event, err error) {
			dberr = err
			lock.Unlock()
		},
	}

	lock.Lock()
	s.AsyncEvent(event)

	// wait
	lock.Lock()

	if dberr != nil {
		return dberr
	}
	return nil
}

func (s *SpaceDB) AsyncEvent(e *Event) {
	s.eventch <- e
}

func (s *SpaceDB) doAsyncLoad(e *Event) {
	dbe := e.data.(*DBEventData)
	err := s.load(dbe.key, dbe.data)
	if e.callback != nil {
		e.callback(e, err)
	}
}

func (s *SpaceDB) doAsyncSave(e *Event) {
	dbe := e.data.(*DBEventData)
	err := s.save(dbe.key, dbe.data)
	if e.callback != nil {
		e.callback(e, err)
	}
}

func (s *SpaceDB) load(name string, data interface{}) error {
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

func (s *SpaceDB) save(name string, data interface{}) error {
	c := s.session.DB("minispace").C("user")
	_, err := c.Upsert(bson.M{"_id": name}, data)

	if err != nil {
		return err
	}

	return nil
}

var sharedSpaceDB *SpaceDB

func NewSpaceDB(ip string) *SpaceDB {
	db := &SpaceDB{
		ip:      ip,
		eventch: make(chan *Event, 64),
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
	go db.dbFlushRoutine()
	return nil
}

func (s *SpaceDB) dbFlushRoutine() {
	var e *Event

	for {
		e = <-s.eventch
		if e == nil {
			continue
		}
		if e.cmd == kEventDBLoad {
			s.doAsyncLoad(e)
		} else if e.cmd == kEventDBSave {
			s.doAsyncSave(e)
		}
	}
}
