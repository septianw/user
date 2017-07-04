package user

import (
	//	"encoding/base64"
	//	"bytes"
	//	"errors"
	//	"fmt"
	//	"reflect"
	"strings"
	"testing"

	//	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Buser struct {
	gorm.Model
	Username string `gorm:"not null;unique_index"`
	Hash     string
}

const uname = "satu"
const moduname = "dua"
const upass = "ini password"
const modupass = "itu password"

var umod = Model{
	func(us *User) error { // Create
		db, err := connect()
		dbu := new(Buser)
		dbu.Username = us.User
		dbu.Hash = string(us.Hash)

		if err != nil {
			return err
		}

		if !db.HasTable(&Buser{}) {
			db.CreateTable(&Buser{})
		}

		db.NewRecord(dbu)
		return db.Create(&dbu).Error
	},
	func(username string, us *User) error { // Read
		db, err := connect()
		dbu := new(Buser)

		if err == nil {
			err = db.Where("Username = ?", username).First(&dbu).Error
			us.User = dbu.Username
			us.Hash = []byte(dbu.Hash)
		}

		return err
	},
	func(us *User) error { // Update
		db, err := connect()

		if err == nil {
			err = db.Model(&Buser{}).Update("Hash", string(us.Hash)).Error
		}

		return err
	},
	func(us *User) error { // Delete
		db, err := connect()

		if err == nil {
			err = db.Where("Username = ?", us.User).Delete(&Buser{}).Error
		}

		return err
	},
}

func connect() (*gorm.DB, error) {
	return gorm.Open("postgres", "host=localhost user=postgres dbname=usertest sslmode=disable password=root")
}

func TestSave(t *testing.T) {
	var u = New(umod)

	u.User = uname
	u.Pass = upass
	t.Log(u)
	u.Save()
	t.Log(u)
	t.Log(string(u.Hash))

	db, _ := connect()
	dbu := new([]Buser)

	db.Where("Username = ?", u.User).Where("Hash = ?", string(u.Hash)).Find(&dbu)

	src := *dbu // get the real object

	t.Log(string(u.Hash))
	if len(src) != 0 {
		t.Log(string(src[0].Hash))

		if len(src) == 0 {
			t.Error("Fail to store to database")
		} else if (strings.Compare(src[0].Username, u.User) != 0) || (strings.Compare(src[0].Hash, string(u.Hash)) != 0) {
			t.Logf("len dba: %+v", len(src))
			t.Error("Fail to store to database")
		}
	}
}

func TestRetrieve(t *testing.T) {
	var u = New(umod)

	u.Retrieve(uname)
	t.Log(u)

	if strings.Compare(u.User, uname) != 0 {
		t.Errorf("u.User should be %+v instead of %+v", uname, u.User)
	}
	u.Retrieve(moduname)
	t.Log(u)

	if strings.Compare(u.User, "") != 0 {
		t.Error("u.User should be empty instead of %+v", u.User)
	}
}

func TestCompare(t *testing.T) {
	var e error
	var u = New(umod)

	e = u.Compare(upass)
	if e == nil {
		t.Errorf("Compare without retrieving should be error instead of %+v", e)
	}

	u.Retrieve(uname)
	t.Log(string(u.Hash))
	t.Log(upass)

	e = u.Compare(upass)
	if e != nil {
		t.Errorf("Compare should not error instead of %+v", e)
	}
}

func TestModify(t *testing.T) {
	var e error
	var u = New(umod)

	e = u.Modify(upass, modupass)
	if e == nil {
		t.Log(u.User)
		t.Errorf("Modify without retrieving should be error instead of %+v", e)
	}
	t.Logf("error: %+v", e)

	u.Retrieve(uname)
	t.Log(string(u.Hash))
	t.Log(upass)

	e = u.Modify(upass, modupass)

	if e != nil {
		t.Errorf("Modify should not error instead of %+v", e)
	}

	t.Log(string(u.Hash))
	t.Log(modupass)

}

func TestRemove(t *testing.T) {
	var e error
	var u = New(umod)

	e = u.Remove()
	if e == nil {
		t.Log(u.User)
		t.Errorf("Modify without retrieving should be error instead of %+v", e)
	}
	t.Logf("error: %+v", e)

	e = u.Retrieve(uname)
	t.Log(e)
	t.Log(u.User)

	e = u.Remove()
	t.Log(e)

	if e != nil {
		t.Errorf("Modify should not error instead of %+v", e)
	}

	//	u := NewUser()

	//	removed := u.Remove("dua")
	//	t.Log(removed)
	//	if removed {
	//		t.Log(u)
	//		t.Log(removed)
	//		t.Errorf("Try to remove non existent user, return value should false instead of %+v", removed)
	//	}

	//	removed = u.Remove("satu")
	//	got := u.Get("satu")
	//	t.Log(got)
	//	t.Log(removed)

	//	if removed && (got == nil) {
	//		t.Log(u)
	//		t.Errorf("User removed, but still present in current object")
	//	}
}
