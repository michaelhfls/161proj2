package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	"testing"
	"reflect"
	"github.com/cs161-staff/userlib"
	_ "encoding/json"
	_ "encoding/hex"
	_ "github.com/google/uuid"
	_ "strings"
	_ "errors"
	_ "strconv"
)


func TestInit(t *testing.T) {
	t.Log("Initialization test")

	// You may want to turn it off someday
	userlib.SetDebugStatus(true)
	// someUsefulThings()  //  Don't call someUsefulThings() in the autograder in case a student removes it
	userlib.SetDebugStatus(false)
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	if len(userlib.KeystoreGetMap()) < 2 {
		t.Error("Failed to store keys on keyserver")
	}
	// t.Log() only produces output if you run with "go test -v"
	t.Log("Got user", u)
	// If you want to comment the line above,
	// write _ = u here to make the compiler happy
	// You probably want many more tests here.
}

func TestGetUser(t *testing.T) {
	a, err := InitUser("Patricia", "bussy")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	if a.Username != "Patricia" {
		t.Error("received different name: " + a.Username, err)
	}

	t.Log("Patricia popped off!!")

	u, err := GetUser("Patricia", "bussy")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	if !reflect.DeepEqual(u, a) {
		t.Error("The user gotten back does not match!", err)
		return
	}
	t.Log("they match good work!")

	_, err = GetUser("Patricia", "bus")
	if err == nil {
		t.Error("Failed to recognize wrong password", err)
		return
	}

	_, err = GetUser("Patty", "bussy")
	if err == nil {
		t.Error("Failed to recognize wrong password", err)
		return
	}

	for k, _ := range userlib.DatastoreGetMap() {
		userlib.DatastoreSet(k, []byte("fake news"))
	}

	_, err = GetUser("Patty", "bussy")
	if err == nil {
		t.Error("should have detected modification", err)
		return
	}

	userlib.DatastoreClear()

	_, err = GetUser("Patty", "bussy")
	if err == nil {
		t.Error("does not exist", err)
		return
	}

}

func TestStorage(t *testing.T) {
	// And some more tests, because
	a, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	b, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}

	if !reflect.DeepEqual(a, b) {
		t.Error("The user gotten back does not match!", err)
		return
	}


	v := []byte("This is a test")
	a.StoreFile("file1", v)

	c, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}

	if len(c.Files) <= 0 {
		t.Error("Failed to set file on userdata")
		return
	}

	encrypt := userlib.PKEEnc()

	//_, err2 := a.LoadFile("file1")
	//if err2 != nil {
	//	t.Error("Failed to upload and download", err2)
	//	return
	//}
	//if !reflect.DeepEqual(v, v2) {
	//	t.Error("Downloaded file is not the same", v, v2)
	//	return
	//}
}

//func TestShare(t *testing.T) {
//	u, err := GetUser("alice", "fubar")
//	if err != nil {
//		t.Error("Failed to reload user", err)
//		return
//	}
//	u2, err2 := InitUser("bob", "foobar")
//	if err2 != nil {
//		t.Error("Failed to initialize bob", err2)
//		return
//	}
//
//	var v, v2 []byte
//	var magic_string string
//
//	v, err = u.LoadFile("file1")
//	if err != nil {
//		t.Error("Failed to download the file from alice", err)
//		return
//	}
//
//	magic_string, err = u.ShareFile("file1", "bob")
//	if err != nil {
//		t.Error("Failed to share the a file", err)
//		return
//	}
//	err = u2.ReceiveFile("file2", "alice", magic_string)
//	if err != nil {
//		t.Error("Failed to receive the share message", err)
//		return
//	}
//
//	v2, err = u2.LoadFile("file2")
//	if err != nil {
//		t.Error("Failed to download the file after sharing", err)
//		return
//	}
//	if !reflect.DeepEqual(v, v2) {
//		t.Error("Shared file is not the same", v, v2)
//		return
//	}
//}
