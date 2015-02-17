package mongod

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/nowk/assert.v2"
	"testing"
)

const tableName = "users"

type User struct {
	Name string
}

var clean = func(t *testing.T) func(*mgo.Database) {
	return func(db *mgo.Database) {
		_, err := db.C(tableName).RemoveAll(bson.M{})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestCloseCallback(t *testing.T) {
	m := New("mongod_test")
	db, err := m.Start()
	if err != nil {
		t.Fatal(err)
	}

	if err := db.C(tableName).Insert(&User{
		Name: "Batman",
	}); err != nil {
		t.Fatal(err)
	}
	m.Stop(clean(t))

	db, err = m.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer m.Stop()

	n, err := db.C(tableName).Count()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, n)
}
