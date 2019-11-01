package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	"github.com/cs161-staff/userlib"
	_ "github.com/google/uuid"
	"reflect"
	_ "strconv"
	_ "strings"
	"testing"
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
	userlib.DatastoreClear()
	userlib.KeystoreClear()
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

	_, error := a.LoadFile("gobears")
	if error == nil {
		t.Error("this file doesnt exist and we still loaded it!")
		return
	}

	v := []byte("This is a test")
	a.StoreFile("file1", v)

	v2, err2 := a.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}

	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

	// More tests
	c, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}

	if !reflect.DeepEqual(a.Files, c.Files) {
		t.Error("Did not update userdata", a.Files, c.Files)
		return
	}

	v3, err3 := c.LoadFile("file1")

	if err3 != nil {
		t.Error("Failed to upload and download", err3)
		return
	}


	if !reflect.DeepEqual(v, v3) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}

	_, err4 := c.LoadFile("file2")
	if err4 == nil {
		t.Error("Failed to recognize nonexistent file")
		return
	}

	f := []byte("This is my second file")
	c.StoreFile("file2", f)

	d, err5 := GetUser("alice", "fubar")
	if err5 != nil {
		t.Error("Failed to reload user", err5)
		return
	}

	f1, err6 := d.LoadFile("file2")
	if err6 != nil {
		t.Error("Failed to retrieve file", err6)
		return
	}

	if !reflect.DeepEqual(f, f1) {
		t.Error("Failed to retrieve file")
		return
	}

	w := []byte("This is my third file")

	d.StoreFile("file2", w)
	w1, err7 := d.LoadFile("file2")
	if err7 != nil {
		t.Error("Failed to load file", err7)
		return
	}

	if !reflect.DeepEqual(w, w1) {
		t.Error("Failed to retrieve file", w, w1)
		return
	}

	if reflect.DeepEqual(w, f) {
		t.Error("Failed to overwrite file")
		return
	}
}

func TestAppend(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)
	a.LoadFile(("file1"))

	// append with current userdata
	add := []byte("Adding information. ")
	err1 := a.AppendFile("file1", add)
	if err1 != nil {
		t.Error("Failed to append file", err1)
		return
	}


	v2, _ := a.LoadFile("file1")
	if !reflect.DeepEqual(append(v, add...), v2) {
		t.Error("Failed to append file", v2)
	}
}


func TestShareFile(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)
	_, error := a.ShareFile("flies1", "patricia")
	if error != nil {
		t.Error("share file that doesnt exist")
		return
	}
	magicstring, error := a.ShareFile("file1", "patricia")
	if error != nil {
		t.Error("Sharing the magic word flopped")
		return
	}
	error = b.ReceiveFile("file1", "alice", magicstring)
	if error != nil {
		t.Error("something flopped with receiving the magic word")
		return
	}
	a.AppendFile("file1", []byte("Patricia is a poopy head!"))
	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("patricia and alice are not loading the same files")
		return
	}
}


func TestShareFile2(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	InitUser("patricia", "bussy")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)
	_, error := a.ShareFile("flies1", "patricia")
	if error == nil {
		t.Error("should've errored out")
		return
	}
}

func TestShareFile3(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	c.AppendFile("file1", []byte("Thats not nice!"))

	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	cfile,_ := c.LoadFile("file1")

	if !reflect.DeepEqual(cfile, afile) {
		t.Error("should be same")
		return
	}
	if !reflect.DeepEqual(cfile, bfile) {
		t.Error("should be same")
		return
	}

}

func TestRevokeFile0(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	a.RevokeFile("file1", "patricia")
	a.AppendFile("file1", []byte("i removed access"))

	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	cfile,_ := c.LoadFile("file1")

	if reflect.DeepEqual(afile, bfile) {
		t.Error("should not be same")
		return
	}
	if reflect.DeepEqual(afile, cfile) {
		t.Error("should not be same")
		return
	}
}

func TestRevokeFile1(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	a.RevokeFile("file1", "patricia")
	b.AppendFile("file1", []byte("i removed access"))

	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	cfile,_ := c.LoadFile("file1")

	if reflect.DeepEqual(afile, bfile) {
		t.Error("should not be same")
		return
	}
	if !reflect.DeepEqual(bfile, cfile) {
		t.Error("should be same")
		return
	}

}

func TestRevokeFile2(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	b.RevokeFile("file1", "gertrude")
	a.AppendFile("file1", []byte("i removed access"))

	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	cfile,_ := c.LoadFile("file1")

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("should be same")
		return
	}
	if reflect.DeepEqual(afile, cfile) {
		t.Error("should not be same")
		return
	}
}

func TestRevokeFile3(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	b.RevokeFile("file1", "gertrude")
	c.AppendFile("file1", []byte("i removed access"))

	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	cfile,_ := c.LoadFile("file1")

	if reflect.DeepEqual(cfile, bfile) {
		t.Error("should not be same")
		return
	}
	if reflect.DeepEqual(afile, cfile) {
		t.Error("should not be same")
		return
	}
}

func TestRevokeFile4(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	b.RevokeFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)
	c.AppendFile("file1", []byte("i removed access"))

	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	cfile,_ := c.LoadFile("file1")

	if reflect.DeepEqual(cfile, bfile) {
		t.Error("should not be same")
		return
	}
	if reflect.DeepEqual(afile, cfile) {
		t.Error("should not be same")
		return
	}
}


func TestShare(t *testing.T) {
	u, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}

	var v, v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice", err)
		return
	}

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
}
