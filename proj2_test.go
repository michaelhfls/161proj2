package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	"github.com/cs161-staff/userlib"
	"reflect"
	"testing"
	_ "strconv"
	_ "errors"
	_ "strings"
	_ "github.com/google/uuid"
	_ "encoding/hex"
	_ "encoding/json"

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

//func TestHack3(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//
//	a, _ := InitUser("alice", "fubar")
//
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	for k, actual := range userlib.DatastoreGetMap() {
//		saved := actual
//		userlib.DatastoreSet(k, []byte("corrupt random file"))
//		err := a.AppendFile("file1", []byte("something random"))
//
//		if err == nil {
//			t.Error("should be corrupted")
//		}
//
//		userlib.DatastoreSet(k, saved)
//	}
//}


func TestShare(t *testing.T) {
	userlib.DatastoreClear()
	userlib.KeystoreClear()

	u, _ := InitUser("alice", "fubar")
	u2, _ := InitUser("bob", "foobar")


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
	a.AppendFile("file1", []byte("Patricia is a poopy head!"))
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
	a.AppendFile("file1", []byte("Patricia is a poopy head!"))
	afile,_ := a.LoadFile("file1")
	bfile,_ := b.LoadFile("file1")
	if !reflect.DeepEqual(afile, bfile) {
		t.Error("patricia and alice are not loading the same files", string(afile), string(bfile))
		return
	}
}
//
//
//func TestRecieve(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	a.StoreFile("oop", []byte("sksksk"))
//	b.StoreFile("oop", []byte("ksksks"))
//	magicstring, _ := a.ShareFile("oop", "patricia")
//	error := b.ReceiveFile("oop", "alice", magicstring)
//	if error == nil {
//		t.Error("not allowed to share file with same filename!")
//		return
//	}
//
//
//}
//
//func TestShareFile3(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	c.AppendFile("file1", []byte("Thats not nice!"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if !reflect.DeepEqual(cfile, afile) {
//		t.Error("should be same")
//		return
//	}
//	if !reflect.DeepEqual(cfile, bfile) {
//		t.Error("should be same")
//		return
//	}
//
//}
//
//func TestRevokeFile0(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	a.RevokeFile("file1", "patricia")
//	a.AppendFile("file1", []byte("i removed access"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if reflect.DeepEqual(afile, bfile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(afile, cfile) {
//		t.Error("should not be same")
//		return
//	}
//}
//
//func TestRevokeFile1(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	a.RevokeFile("file1", "patricia")
//	b.AppendFile("file1", []byte("i removed access"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if reflect.DeepEqual(afile, bfile) {
//		t.Error("should not be same")
//		return
//	}
//	if !reflect.DeepEqual(bfile, cfile) {
//		t.Error("should be same")
//		return
//	}
//
//}
//
//func TestRevokeFile8(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	InitUser("patricia", "bussy")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	a.ShareFile("file1", "patricia")
//	error := a.RevokeFile("file1", "patricia")
//	if error == nil {
//		t.Error("cannot revoke user!")
//	}
//
//
//
//
//}
//
//func TestRevokeFile2(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	b.RevokeFile("file1", "gertrude")
//	a.AppendFile("file1", []byte("i removed access"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if !reflect.DeepEqual(afile, bfile) {
//		t.Error("should be same")
//		return
//	}
//	if reflect.DeepEqual(afile, cfile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(bfile, cfile) {
//		t.Error("should not be same")
//		return
//	}
//}
//
//func TestRevokeFile3(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	b.RevokeFile("file1", "gertrude")
//	c.AppendFile("file1", []byte("i removed access"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if reflect.DeepEqual(cfile, bfile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(afile, cfile) {
//		t.Error("should not be same")
//		return
//	}
//	if !reflect.DeepEqual(afile, bfile) {
//		t.Error("should be same")
//		return
//	}
//}
//
//func TestRevokeFile7(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := a.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "alice", magicstring2)
//
//	a.RevokeFile("file1", "gertrude")
//	a.AppendFile("file1", []byte("i removed access"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if reflect.DeepEqual(cfile, bfile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(afile, cfile) {
//		t.Error("should not be same")
//		return
//	}
//	if !reflect.DeepEqual(afile, bfile) {
//		t.Error("should be same")
//		return
//	}
//}
//
//
//
//func TestRevokeFile4(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	//Sharing and receiving magicword
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	magicstring2, _ := b.ShareFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//
//	b.RevokeFile("file1", "gertrude")
//	c.ReceiveFile("file1", "patricia", magicstring2)
//	c.AppendFile("file1", []byte("i removed access"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//
//	if reflect.DeepEqual(cfile, bfile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(afile, cfile) {
//		t.Error("should not be same")
//		return
//	}
//}
//
//
//
//func TestHack4(t *testing.T) {
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
//	d.ReceiveFile("file1", "belinda", magicstring3)
//
//	a.RevokeFile("file1", "patricia")
//	magicstring4, _ := d.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "belinda", magicstring4)
//
//	b.AppendFile("file1", []byte("im still here"))
//
//	afile,_ := a.LoadFile("file1")
//	bfile,_ := b.LoadFile("file1")
//	cfile,_ := c.LoadFile("file1")
//	dfile,_ := d.LoadFile("file1")
//
//	if !reflect.DeepEqual(afile, bfile) {
//		t.Error("should be same")
//		return
//	}
//	if !reflect.DeepEqual(afile, dfile) {
//		t.Error("should be same")
//		return
//	}
//	if !reflect.DeepEqual(dfile, bfile) {
//		t.Error("should be same")
//		return
//	}
//	if reflect.DeepEqual(cfile, dfile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(cfile, afile) {
//		t.Error("should not be same")
//		return
//	}
//	if reflect.DeepEqual(cfile, bfile) {
//		t.Error("should not be same")
//		return
//	}
//
//}
//func TestHack1(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//
//	InitUser("alice", "fubar")
//	for k, _ := range userlib.DatastoreGetMap() {
//		userlib.DatastoreSet(k, []byte("corrupt random file"))
//	}
//	_, error := GetUser("alice","fubar")
//	if error == nil {
//		t.Error("can't get mutated user!")
//		return
//	}
//}
//
//
//func TestHack5(t *testing.T) {
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
//	d.ReceiveFile("file1", "belinda", magicstring3)
//
//	for k, actual := range userlib.DatastoreGetMap() {
//		saved := actual
//		userlib.DatastoreSet(k, []byte("corrupt random file"))
//		_, error := c.LoadFile("file1")
//		if error != nil {
//			t.Error("corrupted again")
//		}
//		userlib.DatastoreSet(k, saved)
//	}
//}
//
//func TestHack6(t *testing.T) {
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
//	d.ReceiveFile("file1", "belinda", magicstring3)
//
//	for k, actual := range userlib.DatastoreGetMap() {
//		saved := actual
//		userlib.DatastoreSet(k, []byte("corrupt random file"))
//		error := a.RevokeFile("file1", "patricia")
//		if error != nil {
//			t.Error("corrupted again")
//		}
//		userlib.DatastoreSet(k, saved)
//	}
//}
//
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
//	d.ReceiveFile("file1", "belinda", magicstring3)
//
//	for k, actual := range userlib.DatastoreGetMap() {
//		saved := actual
//		userlib.DatastoreSet(k, []byte("corrupt random file"))
//		_, error := a.ShareFile("file1", "gertrude")
//		if error != nil {
//			t.Error("corrupted again")
//		}
//		userlib.DatastoreSet(k, saved)
//	}
//}
//func TestHack8(t *testing.T) {
//	userlib.DatastoreClear()
//	userlib.KeystoreClear()
//	a, _ := InitUser("alice", "fubar")
//	b, _ := InitUser("patricia", "bussy")
//	c, _ := InitUser("gertrude", "clampot")
//
//	v := []byte("This is a test. ")
//	a.StoreFile("file1", v)
//
//	magicstring, _ := a.ShareFile("file1", "patricia")
//	b.ReceiveFile("file1", "alice", magicstring)
//
//	error := c.ReceiveFile("file1", "alice", magicstring)
//	if error == nil {
//		t.Error("he doesnt have access!")
//		return
//	}
//	_, error = c.LoadFile("file1")
//	if error == nil {
//		t.Error("he doesnt have access")
//		return
//	}
//	error = c.AppendFile("file1", []byte("he"))
//	if error == nil {
//		t.Error("no access")
//		return
//	}
//
//}