package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
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
	//IDds []byte // ds of []byte (username + string(uuid)) //TODO : INTEGRATE THISSSS

	Children map[string]uuid.UUID // list of descendents with access to file. username - uuid
	ChildrenDS []byte // digital signature of children

	Parent uuid.UUID // uuid of parent that gave this user access. uuid.NIL if no parent
	ParentDS []byte // digital signature by the parent. NIL if no parent. MARSHAL parent

	SavedMeta map[int][4][]byte // e(username), e(uuid), e(key), ds
	SavedMetaDS [2][]byte // first element is username, second element is DS of user. non encrypted
	ChangesMeta map[int][4][]byte // ds is: i + eUUID + eKey
}
// todo when sharing access: need to write childrenDS and parentDS
func (userdata *User) NewUserFile(username string, UUID uuid.UUID, parent uuid.UUID) *UserFile {
	var f UserFile
	f.Username = username
	f.UUID = UUID
	//f.IDds, _ = userlib.DSSign(userdata.SignKey, []byte(f.Username + f.UUID.String()))

	f.Children = make(map[string]uuid.UUID)
	f.Parent = parent
	if parent != uuid.Nil {
		msg, _ := json.Marshal(f.Parent)

		parentDS, _ := userlib.DSSign(userdata.SignKey, msg)
		f.ParentDS = parentDS
	}

	f.SavedMeta = make(map[int][4][]byte)
	f.ChangesMeta = make(map[int][4][]byte)
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

// Adds one meta at a time to saveddata. Does not need DS.
func (userFile *UserFile) UpdateSavedMetadata(sender string, uuidFile uuid.UUID, encKey []byte) {
	pubKey, _ := GetPublicEncKey(userFile.Username)

	eUsername, _ := userlib.PKEEnc(pubKey, []byte(sender))
	eUUID, _ := userlib.PKEEnc(pubKey, []byte(uuidFile.String()))
	eKey, _ := userlib.PKEEnc(pubKey, encKey)

	userFile.SavedMeta[len(userFile.SavedMeta)] = [4][]byte{eUsername, eUUID, eKey, nil}
}

func (userFile *UserFile) TransferChangesToSavedMeta(meta [4][]byte) {
	userFile.SavedMeta[len(userFile.SavedMeta)] = meta
}

// Updates the metadata of a recipient's changes map, with a file update by sender
func (userFile *UserFile) UpdateMetadata(sender *User, uuidFile uuid.UUID, encKey []byte) {
	pubKey, _ := GetPublicEncKey(userFile.Username)

	eUsername, _ := userlib.PKEEnc(pubKey, []byte(sender.Username))
	eUUID, _ := userlib.PKEEnc(pubKey, []byte(uuidFile.String()))
	eKey, _ := userlib.PKEEnc(pubKey, encKey)

	msg := string(len(userFile.ChangesMeta)) + string(eUUID) + string(eKey)
	ds, _ := userlib.DSSign(sender.SignKey, []byte(msg))

	// todo: need to change the index shit. bc we change index every time we add to saved. or do we need to update this????
	userFile.ChangesMeta[len(userFile.ChangesMeta)] = [4][]byte{eUsername, eUUID, eKey, ds}

	serialUF, _ := json.Marshal(userFile)
	userlib.DatastoreSet(userFile.UUID, serialUF)
}

// Calls UpdateMetadata on this userfile and all of its children and its children...
func (userFile *UserFile) UpdateAllMetadata(sender *User, uuidFile uuid.UUID, encKey []byte) {
	userFile.UpdateMetadata(sender, uuidFile, encKey)
	if len(userFile.Children) > 0 { // TODO: in share access, add another bool later to verify DS
		for _, uuidChild := range userFile.Children {
			userFile, err := RetrieveUserFile(uuidChild)
			if err == nil {
				//userFile.UpdateMetadata(sender, uuidFile, encKey)
				userFile.UpdateAllMetadata(sender, uuidFile, encKey) //todo
			}
		}
	}
}

// Decrypts metadata. Returns all of the elements.
func (userdata *User) DecryptMeta(i int, meta [4][]byte) (string, uuid.UUID, []byte, error) {
	n, err := userdata.Decrypt(meta[0])

	if err != nil {
		return "", uuid.Nil, nil, err
	}

	name := string(n)
	verKey, ok := GetPublicVerKey(name)
	if !ok {
		err = errors.New("user's verification key does not exist")
		return "", uuid.Nil, nil, err
	}

	// Verify metadata first. Only for changeslist.
	if i != -1 {
		msg := string(i) + string(meta[1]) + string(meta[2])
		err := userlib.DSVerify(verKey, []byte(msg), meta[3])
		if err != nil {
			return "", uuid.Nil, nil, err
		}
	}

	// Decrypt uuid, elem[1]
	uuidBlobByte, err1 := userdata.Decrypt(meta[1])
	uuidBlob, err2 := uuid.ParseBytes(uuidBlobByte)

	if err1 != nil {
		return "", uuid.Nil, nil, err1
	} else if err2 != nil {
		return "", uuid.Nil, nil, err2
	}

	key, err := userdata.Decrypt(meta[2])
	if err != nil {
		return "", uuid.Nil, nil, err
	}

	return name, uuidBlob, key, nil

}


// Decrypts ciphertext using the user's private decryption key.
func (userdata *User) Decrypt(ciphertext []byte) (plaintext []byte, err error) {
	plaintext, err = userlib.PKEDec(userdata.DecKey, ciphertext)
	return
}

// Retrieve the public encryption key in Keystore under the name username.
func GetPublicEncKey(username string) (userlib.PKEEncKey, bool) {
	return userlib.KeystoreGet(username + "enc")
}

// Retrieve the public signature key in Keystore under the name username.
func GetPublicVerKey(username string) (userlib.DSVerifyKey, bool) {
	return userlib.KeystoreGet(username + "sign")
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
// The name and length of the file should NOT be revealed to the datastore!
// key length: 16 bytes ..
func (userdata *User) StoreFile(filename string, data []byte) {
	// Encrypt and sign file, upload Blob to datastore
	uuidFile, encKey := UploadFile(data, userdata.SignKey)

	// Create User File and upload to datastore
	uuidUserFile := uuid.New()
	userFile := userdata.NewUserFile(userdata.Username, uuidUserFile, uuid.Nil)
	userFile.UpdateMetadata(userdata, uuidFile, encKey)

	serialUserFile, _ := json.Marshal(userFile)
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

// returns true if filename exists
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
	userFile = userFile.RetrieveOwner()
	//for err == nil && userFile.Parent != uuid.Nil {
	//	userFile, err = RetrieveUserFile(userFile.Parent)
	//}

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
		if _, ok := verified[name]; !ok {
			return nil, errors.New("can't load, saved was corrupted")
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
			return file, err
		}

		file = append(file, fileBlock...)
	}

	return file, nil
}

func EvaluateMetadata(user *User, meta [4][]byte, index int) ([]byte, error){
	name, uuidBlob, key, err := user.DecryptMeta(index, meta)
	if err != nil {
		return nil, err
	}
	// Retrieve the blob
	serialBlob, ok := userlib.DatastoreGet(uuidBlob)
	if !ok {
		return nil, errors.New("file does not exist")
	}

	var blob Blob
	_ = json.Unmarshal(serialBlob, &blob)

	// Verify the blob's signature
	verKey, ok := GetPublicVerKey(name)
	err = userlib.DSVerify(verKey, blob.Data, blob.Check)
	if err != nil {
		return nil, err
	}

	decryptedFile := userlib.SymDec(key, blob.Data)

	return decryptedFile, nil
}

// Get owner of file
func (userFile *UserFile) RetrieveOwner() *UserFile {
	owner := userFile
	for owner.Parent != uuid.Nil {
		// Verify that parent is signed first
		parent, err := RetrieveUserFile(owner.Parent)
		if err != nil {
			break
		}

		verKey, ok := GetPublicVerKey(parent.Username)
		if !ok {
			break
		}
			msg, _ := json.Marshal(userFile.Parent)
			err = userlib.DSVerify(verKey, msg, userFile.ParentDS)
			if err != nil {
				break
			}

			owner = parent
	}

	return owner
}
// Checks to see if the user is in one of the permissions, either parent or children.
func (userFile *UserFile) ValidUsers() map[string]uuid.UUID {
	verified := make(map[string]uuid.UUID)

	// Traverse up to the owner of the file
	owner := userFile.RetrieveOwner()

	// could be recurisve problem in traverse
	//fmt.Print(userFile.Username)
	//if userFile.Username == "bob" {
	//	fmt.Print("STARTING")
	//}
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
	// Check that the file exists. Return empty string if false.
	uuidUF, ok := userdata.GetUUIDFromFileName(filename)
	if !ok {
		return "", errors.New("file does not exist")
	}

	// Retrieve the UserFile
	uf, err := RetrieveUserFile(uuidUF)
	if err != nil {
		return "", errors.New("file does not exist")
	}

	// Create a new UserFile for the recipient
	uuidNewUF := uuid.New()
	newUF := userdata.NewUserFile(recipient, uuidNewUF, uf.UUID)

	// Re-encrypt metadata for recipient
	verified := uf.ValidUsers()

	// First, verify saved. Add saved to newUF.
	if len(uf.SavedMeta) > 0 {
		name := string(uf.SavedMetaDS[0])
		if _, ok := verified[name]; !ok {
			return "", errors.New("can't share, saved was corrupted")
		}

		msg, _ := json.Marshal(uf.SavedMeta)
		key, _ := GetPublicVerKey(name)
		err := userlib.DSVerify(key, msg, uf.SavedMetaDS[1])
		if err != nil {
			return "", err
		}

		// Add this to the new uf's save
		for _, meta := range uf.SavedMeta {
			editor, uuidFile, key, err := userdata.DecryptMeta(-1, meta)
			if err != nil {
				return "", nil
			}
			newUF.UpdateSavedMetadata(editor, uuidFile, key)
		}
	}

	// Then, verify changes. Add changes meta to saved for new UF.
	for index := 0; index < len(uf.ChangesMeta); index++ {
		n, err := userlib.PKEDec(userdata.DecKey, uf.ChangesMeta[index][0])
		name := string(n)
		if err != nil {
			return "", errors.New("RSA decryption failed")
		}

		// Verify that the sender is valid
		if _, ok := verified[name]; !ok {
			return "", errors.New("rest of file corrupted")
		}

		name, uuidFile, key, err := userdata.DecryptMeta(index, uf.ChangesMeta[index])
		if err != nil {
			return "", err
		}

		newUF.UpdateSavedMetadata(name, uuidFile, key)
	}

	// Now sign saved.
	newUF.SavedMetaDS[0] = []byte(userdata.Username)
	msg, _ := json.Marshal(newUF.SavedMeta)
	newUF.SavedMetaDS[1], _ = userlib.DSSign(userdata.SignKey, msg)

	// For all UFs: Append each meta in changes to saved.
	for _, UUID := range verified {
		otherUF, err := RetrieveUserFile(UUID)
		if err == nil {
			for _, meta := range otherUF.ChangesMeta {
				otherUF.TransferChangesToSavedMeta(meta)
			}

			// Sign and re-upload.
			msg, _ := json.Marshal(otherUF.SavedMeta)
			ds, _ := userlib.DSSign(userdata.SignKey, msg)
			otherUF.SavedMetaDS[0] = []byte(userdata.Username)
			otherUF.SavedMetaDS[1] = ds

			serialUF, _ := json.Marshal(otherUF)
			userlib.DatastoreSet(otherUF.UUID, serialUF)
		}
	}

	// Include the new UF in our children. Sign it.
	uf.Children[recipient] = newUF.UUID
	msg, _ = json.Marshal(uf.Children)
	ds, _ := userlib.DSSign(userdata.SignKey, msg)
	uf.ChildrenDS = ds

	// Upload the new UF
	serialUF, _ := json.Marshal(newUF)
	userlib.DatastoreSet(newUF.UUID, serialUF)

	// Reupload our UF
	serialUF, _ = json.Marshal(newUF)
	userlib.DatastoreSet(newUF.UUID, serialUF)

	// Generate the magic string: w/ random uuid and signature. encrypted ofc
	pubKey, _ := GetPublicEncKey(recipient)
	byteUUID, _ := json.Marshal(newUF.UUID.String())
	eUUID, _ := userlib.PKEEnc(pubKey, byteUUID)
	msDS, _ := userlib.DSSign(userdata.SignKey, eUUID)

	magic_string = string(eUUID) + string(msDS)

	return magic_string, nil
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {

	// if filename already exists error out
	_, notOK := userdata.GetUUIDFromFileName(filename)
	if notOK {
		return errors.New("file already exists")
	}

	eUUID := []byte(magic_string[:256])
	msDS := []byte(magic_string[256:])

	// Verify signer
	verKey, _ := GetPublicVerKey(sender)
	err := userlib.DSVerify(verKey, eUUID, msDS)
	if err != nil {
		return errors.New("authenticity compromised")
	}

	// Decrypt magic string with private key to retrieve UUID
	msg, _ := userlib.PKEDec(userdata.DecKey, eUUID)
	var uuidUF uuid.UUID
	_ = json.Unmarshal(msg, &uuidUF)

	// Assign hashed file name to UUID
	userdata.SetFileNameToUUID(filename, uuidUF)

	// Access UF to confirm
	_, err = RetrieveUserFile(uuidUF)
	if err != nil {
		return err
	}

	// Reload userdata
	var blob Blob
	err = userdata.UploadUser(&blob)

	if err != nil {
		return err
	}

	return nil
}

// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	// retrieve userfile
	//uuidUF, ok := userdata.GetUUIDFromFileName(filename)
	//if !ok {
	//	return errors.New("file does not exist")
	//}

	// verify saved and changes --> add changes to saved and sign


	// do this for every user EXCEPT revoked user (unless you have to then it's ok)
	// reupload every UF
	// verify UF.children
	// go to uuid of target child
	// set parent to none
	// then delete their UF //todo: make sure that errors pop up if child tries to access removed parent
	// remove child from children. re-sign
	//reserialize and reupload our UF
	return
}
