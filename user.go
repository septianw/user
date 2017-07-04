package user

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	//	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Model struct {
	Create func(*User) error
	Read   func(string, *User) error
	Update func(*User) error
	Delete func(*User) error
}

type User struct {
	Model
	//	Log  func(string, string) // this package need this to be implemented
	User string
	Hash []byte
	Pass string
}

const (
	SALTLEN = 64
	KEYLEN  = 64
	COST    = 14
)

func New(fc Model) *User {
	user := new(User)

	user.Model.Create = fc.Create
	user.Model.Read = fc.Read
	user.Model.Update = fc.Update
	user.Model.Delete = fc.Delete

	return user
}

func crypt(pass []byte) ([]byte, error) {
	sha := sha512.Sum384(pass)
	shabase := base64.StdEncoding.EncodeToString(sha[:])
	return bcrypt.GenerateFromPassword([]byte(shabase), COST)
}

func compare(hashed []byte, plain []byte) error {
	sha := sha512.Sum384(plain)
	shabase := base64.StdEncoding.EncodeToString(sha[:])
	return bcrypt.CompareHashAndPassword(hashed, []byte(shabase))
}

func (u *User) Remove() error {
	var red = make(chan error, 1)
	var redd error

	if strings.Compare(u.User, "") == 0 {
		fmt.Println("error")
		redd = errors.New("module/user:Remove: User need to retrieved first.")
	} else {
		go func() {
			e := u.Delete(u)
			red <- e
		}()
		redd = <-red
		if redd == nil {
			u.User = ""
			u.Hash = nil
		}
	}

	return redd
}

func (u *User) Modify(pass string, modpass string) error {
	var red = make(chan error, 1)
	var redd error

	if strings.Compare(u.User, "") == 0 {
		fmt.Println("error")
		redd = errors.New("module/user:Modify: User need to retrieved first.")
	} else {
		go func() {
			e := compare(u.Hash, []byte(pass))

			if e == nil {
				u.Hash, e = crypt([]byte(modpass))
				if e == nil {
					e = u.Update(u)
				}
			}
			red <- e
		}()
		redd = <-red
	}

	return redd
}

/*
	Compare password with password in current user object
*/
func (u *User) Compare(pass string) error {
	var red = make(chan error, 1)
	var redd error

	if strings.Compare(u.User, "") == 0 {
		redd = errors.New("module/user:Compare: User need to retrieved first.")
	} else {
		go func() {
			err := compare(u.Hash, []byte(pass))
			red <- err
		}()
		redd = <-red
	}

	return redd
}

/*
	Retrieve user from database and put it in to user object
*/
func (u *User) Retrieve(username string) error {
	var red = make(chan error, 1)
	var redd error

	go func() {
		e := u.Read(username, u)
		red <- e
	}()

	redd = <-red

	return redd
}

/*
	Save user to database, including hash and salt.
	Pre function status:
	  - Current object have plaintext password
	  - Current object still have no hash
	  - Current object still stored in memory
	  - No record written to database
	Post function status:
	  - Current object have no plaintext password anymore
	  - Current object have hash
	  - Current object stored in memory and database
	  - One record written to database
	  - Same username will never written to database
	  - If same username found in database return error
*/
func (u *User) Save() error {
	//	var out bool
	var err error
	var hashed = make(chan []byte, 1)
	var errchan = make(chan error, 1)

	/* Generate salt */
	go func() {
		hash, _ := crypt([]byte(u.Pass))
		hashed <- hash
	}()
	u.Hash = <-hashed
	u.Pass = "" // reset to empty

	/* do low level store process on another thread */
	if u.Hash != nil {
		go func() {
			e := u.Create(u)
			errchan <- e
		}()
	}

	err = <-errchan
	//	if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint") {
	//		err = u.read(u.User, u)
	//	}

	return err
}
