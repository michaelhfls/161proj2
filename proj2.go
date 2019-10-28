package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	"fmt"
	// You neet to add with
	// go get github.com/cs161-staff/userlib
	"github.com/cs161-staff/userlib"

	// Life is much easier with json:  You are
	// going to want to use this so you can easily
	// turn complex structures into strings etc...
	"encoding/json"

	// Likewise useful for debugging etc
	"encoding/hex"

	// UUIDs are generated right based on the crypto RNG
	// so lets make life easier and use those too...
	//
	// You need to add with "go get github.com/google/uuid"
	"github.com/google/uuid"

	// Useful for debug messages, or string manipulation for datastore keys
	"strings"

	// Want to import errors
	"errors"

	// optional
	_ "strconv"

	// if you are looking for fmt, we don't give you fmt, but you can use userlib.DebugMsg
	// see someUsefulThings() below
)

// This serves two purposes: It shows you some useful primitives and
// it suppresses warnings for items not being imported
func someUsefulThings() {
	// Creates a random UUID
	f := uuid.New()
	userlib.DebugMsg("UUID as string:%v", f.String())

	// Example of writing over a byte of f
	f[0] = 10
	userlib.DebugMsg("UUID as string:%v", f.String())

	// takes a sequence of bytes and renders as hex
	h := hex.EncodeToString([]byte("fubar"))
	userlib.DebugMsg("The hex: %v", h)

	// Marshals data into a JSON representation
	// Will actually work with go structures as well
	d, _ := json.Marshal(f)
	userlib.DebugMsg("The json data: %v", string(d))
	var g uuid.UUID
	json.Unmarshal(d, &g)
	userlib.DebugMsg("Unmashaled data %v", g.String())

	// This creates an error type
	userlib.DebugMsg("Creation of error %v", errors.New(strings.ToTitle("This is an error")))

	// And a random RSA key.  In this case, ignoring the error
	// return value
	var pk userlib.PKEEncKey
        var sk userlib.PKEDecKey
	pk, sk, _ = userlib.PKEKeyGen()
	userlib.DebugMsg("Key is %v, %v", pk, sk)
}

// Helper function: Takes the first 16 bytes and
// converts it into the UUID type
func bytesToUUID(data []byte) (ret uuid.UUID) {
	for x := range ret {
		ret[x] = data[x]
	}
	return
}

// START CODE HERE

// The structure definition for a user record. Sign this every time it is uploaded.
type User struct {
	Username string
	Files map[string]string // Dictionary with key = encrypted hashed file names, value = encrypted UUID of File-User Node
	DecKey userlib.PKEDecKey
	SignKey userlib.DSSignKey

	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

//init func for user object
func NewUser(username string) (*User, error) {
	var u User
	u.Username = username
	u.Files = make(map[string]string)

	// TODO: Set encryption and signature keys, upload onto Keystore
	encKey, decKey, err := userlib.PKEKeyGen()
	if err == nil {
		signKey, verifyKey, err := userlib.DSKeyGen()
		if err == nil {
			err := userlib.KeystoreSet(u.Username + "enc", encKey)
			if err == nil {
				err := userlib.KeystoreSet(u.Username + "sign", verifyKey)
				if err == nil {
					u.DecKey = decKey
					u.SignKey = signKey
				}
			}
		}
	}
	return &u, err
}

// I created extra blob structure so we have a way to store encrypted userdata and hmac
type Blob struct {
	EncryptUD []byte
	Hmac []byte

	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

//init func for blob object
func NewBlob(EncryptUD []byte, Hmac []byte) *Blob{
	var b Blob
	b.EncryptUD = EncryptUD
	b.Hmac = Hmac
	return &b
}

// When initializing a user, this function is called to upload
// the user's public keys onto Keystore.
func InitKeys() {

}

// Call this function to retrieve the user's deterministic keys in a
// stateless manner. We should have one of these for each specific key we need!
func RetrieveKeys(username string, password string) {

}

// Retrieve the user's private encryption key.
func GetPrivEncrKey() {

}

// Retrieve the user's private signature key.
func GetPrivSigKey() {

}

// Retrieve the public encryption key in Keystore under the name username.
func GetPublicEncrKey(username string) {

}

// Retrieve the public signature key in Keystore under the name username.
func GetPublicSigKey(username string) {

}

// Retrieve the UUID associated with the userdata.
func GetUserUUID (username string, password string) {

}

//creates three symmetric keys
func HKDF(key []byte, msg []byte) ([]byte, []byte, []byte) {
	hmac, _ := userlib.HMACEval(key, msg)
	return hmac[0:16], hmac[16:32], hmac[32:48]
}

// This creates a user.  It will only be called once for a user
// (unless the keystore and datastore are cleared during testing purposes)

// It should store a copy of the userdata, suitably encrypted, in the
// datastore and should store the user's public key in the keystore.

// The datastore may corrupt or completely erase the stored
// information, but nobody outside should be able to get at the stored
// User data: the name used in the datastore should not be guessable
// without also knowing the password and username.

// You are not allowed to use any global storage other than the
// keystore and the datastore functions in the userlib library.

// You can assume the user has a STRONG password
func InitUser(username string, password string) (userdataptr *User, err error) {
	// Create deterministic keys and UUID for the userdata
	hash := userlib.Argon2Key([]byte(password), []byte(username), 16)
	uuidKey, encrKey, hmacKey := HKDF(hash, []byte(username + password))
	uuidUD := bytesToUUID(uuidKey)
	ud, err := NewUser(username)

	// Error check
	if err != nil {
		return nil, err
	}

	// Serialize, encrypt, and HMAC the userdata
	serialUD, err1 := json.Marshal(ud)
	encryptUD := userlib.SymEnc(encrKey, userlib.RandomBytes(16), serialUD)
	hmac, err2 := userlib.HMACEval(hmacKey, encryptUD)

	// Error check
	if err1 != nil || err2 != nil {
		return nil, err
	}

	// Serialize blob and upload to Datastore
	blob := NewBlob(encryptUD, hmac)
	serialBlob, err := json.Marshal(blob)
	fmt.Println(serialBlob)
	userlib.DatastoreSet(uuidUD, serialBlob)

	// Error check
	if err != nil {
		return nil, err
	}

	return nil, err
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	hash := userlib.Argon2Key([]byte(password), []byte(username), 16)
	key1, key2, key3 := HKDF(hash, []byte(username + password))
	uuid := bytesToUUID(key1)

	serial_blob, boolean := userlib.DatastoreGet(uuid)
	if boolean == false {
		error := errors.New("user cannot be found")
		return nil, error
	}
	var blob Blob
	json.Unmarshal(serial_blob, &blob)
	fmt.Println(blob)
	hmaccheck, _ := userlib.HMACEval(key3, blob.EncryptUD)
	if userlib.HMACEqual(blob.Hmac, hmaccheck) == false {
		error := errors.New("user file was corrupted")
		return nil, error
	}

	decrypt_ud := userlib.SymDec(key2, blob.EncryptUD)
	var ud *User
	json.Unmarshal(decrypt_ud, ud)
	return ud, nil
}

// This stores a file in the datastore.
//
// The name and length of the file should NOT be revealed to the datastore!
func (userdata *User) StoreFile(filename string, data []byte) {
	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.

func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	return
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	return
}

// This creates a sharing record, which is a key pointing to something
// in the datastore to share with the recipient.

// This enables the recipient to access the encrypted file as well
// for reading/appending.

// Note that neither the recipient NOR the datastore should gain any
// information about what the sender calls the file.  Only the
// recipient can access the sharing record, and only the recipient
// should be able to know the sender.

func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {

	return
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	return nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	return
}
