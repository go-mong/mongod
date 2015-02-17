package mongod

import (
	"gopkg.in/mgo.v2"
)

type Mongod interface {
	Clone() (*mgo.Session, *mgo.Database)
	Session() *mgo.Session
	Start() (*mgo.Database, error)
	Stop(...func(*mgo.Database))
}

type Config struct {
	Addr     string
	Database string
}

// mongod provides a way to start/stop an mgo.Session
type mongod struct {
	*Config

	// session is the original dial session
	session *mgo.Session

	// db is the database saved to run stop callbacks on
	db *mgo.Database

	// clones are a collection of cloned database, this does not include the above
	// database
	clones []*mgo.Session
}

var _ Mongod = &mongod{}

func New(name string, opts ...func(c *Config)) *mongod {
	m := &mongod{
		Config: &Config{
			Addr:     "127.0.0.1:27017",
			Database: name,
		},
	}
	for _, v := range opts {
		v(m.Config)
	}
	return m
}

// Session returns the original mgo.Sesssion
func (m mongod) Session() *mgo.Session {
	return m.session
}

// Clone creates clones from the original session appending them to clones to be
// closed on Stop
func (m *mongod) Clone() (*mgo.Session, *mgo.Database) {
	c := m.session.Clone()
	m.clones = append(m.clones, c)
	return c, c.DB(m.Database)
}

// start dials and assigns the original session and database instances
func (m *mongod) start() error {
	ses, err := mgo.Dial(m.Addr)
	if err != nil {
		return err
	}
	m.session = ses
	m.db = ses.DB(m.Database)
	return nil
}

func (m *mongod) Start() (*mgo.Database, error) {
	err := m.start()
	if err != nil {
		return nil, err
	}
	_, db := m.Clone()
	return db, nil
}

// Stop closes the main session and all created clones. It also takes callback
// funcs which are passed an instance of the database
func (m *mongod) Stop(fn ...func(*mgo.Database)) {
	defer m.session.Close()
	defer func() {
		for _, v := range m.clones {
			v.Close()
		}

		// clean out clones
		m.clones = make([]*mgo.Session, 0, 1)
	}()

	for _, v := range fn {
		v(m.db)
	}
}
