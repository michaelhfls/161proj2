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
	u, err := GetUser("Patricia", "bussy")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	u, err = GetUser("Patricia", "bussy")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	u, err = GetUser("Patricia", "bussy")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	if !reflect.DeepEqual(u, a) {
		t.Error("The user gotten back does not match!", err)
		return
	}

	_, err = GetUser("Patricia", "bus")
	if err == nil {
		t.Error("Failed to recognize wrong password", err)
		return
	}

	_, err = GetUser("Patty", "bussy")
	if err == nil {
		t.Error("Failed to recognize wrong username", err)
		return
	}

	userlib.DatastoreClear()
	userlib.KeystoreClear()

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
	a, _ := InitUser("alice", "fubar")

	file, error := a.LoadFile("gobears")
	if error == nil {
		t.Error("this file doesnt exist and we still loaded it!")
		return
	}
	if file != nil {
		t.Error("returned wasn't nil")
		return
	}

	v := []byte("This is a test")
	a.StoreFile("file1", v)

	v2, _ := a.LoadFile("file1")

	if !reflect.DeepEqual(v, v2) {
		t.Error("Downloaded file is not the same", v, v2)
		return
	}
}

func TestHack2(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	for k, actual := range userlib.DatastoreGetMap() {
		saved := actual
		userlib.DatastoreSet(k, []byte("corrupt random file"))
		_, error := a.LoadFile("file")
		if error == nil {
			t.Error("was supposed to be corrupted")
		}
		userlib.DatastoreSet(k, saved)
	}
}

func TestAppend(t *testing.T) {

	a, err5 := GetUser("alice", "fubar")
	if err5 != nil {
		t.Error("Failed to reload user", err5)
		return
	}

	v := []byte("This is a test. ")
	a.StoreFile("abc", v)
	v0, err0 := a.LoadFile("abc")

	if err0 != nil {
		t.Error("Failed to load file", err0)

	}

	if !reflect.DeepEqual(v, v0) {
		t.Error("Failed to load file", v0)
	}

	// append with current userdata
	add := []byte("Adding information. ")
	err1 := a.AppendFile("abc", add)
	if err1 != nil {
		t.Error("Failed to append file", err1)
		return
	}

	v2, err2 := a.LoadFile("abc")
	if err2 != nil {
		t.Error("Failed to load file", err2)
		return
	}

	v = append(v, add...)
	if !reflect.DeepEqual(v, v2) {
		t.Error("Failed to append file", v2)
	}

	// writes test with append with retrieved userdata using getuser
	b, err3 := GetUser("alice", "fubar")
	if err3 != nil {
		t.Error("Failed to reload user", err3)
		return
	}
	add2 := []byte("Now continuing... ")
	err4 := b.AppendFile("abc", add2)
	if err4 != nil {
		t.Error("Failed to append file", err4)
		return
	}

	v2, err2 = b.LoadFile("abc")
	if err2 != nil {
		t.Error("Failed to load file", err2)
		return
	}

	v = append(v, add2...)
	if !reflect.DeepEqual(v, v2) {
		t.Error("Failed to append file", v, v2)
	}

	c, err3 := GetUser("alice", "fubar")
	if err3 != nil {
		t.Error("Failed to reload user", err3)
		return
	}

	v2, err2 = c.LoadFile("abc")
	if err2 != nil {
		t.Error("Failed to load file", err2)
		return
	}

	if !reflect.DeepEqual(v, v2) {
		t.Error("Failed to append file", v, v2)
	}
}

func TestHack3(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	a, _ := InitUser("alice", "fubar")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	for k, actual := range userlib.DatastoreGetMap() {
		saved := actual
		userlib.DatastoreSet(k, []byte("corrupt random file"))
		err := a.AppendFile("file1", []byte("something random"))

		if err == nil {
			t.Error("should be corrupted")
		}

		userlib.DatastoreSet(k, saved)
	}
}


func TestShare(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	u, _ := InitUser("alice", "fubar")
	u2, _ := InitUser("bob", "foobar")
	u3, _ := InitUser("charles", "barfoo")

	magic_string, err := u.ShareFile("file1", "bob")
	if err == nil {
		t.Error("File doesn't exist")
		return
	}
	if magic_string != "" {
		t.Error("return empty string when sharing nonexistent file")
		return
	}
	u.StoreFile("file1", []byte("please pass"))
	magic_string, err = u.ShareFile("file1", "bob")

	err = u2.ReceiveFile("file1", "bob", magic_string)
	if err == nil {
		t.Error("Failed to check authenticity", err)
		return
	}

	err = u2.ReceiveFile("file1", "alice", "fake string")
	if err == nil {
		t.Error("Failed to check authenticity", err)
		return
	}

	err = u3.ReceiveFile("file1", "alice", magic_string)
	if err == nil {
		t.Error("Failed to ensure only recipient has access", err)
		return
	}

	err = u2.ReceiveFile("file1", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the file", err)
		return
	}

	v2, err := u2.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}

	v, err := u.LoadFile("file1")
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", string(v), " | ", string(v2))
		return
	}

	t.Log("YAY")
}

func TestShareFile0(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")

	//Sharing and receiving magicword
	a.StoreFile("file1", []byte("This is a test. "))
	magicstring, _ := a.ShareFile("file1", "patricia")
	_ = b.ReceiveFile("file1", "alice", magicstring)
	a.StoreFile("file1", []byte("new file"))
	_ = a.AppendFile("file1", []byte("Patricia is a poopy head!"))
	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	if reflect.DeepEqual(afile, bfile) {
		t.Error("They should not share same file i think")
		return
	}
}

func TestShareFile(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")

	//Sharing and receiving magicword
	a.StoreFile("file1", []byte("This is a test. "))
	magicstring, error := a.ShareFile("file1", "patricia")
	if error != nil {
		t.Error("Sharing the magic word flopped")
		return
	}
	error = b.ReceiveFile("file1", "alice", magicstring)
	if error != nil {
		t.Error(error)
	}

	error = a.AppendFile("file1", []byte("Patricia is a poopy head!"))
	if error != nil {
		t.Error(error)
	}

	afile, err := a.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	bfile, err := b.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("patricia and alice are not loading the same files", string(afile), " | ", string(bfile))
		return
	}
}


func TestReceive(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	a.StoreFile("oop", []byte("sksksk"))
	b.StoreFile("oop", []byte("ksksks"))
	magicstring, _ := a.ShareFile("oop", "patricia")
	error := b.ReceiveFile("oop", "alice", magicstring)
	if error == nil {
		t.Error("not allowed to share file with same filename!")
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
	err := b.ReceiveFile("file1", "alice", magicstring)
	if err != nil {
		t.Error(err)
	}

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	err = c.ReceiveFile("file1", "patricia", magicstring2)
	if err != nil {
		t.Error(err)
	}

	err = c.AppendFile("file1", []byte("That's not nice!"))
	if err != nil {
		t.Error(err)
	}

	afile, err := a.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	bfile, err := b.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}
	cfile, err := c.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(cfile, afile) {
		t.Error("should be same", string(cfile), " | ", string(afile))
		return
	}

	if !reflect.DeepEqual(cfile, bfile) {
		t.Error("should be same", string(cfile), " | ", string(bfile))
		return
	}

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("should be same", string(afile), " | ", string(bfile))
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
	err := b.ReceiveFile("file1", "alice", magicstring)
	if err != nil {
		t.Error(err)
	}

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	err = c.ReceiveFile("file1", "patricia", magicstring2)
	if err != nil {
		t.Error(err)
	}

	err = a.RevokeFile("file1", "patricia")
	if err != nil {
		t.Error(err)
	}

	app := "i removed access"
	err = a.AppendFile("file1", []byte(app))
	if err != nil {
		t.Error(err)
	}

	afile, err := a.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(afile, append(v, app...)) {
		t.Error("should be same")
		return
	}

	bfile, err := b.LoadFile("file1")
	if err == nil {
		if reflect.DeepEqual(afile, bfile) {
			t.Error("should not be same")
			return
		}
	}

	cfile, err := c.LoadFile("file1")
	if err == nil {
		if reflect.DeepEqual(afile, cfile) {
			t.Error("should not be same")
			return
		}
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
	err := b.ReceiveFile("file1", "alice", magicstring)
	if err != nil {
		t.Error(err)
	}

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	err = c.ReceiveFile("file1", "patricia", magicstring2)
	if err != nil {
		t.Error(err)
	}

	err = a.RevokeFile("file1", "patricia")
	if err != nil {
		t.Error(err)
	}

	err = b.AppendFile("file1", []byte("i removed access"))
	if err == nil {
		t.Error("should revoke access")
	}

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

func TestRevokeFile8(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	_, _ = InitUser("patricia", "bussy")

	//Sharing and receiving magicword
	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	a.ShareFile("file1", "patricia")
	err := a.RevokeFile("file1", "patricia")
	if err != nil {
		t.Error("cannot revoke user!")
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
	if reflect.DeepEqual(bfile, cfile) {
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
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("should be same")
		return
	}
}

func TestRevokeFile7(t *testing.T) {
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

	magicstring2, _ := a.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "alice", magicstring2)

	a.RevokeFile("file1", "gertrude")
	a.AppendFile("file1", []byte("i removed access"))

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
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("should be same")
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



func TestHack4(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	p, _ := InitUser("patricia", "bussy")
	g, _ := InitUser("gertrude", "clampot")
	b, _ := InitUser("belinda", "mussels")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	err := p.ReceiveFile("file1", "alice", magicstring)
	if err != nil {
		t.Error(err)
	}

	magicstring2, _ := p.ShareFile("file1", "gertrude")
	err = g.ReceiveFile("file1", "patricia", magicstring2)
	if err != nil {
		t.Error(err)
	}

	magicstring3, _ := a.ShareFile("file1", "belinda")
	err = b.ReceiveFile("file1", "alice", magicstring3)
	if err != nil {
		t.Error(err)
	}

	err = a.RevokeFile("file1", "patricia")
	if err != nil {
		t.Error(err)
	}

	_,err = p.LoadFile("file1")
	if err == nil {
		t.Error("should not access file")
	}
	_,err = g.LoadFile("file1")
	if err == nil {
		t.Error("should not access file")
	}

	magicstring4, err := b.ShareFile("file1", "patricia")
	if err != nil {
		t.Error(err)
	}

	err = p.ReceiveFile("file1", "belinda", magicstring4)
	if err == nil {
		t.Error("file should already exist")
	}

	err = p.ReceiveFile("file2", "belinda", magicstring4)
	if err != nil {
		t.Error(err)
	}

	err = p.AppendFile("file2", []byte("im still here"))
	if err != nil {
		t.Error(err)
	}

	afile,err := a.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	pfile,err := p.LoadFile("file2")
	if err != nil {
		t.Error(err)
	}

	gfile,err := g.LoadFile("file1")
	if err == nil {
		t.Error("should not access file")
	}

	bfile,err := b.LoadFile("file1")
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(afile, pfile) {
		t.Error("should be same", string(afile), " | ", string(pfile))
		return
	}
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("should be same", string(afile), " | ", string(pfile))
		return
	}
	if !reflect.DeepEqual(bfile, pfile) {
		t.Error("should be same", string(afile), " | ", string(pfile))
		return
	}
	if reflect.DeepEqual(gfile, bfile) {
		t.Error("should not be same", string(afile), " | ", string(pfile))
		return
	}
	if reflect.DeepEqual(gfile, afile) {
		t.Error("should not be same", string(afile), " | ", string(pfile))
		return
	}
	if reflect.DeepEqual(gfile, pfile) {
		t.Error("should not be same", string(afile), " | ", string(pfile))
		return
	}
}

func TestHack1(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	_, _ = InitUser("alice", "fubar")
	for k, _ := range userlib.DatastoreGetMap() {
		userlib.DatastoreSet(k, []byte("corrupt random file"))
	}
	_, error := GetUser("alice","fubar")
	if error == nil {
		t.Error("can't get mutated user!")
		return
	}
}


func TestHack5(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	p, _ := InitUser("patricia", "bussy")
	g, _ := InitUser("gertrude", "clampot")
	b, _ := InitUser("belinda", "mussels")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	err := p.ReceiveFile("file1", "alice", magicstring)
	if err != nil {
		t.Error(err)
	}

	magicstring2, _ := p.ShareFile("file1", "gertrude")
	err = g.ReceiveFile("file1", "patricia", magicstring2)
	if err != nil {
		t.Error(err)
	}

	magicstring3, _ := a.ShareFile("file1", "belinda")
	err = b.ReceiveFile("file1", "alice", magicstring3)
	if err != nil {
		t.Error(err)
	}

	for k, actual := range userlib.DatastoreGetMap() {
		saved := actual
		userlib.DatastoreSet(k, []byte("corrupt random file"))
		_, err = g.LoadFile("file1")
		if err == nil {
			t.Error("corrupted again")
		}
		userlib.DatastoreSet(k, saved)
	}
}

func TestHack6(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")
	d, _ := InitUser("belinda", "mussels")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	magicstring2, _ := b.ShareFile("file1", "gertrude")
	c.ReceiveFile("file1", "patricia", magicstring2)

	magicstring3, _ := a.ShareFile("file1", "belinda")
	d.ReceiveFile("file1", "alice", magicstring3)

	for k, actual := range userlib.DatastoreGetMap() {
		saved := actual
		userlib.DatastoreSet(k, []byte("corrupt random file"))
		err := a.RevokeFile("file1", "patricia")
		if err != nil {
			t.Error(err)
		}
		userlib.DatastoreSet(k, saved)
	}
}

//func TestHack7(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//	d, _ := InitUser("belinda", "mussels")
//
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	magicstring3, _ := a.ShareFile("file1", "belinda")
//	d.ReceiveFile("file1", "alice", magicstring3)
//
//	for k, actual := range userlib.DatastoreGetMap() {
//		saved := actual
//		userlib.DatastoreSet(k, []byte("corrupt random file"))
//		_, err := a.ShareFile("file1", "gertrude")
//		if err != nil {
//			t.Error("corrupted again")
//		}
//		userlib.DatastoreSet(k, saved)
//	}
//}

func TestHack8(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()
	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("patricia", "bussy")
	c, _ := InitUser("gertrude", "clampot")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, _ := a.ShareFile("file1", "patricia")
	b.ReceiveFile("file1", "alice", magicstring)

	error := c.ReceiveFile("file1", "alice", magicstring)
	if error == nil {
		t.Error("he doesnt have access!")
		return
	}
	_, error = c.LoadFile("file1")
	if error == nil {
		t.Error("he doesnt have access")
		return
	}
	error = c.AppendFile("file1", []byte("he"))
	if error == nil {
		t.Error("no access")
		return
	}
}

func TestAppendMore(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	a, _ := InitUser("alice", "fubar")
	b, _ := InitUser("bob", "bussy")
	c, _ := InitUser("charles", "clampot")
	d, _ := InitUser("darlene", "dahlia")

	v := []byte("This is a test. ")
	a.StoreFile("file1", v)

	magicstring, err := a.ShareFile("file1", "bob")
	if err != nil {
		t.Error(err)
		return
	}

	err = b.ReceiveFile("file1", "alice", magicstring)
	if err != nil {
		t.Error(err)
		return
	}

	add := []byte("Adding more. ")
	v = append(v, add...)
	err = a.AppendFile("file1", add)
	if err != nil {
		t.Error(err)
		return
	}

	// CHECK APPEND HERE
	afile, err := a.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	bfile, err := b.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(afile, v) {
		t.Error("Downloaded file is not the same", v, afile)
		return
	}

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("File is not being shared properly", afile, bfile)
		return
	}

	add = []byte("And some more. ")
	v = append(v, add...)
	err = b.AppendFile("file1", add)
	if err != nil {
		t.Error(err)
		return
	}

	// CHECK ANOTHER APPEND HERE
	a, _ = GetUser("alice", "fubar")
	b, _ = GetUser("bob", "bussy")

	afile, err = a.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	bfile, err = b.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(afile, v) {
		t.Error("Downloaded file is not the same", v, afile)
		return
	}

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("File is not being shared properly", afile, bfile)
		return
	}

	// Add C !
	magicstring, err = a.ShareFile("file1", "charles")
	if err != nil {
		t.Error(err)
		return
	}

	bfile, err = b.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(bfile, v) {
		t.Error("Downloaded file is not the same", string(v), string(bfile))
		return
	}

	err = c.ReceiveFile("filec", "alice", magicstring)
	if err != nil {
		t.Error(err)
		return
	}

	a, _ = GetUser("alice", "fubar")
	b, _ = GetUser("bob", "bussy")
	c, _ = GetUser("charles", "clampot")

	afile, err = a.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	bfile, err = b.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	cfile, err := c.LoadFile("filec")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(afile, v) {
		t.Error("Downloaded file is not the same", v, afile)
		return
	}

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("File is not being shared properly", afile, bfile)
		return
	}

	if !reflect.DeepEqual(afile, cfile) {
		t.Error("File is not being shared properly", afile, cfile)
		return
	}

	add = []byte("Now for some real fun! ")
	v = append(v, add...)
	err = c.AppendFile("filec", add)
	if err != nil {
		t.Error(err)
		return
	}

	add = []byte("Just kidding i'm not as creative as michael uwu ")
	v = append(v, add...)
	err = b.AppendFile("file1", add)
	if err != nil {
		t.Error(err)
		return
	}

	afile, err = a.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	cfile, err = c.LoadFile("filec")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(afile, v) {
		t.Error("Downloaded file is not the same", string(v), string(afile))
		return
	}

	if !reflect.DeepEqual(afile, cfile) {
		t.Error("File is not being shared properly", string(afile),
			" | ", string(bfile))
		return
	}

	err = a.RevokeFile("file1", "charles")
	if err != nil {
		t.Error(err)
		return
	}

	// Add D !
	magicstring, err = b.ShareFile("file1", "darlene")
	if err != nil {
		t.Error(err)
		return
	}

	err = d.ReceiveFile("file1", "bob", magicstring)
	if err != nil {
		t.Error(err)
		return
	}

	add = []byte("so much blood tears and SWEAT (hehe) ")
	v = append(v, add...)
	err = d.AppendFile("file1", add)
	if err != nil {
		t.Error(err)
		return
	}

	afile, err = a.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	bfile, err = b.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	cfile, err = c.LoadFile("filec")
	if err == nil {
		t.Error("NOPE")
	}

	dfile, err := d.LoadFile("file1")
	if err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(afile, v) {
		t.Error("Downloaded file is not the same", v, afile)
		return
	}

	if !reflect.DeepEqual(afile, bfile) {
		t.Error("File is not being shared properly", afile, bfile)
		return
	}

	if !reflect.DeepEqual(bfile, dfile) {
		t.Error("File is not being shared properly", bfile, dfile)
		return
	}

	if reflect.DeepEqual(dfile, cfile) {
		t.Error("File is not being shared properly", dfile, cfile)
		return
	}
}