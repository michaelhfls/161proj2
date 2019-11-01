package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	// You need to add with
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
	_ = json.Unmarshal(d, &g)
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

// The structure definition for a user record. HMAC this every time it is uploaded.
type User struct {
	Username string
	Files map[uuid.UUID]uuid.UUID // Dictionary with key = encrypted hashed file names, value = encrypted UUID of File-User Node
	DecKey userlib.PKEDecKey // User's private key (RSA)
	SignKey userlib.DSSignKey // User's private key (Digital Signatures)

	// Info to re-upload userdata :( we should think of another design tbh
	UUID uuid.UUID
	EncKey []byte
	HMACKey []byte

	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

//init func for user object
func NewUser(username string) (*User, error) {
	var u User
	u.Username = username
	u.Files = make(map[uuid.UUID]uuid.UUID)

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

// Blob structure that stores userdata or files, and their HMAC or digital signature.
type Blob struct {
	Data []byte
	Check []byte

	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

//init func for blob object
func NewBlob(data []byte, check []byte) *Blob{
	var b Blob
	b.Data = data
	b.Check = check
	return &b
}

func (blob *Blob) VerifyHMAC(key []byte) (verify bool, err error) {
	hmac, err := userlib.HMACEval(key, blob.Data)
	if err != nil {
		return false, err
	}

	if !userlib.HMACEqual(blob.Check, hmac) {
		err := errors.New("file was corrupted")
		return true, err
	}
	return true, err
}

// The structure definition for a user file node. Not encrypted.
type UserFile struct {
	Username string
	UUID uuid.UUID

	Children map[string]uuid.UUID // list of descendents with access to file. username - uuid
	ChildrenDS []byte // digital signature of children

	Parent uuid.UUID // uuid of parent that gave this user access. uuid.NIL if no parent
	ParentDS []byte // digital signature by the parent. NIL if no parent

	SavedMeta map[int][4][]byte // e(username), e(uuid), e(key), ds
	SavedMetaDS [2][]byte // first element is username, second element is DS of user
	ChangesMeta map[int][4][]byte

	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
}

func (userdata *User) NewUserFile(parent uuid.UUID, parentDS []byte, savedMeta map[int][4][]byte, savedMetaDS [2][]byte, changesMeta map[int][4][]byte) *UserFile{
	var f UserFile
	f.Username = userdata.Username
	f.Children = make(map[string]uuid.UUID)
	f.Parent = parent
	f.ParentDS = parentDS
	f.SavedMeta = savedMeta
	f.SavedMetaDS = savedMetaDS
	f.ChangesMeta = changesMeta

	// todo: need to write childrenDS and parentDS in new user file or something idk. do when sharing access???

	return &f
}

func RetrieveUserFile(uuidUF uuid.UUID) (*UserFile, error) {
	serialUserFile, boolean := userlib.DatastoreGet(uuidUF)

	if !boolean {
		err := errors.New("file cannot be found")
		return nil, err
	}

	var userFile UserFile
	_ = json.Unmarshal(serialUserFile, &userFile)

	return &userFile, nil
}

// Updates the metadata of a recipient's changes map, with a file update by sender
func (userFile *UserFile) UpdateMetadata(recipient string, sender *User, uuidFile uuid.UUID, encKey []byte) {
	pubKey, _ := GetPublicEncKey(recipient)

	eUsername, _ := userlib.PKEEnc(pubKey, []byte(sender.Username))
	eUUID, _ := userlib.PKEEnc(pubKey, []byte(uuidFile.String()))
	eKey, _ := userlib.PKEEnc(pubKey, encKey)

	msg := string(len(userFile.ChangesMeta)) + string(eUUID) + string(eKey)
	ds, _ := userlib.DSSign(sender.SignKey, []byte(msg))

	userFile.ChangesMeta[len(userFile.ChangesMeta)] = [4][]byte{eUsername, eUUID, eKey, ds}

	serialUF, _ := json.Marshal(userFile)
	userlib.DatastoreSet(userFile.UUID, serialUF)
}
// Calls UpdateMetadata on this userfile and all of its children and its children...
func (userFile *UserFile) UpdateAllMetadata(sender *User, uuidFile uuid.UUID, encKey []byte) {
	userFile.UpdateMetadata(userFile.Username, sender, uuidFile, encKey)
	if len(userFile.Children) > 0 { // TODO: add another bool later to verify DS when we share access
		for _, uuidChild := range userFile.Children {
			userFile, err := RetrieveUserFile(uuidChild)
			if err == nil {
				//userFile.UpdateMetadata(name, sender, uuidFile, encKey)
				userFile.UpdateAllMetadata(sender, uuidFile, encKey)
			}
		}
	}
}

// Call this function to retrieve the user's deterministic keys in a
// stateless manner. We should have one of these for each specific key we need!
func RetrieveKeys(username string, password string) {

}

// Decrypts ciphertext using the user's private decryption key.
func (userdata *User) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	plaintext, err = userlib.PKEDec(userdata.DecKey, ciphertext)
	return
}

// Retrieve the user's private signature key.
func GetPrivSigKey() {

}

// Retrieve the public encryption key in Keystore under the name username.
func GetPublicEncKey(username string) (userlib.PKEEncKey, bool) {
	return userlib.KeystoreGet(username + "enc")
}

// Retrieve the public signature key in Keystore under the name username.
func GetPublicVerKey(username string) (userlib.DSVerifyKey, bool) {
	return userlib.KeystoreGet(username + "sign")
}

// Retrieve the UUID associated with the userdata.
func GetUserUUID (username string, password string) {

}

//creates three symmetric keys
func UserHKDF(username string, password string) ([]byte, []byte, []byte) {
	hash := userlib.Argon2Key([]byte(password), []byte(username), 16)
	hmac, _ := userlib.HMACEval(hash, []byte(username + password))
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
	uuidKey, encKey, hmacKey := UserHKDF(username, password)
	uuidUD := bytesToUUID(uuidKey)
	ud, err := NewUser(username)
	// Error check
	if err != nil {
		return nil, err
	}
	ud.UUID = uuidUD
	ud.EncKey = encKey
	ud.HMACKey = hmacKey
	//// Serialize, encrypt, and HMAC the userdata
	//serialUD, err1 := json.Marshal(ud)
	//encryptUD := userlib.SymEnc(encKey, userlib.RandomBytes(16), serialUD)
	//hmac, err2 := userlib.HMACEval(hmacKey, encryptUD)
	//
	//// Error check
	//if err1 != nil {
	//	return nil, err1
	//} else if err2 != nil {
	//	return nil, err2
	//}
	//
	//// Serialize blob and upload to Datastore
	//blob := NewBlob(encryptUD, hmac)
	//serialBlob, err := json.Marshal(blob)
	//userlib.DatastoreSet(uuidUD, serialBlob)

	// Upload userdata
	var blob Blob
	err = ud.UploadUser(&blob)

	// Error check
	if err != nil {
		return nil, err
	}
	return ud, err
}

// Uploads the userdata. Use this to also update userdata.
func (userdata *User) UploadUser(blob *Blob) error {
	// Serialize, encrypt, and HMAC the userdata
	serialUD, err1 := json.Marshal(userdata)
	encryptUD := userlib.SymEnc(userdata.EncKey, userlib.RandomBytes(16), serialUD)
	hmac, err2 := userlib.HMACEval(userdata.HMACKey, encryptUD)

	// Error check
	if err1 != nil {
		return err1
	} else if err2 != nil {
		return err2
	}

	// Serialize blob and upload to Datastore
	blob.Data = encryptUD
	blob.Check = hmac
	serialBlob, err := json.Marshal(blob)
	userlib.DatastoreSet(userdata.UUID, serialBlob)

	return err
}

// This fetches the user information from the Datastore.  It should
// fail with an error if the user/password is invalid, or if the user
// data was corrupted, or if the user can't be found.
func GetUser(username string, password string) (userdataptr *User, err error) {
	// Retrieve keys and UUID of the userdata
	uuidKey, encKey, hmacKey := UserHKDF(username, password)
	uuidUD := bytesToUUID(uuidKey)

	// Retrieve userdata
	serialBlob, boolean := userlib.DatastoreGet(uuidUD)

	if boolean == false {
		err := errors.New("user cannot be found")
		return nil, err
	}
	var blob Blob
	_ = json.Unmarshal(serialBlob, &blob)

	verify, err := blob.VerifyHMAC(hmacKey)
	if err != nil || !verify {
		return nil, err
	}

	decryptUD := userlib.SymDec(encKey, blob.Data)
	var ud User
	_ = json.Unmarshal(decryptUD, &ud)
	return &ud, nil
}

// Encrypts file with random key, signs it with user's sign key, and
// puts into Blob structure and upload to datastore
func UploadFile(data []byte, signKey userlib.DSSignKey) (uuid.UUID, []byte){
	encKey := userlib.RandomBytes(16)
	encryptedData := userlib.SymEnc(encKey, userlib.RandomBytes(16), data)
	ds, _ := userlib.DSSign(signKey, encryptedData)

	serialBlob, _ := json.Marshal(NewBlob(encryptedData, ds))
	uuidFile := uuid.New()
	userlib.DatastoreSet(uuidFile, serialBlob)

	return uuidFile, encKey
}

// This stores a file in the datastore.
//
// The name and length of the file should NOT be revealed to the datastore!
// key length: 16 bytes ..
func (userdata *User) StoreFile(filename string, data []byte) {
	// Encrypt and sign file, upload Blob to datastore
	uuidFile, encKey := UploadFile(data, userdata.SignKey)

	// Create User File and upload to datastore
	userFile := userdata.NewUserFile(uuid.Nil, nil, make(map[int][4][]byte), [2][]byte{}, make(map[int][4][]byte))
	userFile.UpdateMetadata(userdata.Username, userdata, uuidFile, encKey)

	serialUserFile, _ := json.Marshal(userFile)
	uuidUserFile := uuid.New()
	userlib.DatastoreSet(uuidUserFile, serialUserFile)

	// Update userdata and upload to datastore
	userdata.SetFileNameToUUID(filename, uuidUserFile)
	var blob Blob
	_ = userdata.UploadUser(&blob)
}

// Not encrypted bc thinking what's the point. Hashed tho to hide file name length
func (userdata *User) SetFileNameToUUID(filename string, uuid uuid.UUID) {
	name, _ := userlib.HMACEval(userdata.HMACKey, []byte(filename))
	fakeUUID := bytesToUUID(name)
	userdata.Files[fakeUUID] = uuid

}

func (userdata *User) GetUUIDFromFileName(filename string) (uuid uuid.UUID, ok bool) {
	name, _ := userlib.HMACEval(userdata.HMACKey, []byte(filename))
	fakeUUID := bytesToUUID(name)
	uuid, ok = userdata.Files[fakeUUID]
	return
}

// This adds on to an existing file.
//
// Append should be efficient, you shouldn't rewrite or reencrypt the
// existing file, but only whatever additional information and
// metadata you need.
func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	// need to retrieve the userfile
	uuidUserFile, ok := userdata.GetUUIDFromFileName(filename)

	if !ok {
		return errors.New("file does not exist")
	}

	userFile, err := RetrieveUserFile(uuidUserFile)
	if err != nil {
		return err
	}

	// Upload file
	uuidNewFile, encKey := UploadFile(data, userdata.SignKey)

	// retrieve owner of file
	for ; err == nil && userFile.Parent != uuid.Nil; {
		userFile, err = RetrieveUserFile(userFile.Parent)
	}

	// iterate through children...
	if err != nil {
		return err
	}

	userFile.UpdateAllMetadata(userdata, uuidNewFile, encKey)
	return nil
}

// This loads a file from the Datastore.
//
// It should give an error if the file is corrupted in any way.
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	uuidUF, ok := userdata.GetUUIDFromFileName(filename)

	if !ok {
		return nil, errors.New("file does not exist")
	}

	userFile, err := RetrieveUserFile(uuidUF)
	if err != nil {
		return nil, err
	}

	var file []byte
	verified := userFile.ValidUsers()
	if len(userFile.SavedMeta) > 0 {
		// Verify signer is a permissible user
		name := string(userFile.SavedMetaDS[0])
		if _, ok := verified[name]; !ok || userdata.Username != name {
			return nil, errors.New("saved was corrupted")
		}

		// Check digital signature
		msg, _ := json.Marshal(userFile.SavedMeta)
		key, _ := GetPublicVerKey(string(userFile.SavedMetaDS[0]))
		err := userlib.DSVerify(key, msg, userFile.SavedMetaDS[1])

		// Error check
		if err != nil {
			return nil, errors.New("saved metadata is invalid")
		}

		// Evaluate items in SavedMeta. We do NOT need to verify each signature.
		for _, elem := range userFile.SavedMeta {
			fileBlock, err := EvaluateMetadata(userdata, elem,-1)
			if err != nil {
				return nil, err
			}
			file = append(file, fileBlock...)

		}
	}

	// Evaluate items in ChangesMeta.
	for index := 0; index < len(userFile.ChangesMeta); index++ {
		username, err := userlib.PKEDec(userdata.DecKey, userFile.ChangesMeta[index][0])

		if err != nil {
			return nil, errors.New("RSA decryption failed")
		}

		// Verify that the sender is valid
		if _, ok := verified[string(username)]; !ok {
			return file, errors.New("rest of file corrupted")
		}

		fileBlock, err := EvaluateMetadata(userdata, userFile.ChangesMeta[index], index)
		if err != nil {
			return nil, err
		}
		file = append(file, fileBlock...)
	}

	return file, nil
}

func EvaluateMetadata(user *User, meta [4][]byte, index int) ([]byte, error){
	// Decrypt username, elem[0]
	n, err0 := user.Decrypt(meta[0])

	if err0 != nil {
		return nil, err0
	}

	name := string(n)
	verKey, ok := GetPublicVerKey(name)

	if !ok {
		return nil, errors.New("user's verification key does not exist")
	}

	// Verify metadata first. Only for changeslist.
	if index != -1 {
		msg := string(index) + string(meta[1]) + string(meta[2])
		err := userlib.DSVerify(verKey, []byte(msg), meta[3])
		if err != nil {
			return nil, err
		}
	}

	// Decrypt uuid, elem[1]
	uuidByte, err1 := user.Decrypt(meta[1])
	uuidBlob, err2 := uuid.ParseBytes(uuidByte)

	if err1 != nil {
		return nil, err1
	} else if err2 != nil {
		return nil, err2
	}

	// Retrieve the blob
	serialBlob, ok := userlib.DatastoreGet(uuidBlob)
	if !ok {
		return nil, errors.New("file does not exist")
	}

	var blob Blob
	_ = json.Unmarshal(serialBlob, &blob)

	// Verify the blob's signature
	err := userlib.DSVerify(verKey, blob.Data, blob.Check)
	if err != nil {
		return nil, err
	}

	// Decrypt key, elem[2], and decrypt file
	eKey, err := user.Decrypt(meta[2])
	if err != nil {
		return nil, err
	}

	decryptedFile := userlib.SymDec(eKey, blob.Data)

	return decryptedFile, nil
}

// Checks to see if the user is in one of the permissions, either parent or children.
func (userFile *UserFile) ValidUsers() map[string]uuid.UUID {
	verified := make(map[string]uuid.UUID)
	// Traverse up to the owner of the file and add the owner to our verified list

	owner := userFile
	for owner.Parent != uuid.Nil {
		// Verify that parent is signed first
		verKey, ok := GetPublicVerKey(owner.Username)
		if ok {
			err := userlib.DSVerify(verKey, []byte(userFile.Parent.String()), userFile.ParentDS)
			if err == nil {
				parent, err := RetrieveUserFile(owner.Parent)
				if err != nil {
					break
				}
				owner = parent
			}
		}
	}

	_ = Traverse(&verified, owner)
	return verified
}

func Traverse(verified *map[string]uuid.UUID, userFile *UserFile) error {
	(*verified)[userFile.Username] = userFile.UUID

	// Verify children
	verKey, ok := GetPublicVerKey(userFile.Username)
	if !ok {
		return errors.New("no public ver key")
	}

	serialChildren, _ := json.Marshal(userFile.Children)
	err := userlib.DSVerify(verKey, serialChildren, userFile.ChildrenDS)
	if err != nil {
		return err
	}

	// Visit children
	for _, uuidChild := range userFile.Children {
		uFile, err := RetrieveUserFile(uuidChild)
		if err != nil {
			return errors.New("user does not have permission")
		}
		_ = Traverse(verified, uFile)
	}
	return nil
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
