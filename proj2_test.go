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
	a, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)
	v0, err0 := a.LoadFile(("file1"))

	if err0 != nil {
		t.Error("Failed to load file", err0)

	}

	if !reflect.DeepEqual(v, v0) {
		t.Error("File doesn't match what we put in", v0)
	}

	// append with current userdata
	add := []byte("Adding information. ")
	t.Log("start")
	err1 := a.AppendFile("file1", add)
	if err1 != nil {
		t.Error("Failed to append file", err1)
		return
	}


	v2, err2 := a.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to load file", err2)
		return
	}

	if !reflect.DeepEqual(append(v, add...), v2) {
		t.Error("Failed to append file", v2)
	}



	// writes test swith append with retrieved userdata using getuser
}


func TestRevokeFile(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	b, err := InitUser("patricia", "bussy")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	c, err := InitUser("gertrude", "clampot")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	d, err := InitUser("ryshandala", "paninipress")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}



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
	error = c.ReceiveFile("file1", "alice", magicstring)
	if error != nil {
		t.Error("something flopped with receiving the magic word")
		return
	}
	magicstring2, error := b.ShareFile("file1", "ryshandala")
	if error != nil {
		t.Error("Sharing the magic word flopped")
		return
	}
	error = d.ReceiveFile("file1", "patricia", magicstring2)
	if error != nil {
		t.Error("something flopped with receiving the magic word")
		return
	}

	x := []byte("Eurekaaa!!!")
	b.StoreFile("file1", x)
	stuff, error := b.LoadFile("file1")
	if error == nil {
		t.Log(string(stuff))
		t.Error("should error out but didnt")
		return
	}


	//check load access
	afile, error := a.LoadFile("file1")
	if error != nil {
		t.Error("Alice flopped loading the file after sharing with patricia")
		return
	}
	bfile, error := b.LoadFile("file1")
	if error != nil {
		t.Error("Patricia could not load the file that was shared with her")
		return
	}
	cfile, error := c.LoadFile("file1")
	if error != nil {
		t.Error("Gertrude flopped loading the file after sharing with patricia")
		return
	}
	dfile, error := d.LoadFile("file1")
	if error != nil {
		t.Error("Ryshandala could not load the file that was shared with her")
		return
	}
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("patricia and alice are not loading the same files")
		return
	}
	if !reflect.DeepEqual(afile, cfile) {
		t.Error("alice and gertrude are not loading the same files")
		return
	}
	if !reflect.DeepEqual(afile, dfile) {
		t.Error("alice and ryshandala are not loading the same files")
		return
	}


	error = a.AppendFile("file1", []byte("Patricia is a poopy head!"))
	if error != nil {
		t.Error("Alice couldn't append to her own file after sharing it!")
		return
	}
	afile, error = a.LoadFile("file1")
	if error != nil {
		t.Error("Alice flopped loading the file after sharing with patricia and appending")
		return
	}
	bfile, error = b.LoadFile("file1")
	if error != nil {
		t.Error("Patricia could not load the file that was shared with her after alice appended")
		return
	}
	cfile, error = c.LoadFile("file1")
	if error != nil {
		t.Error("Gertrude could not load the file that was shared with her after alice appended")
		return
	}
	dfile, error = d.LoadFile("file1")
	if error != nil {
		t.Error("Ryshandala could not load the file that was shared with her after alice appended")
		return
	}
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("patricia and alice are not loading the same files")
		return
	}
	if !reflect.DeepEqual(afile, cfile) {
		t.Error("patricia and gertrude are not loading the same files")
		return
	}
	if !reflect.DeepEqual(afile, dfile) {
		t.Error("patricia and ryshandala are not loading the same files")
		return
	}


	error = d.AppendFile("file1", []byte("That was rude of you Alice!"))
	if error != nil {
		t.Error("Ryshandala couldn't append to the shared file")
		return
	}
	afile, error = a.LoadFile("file1")
	if error != nil {
		t.Error("Alice flopped loading the file")
		return
	}
	bfile, error = b.LoadFile("file1")
	if error != nil {
		t.Error("Patricia could not load the file")
		return
	}
	cfile, error = c.LoadFile("file1")
	if error != nil {
		t.Error("Gertrude could not load the file")
		return
	}
	dfile, error = d.LoadFile("file1")
	if error != nil {
		t.Error("Ryshandala could not load the file")
		return
	}
	if !reflect.DeepEqual(dfile, afile) {
		t.Error("ryshandala and alice are not loading the same files")
		return
	}
	if !reflect.DeepEqual(dfile, bfile) {
		t.Error("ryshandala and patricia are not loading the same files")
		return
	}
	if !reflect.DeepEqual(dfile, cfile) {
		t.Error("ryshandala and gertrude are not loading the same files")
		return
	}


	//revoking files
	error = a.RevokeFile("file1", "patricia")
	if error != nil {
		t.Error("Alice couldnt revoke patricia!")
		return
	}

	error = a.AppendFile("file1", []byte("patricia was toxic"))
	if error != nil {
		t.Error("Alice couldnt append after revoking patricia")
		return
	}

	afile, error = a.LoadFile("file1")
	if error != nil {
		t.Error("Alice flopped")
		return
	}
	bfile, error = b.LoadFile("file1")
	if error != nil {
		t.Error("Patricia no")
		return
	}
	cfile, error = c.LoadFile("file1")
	if error != nil {
		t.Error("Gertrude flopped")
		return
	}
	dfile, error = d.LoadFile("file1")
	if error != nil {
		t.Error("Ryshandala could not")
		return
	}
	if reflect.DeepEqual(afile, bfile) {
		t.Error("Alice revoked access but Patricia was still updated")
		return
	}
	if !reflect.DeepEqual(afile, cfile) {
		t.Error("Alice should still be shared with Gertrude!")
		return
	}
	error = d.AppendFile("file1", []byte("im sad"))
	if error != nil {
		t.Error("whatever doesnt matter")
		return
	}
	afile, error = a.LoadFile("file1")
	if error != nil {
		t.Error("Oh no alice")
		return
	}
	dfile,error = d.LoadFile("file1")
	if error != nil {
		t.Error("o no ryshandala")
		return
	}
	if reflect.DeepEqual(afile, dfile) {
		t.Error("Alice and ryshandala shouldnt be connected")
		return
	}
	error = c.AppendFile("file1", []byte("morestuff"))
	if error != nil {
		t.Error("patricia couldnt append anymore")
	}
	afile, error = a.LoadFile("file1")
	if error != nil {
		t.Error("Alice flopped")
		return
	}
	cfile, error = c.LoadFile("file1")
	if error != nil {
		t.Error("Patricia no")
		return
	}
	if !reflect.DeepEqual(afile, cfile) {
		t.Error("should be same")
		return
	}




	//Get access again??

	error = b.ReceiveFile("file1", "Alice", magicstring)
	if error != nil {
		t.Error("Patricia shouldnt be able to reuse magic string to regain access")
		return
	}
	error = b.AppendFile("file1", []byte("screw u alice!"))
	if error != nil {
		t.Error("append failed")
		return
	}
	afile, error = a.LoadFile("file1")
	if error != nil {
		t.Error("s")
		return
	}
	bfile, error = b.LoadFile("file1")
	if error != nil {
		t.Error("s")
		return
	}
	if reflect.DeepEqual(afile, bfile) {
		t.Error("patricia managed to regain access")
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
