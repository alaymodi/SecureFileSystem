package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	// You neet to add with
	// go get github.com/cs161-staff/userlib
	"github.com/cs161-staff/userlib"
	 _ "reflect"
	_ "strconv"

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

type Access struct {
	File_information []byte
	Access_Structs_shared map[string][]byte
}


type File struct {
	IV [][]byte
	Content_UUIDs []userlib.UUID
	Content_HMACs [][]byte
	Size int
	//hmac
}

// The structure definition for a user record
type User struct {
	Uuid userlib.UUID
	Username string
	Hmackey []byte
	IV [] byte
	Secret_key userlib.PKEDecKey
	Sign_key userlib.DSSignKey
	Files_owned map[string]string
	Access_shared_information map[string][]byte
	// You can add other fields here if you want...
	// Note for JSON to marshal/unmarshal, the fields need to
	// be public (start with a capital letter)
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
	var userdata User

	usr_argon2key := userlib.Argon2Key([]byte(username + password), []byte("nosalt"), 16)
	usr_uuid, err := uuid.FromBytes(usr_argon2key)
	if err != nil {
		return nil, errors.New("error creating a uuid from argon2key")
	}
	userdata.Uuid = usr_uuid
	userdata.Username = username
	//userdata.Hmackey =
	userlib.Argon2Key([]byte(username + "sad"), []byte("salty"), 16)//userlib.RandomBytes(16) //  []byte("0123456789012345")//
	userdata.IV = userlib.RandomBytes(16)

	pkenckey, pkedeckey, err1 := userlib.PKEKeyGen()
	if err1 != nil {
		return nil, errors.New("could not generate a public key pair for asymmetric encryption")
	}
	userdata.Secret_key = pkedeckey

	err2 := userlib.KeystoreSet(username, pkenckey)
	if err2 != nil {
		return nil, errors.New("could not set public encryption key in keystore ")
	}

	dssignkey, dsverifykey, err3 := userlib.DSKeyGen()
	if err3 != nil {
		return nil, errors.New("could not generate digital signature for asymmetric signing")
	}
	userdata.Sign_key = dssignkey

	err4 := userlib.KeystoreSet(username+"sig", dsverifykey)
	if err4 != nil {
		return nil, errors.New("could not set digital verify key in keystore")
	}

	userdata.Files_owned = make(map[string]string)
	userdata.Access_shared_information = make(map[string][]byte)

	// encrypt usr struct and place into datastore
	usr_struct_bytes, error := json.Marshal(userdata)
	if error != nil {
		return nil, errors.New("could not marshal usr struct properly")
	}
	enc_user_struct := userlib.SymEnc([]byte(userdata.Uuid.String() + userdata.Username)[:16], userlib.RandomBytes(16), usr_struct_bytes)
	user_struct_sig, err5 := userlib.HMACEval(userlib.Argon2Key([]byte(username + "sad"), []byte("salty"), 16), enc_user_struct)//userdata.Hmackey, enc_user_struct)
	if err5 != nil {
		return nil, errors.New("encrypted user struct could not be signed")
	}

	userlib.DebugMsg("length %v", len([]byte(userdata.Uuid.String())))
	userlib.DatastoreSet(userdata.Uuid, []byte(string(user_struct_sig) + string(enc_user_struct)))
	return &userdata, nil
}

func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	//userlib.DebugMsg(username)
	usr_argon2key := userlib.Argon2Key([]byte(username + password), []byte("nosalt"), 16)
	usr_uuid, err := uuid.FromBytes(usr_argon2key)

	hmac_enc_usr_struct, ok := userlib.DatastoreGet(usr_uuid)
	if len(hmac_enc_usr_struct) < 64 {
		return nil, errors.New("error malicious user must have tampered with it")
	}
	if !ok {
		return nil, errors.New("could not find a user struct for this user")
	}
	hmac_check , err2:= userlib.HMACEval(userlib.Argon2Key([]byte(username + "sad"), []byte("salty"), 16),hmac_enc_usr_struct[64:])//userdata.Hmackey, hmac_enc_usr_struct[64:]) //[]byte("0123456789012345"), dec_usr_struct_hmac[64:])//
	if err2 != nil {
		return nil, errors.New("could not call hmac on usr struct in byte form")
	}

	append_hmac := hmac_enc_usr_struct[:64]
	eq := userlib.HMACEqual(append_hmac, hmac_check)
	//userlib.DebugMsg("got user struct %v", hmac_enc_usr_struct[64:])
	if !eq {
		userlib.DebugMsg("error here?/?")
		return nil, errors.New("hmac is not the same so user struct has been tampered with")
	}

	dec_usr_struct_hmac := userlib.SymDec([]byte(usr_uuid.String() + userdata.Username)[:16], hmac_enc_usr_struct[64:])
	err1 := json.Unmarshal(dec_usr_struct_hmac, &userdata)
	if err1 != nil {
		return nil, errors.New("could not unmarshal user struct")
	}

	return &userdata, nil
}


func (userdata *User) StoreFile(filename string, data []byte) {

	if userdata.Uuid == uuid.Nil || userdata.Username == ""  || userdata.IV == nil{
		  errors.New("store user has not been initialized")
		return
	}
	_, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		errors.New("store user has not been initialized")
		return
	}
	_, ok = userlib.KeystoreGet(userdata.Username)
	if !ok {
		errors.New("store user has not been initialized")
		return
	}
	userdata.Files_owned[filename] = "owned"
	access_loc := uuid.New()
	sym_enc_key := userlib.RandomBytes(16)
	hmac_key := userlib.RandomBytes(16)

	userdata.Access_shared_information[filename] = []byte(string(sym_enc_key) + string(hmac_key) + access_loc.String())

	//save new user state
	usr_struct_bytes, error := json.Marshal(userdata)
	if error != nil {
		errors.New("could not marshal usr struct properly")
		return
	}
	enc_user_struct := userlib.SymEnc([]byte(userdata.Uuid.String() + userdata.Username)[:16], userlib.RandomBytes(16), []byte (string(usr_struct_bytes)))

	user_struct_sig, err := userlib.HMACEval(userlib.Argon2Key([]byte(userdata.Username + "sad"), []byte("salty"), 16),enc_user_struct)//userdata.Hmackey, enc_user_struct)
	if err != nil {
		errors.New("encrypted user struct could not be signed")
		return
	}

	userlib.DatastoreSet(userdata.Uuid, []byte(string(user_struct_sig)+string(enc_user_struct)))

	var _access Access
	file_location := uuid.New()
	file_sym_key := userlib.RandomBytes(16)
	file_hmac_key := userlib.RandomBytes(16)
	_access.Access_Structs_shared = make(map[string][]byte)
	_access.File_information = []byte(string(file_sym_key) + string(file_hmac_key) + file_location.String())

	//save new user state
	access_struct_bytes, error := json.Marshal(_access)
	if error != nil {
		//errors.New("could not marshal usr struct properly")
	}

	enc_access_struct := userlib.SymEnc(userdata.Access_shared_information[filename][:16], userlib.RandomBytes(16), []byte (string(access_struct_bytes)))

	access_struct_sig, err := userlib.HMACEval(userdata.Access_shared_information[filename][16:32], enc_access_struct)
	if err != nil {
		//errors.New("encrypted user struct could not be signed")
		return
	}
	userlib.DatastoreSet(access_loc, []byte (string(access_struct_sig) + string(enc_access_struct)))

	_access.StoreFile(filename, file_location, data, file_sym_key, file_hmac_key)
}

func (access *Access) StoreFile(filename string, location userlib.UUID, data []byte, sym_key []byte, hmac_key []byte) {
	// create file

	var userfile File

	data_uuid := uuid.New()
	data_iv := userlib.RandomBytes(16)
	data_hmac := userlib.RandomBytes(16)
	enc_data_slice := userlib.SymEnc([]byte(data_uuid.String() + string(data_iv))[:16], userlib.RandomBytes(16), data)
	enc_data_sig, err1 := userlib.HMACEval(data_hmac, enc_data_slice)
	if err1 != nil {
		//errors.New("encrypted data slice could not be signed")
		return
	}
	userlib.DatastoreSet(data_uuid, []byte(string(enc_data_sig) + string(enc_data_slice)))

	userfile.IV = append(userfile.IV, data_iv)
	userfile.Content_UUIDs = append(userfile.Content_UUIDs, data_uuid)
	userfile.Content_HMACs = append(userfile.Content_HMACs, data_hmac)
	userfile.Size = 1

	// encrypt the file struct it self and sign then place in data store

	file_struct_bytes, err2 := json.Marshal(userfile)
	if err2 != nil {
		//errors.New("could not marshal file struct properly")
		return
	}
	// encrypt then sign
	enc_file_struct := userlib.SymEnc(sym_key[:16], userlib.RandomBytes(16), file_struct_bytes)

	file_struct_sig, err3 := userlib.HMACEval(hmac_key, enc_file_struct)
	if err3 != nil {
		//errors.New("encrypted file struct could not be signed")
		return
	}
	userlib.DebugMsg("location stored %v", []byte(location.String()))
	userlib.DatastoreSet(location, []byte (string(file_struct_sig) + string(enc_file_struct)))

}

func (userdata *User) AppendFile(filename string, data[]byte) (err error) {
	if userdata.Uuid == uuid.Nil || userdata.Username == ""  || userdata.IV == nil{
		return  errors.New("share user has not been initialized")
	}
	_, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		return  errors.New("share user has not been initialized")
	}
	_, ok = userlib.KeystoreGet(userdata.Username)
	if !ok {
		return  errors.New("share user has not been initialized")
	}
	var access Access
	access, err = LoadFileInfo(userdata, filename)
	if err != nil {
		return err
	}
	file_sym_key := access.File_information[0:16]
	file_hmac_key := access.File_information[16:32]
	file_location_bytes := access.File_information[32:]

	//if err != nil {
	//	return errors.New("could not decrypt  file information")
	//}

	err = access.AppendFile(filename, data, file_sym_key, file_hmac_key, file_location_bytes)
	if err != nil {
		return err
	}

	return nil
}

func (_access *Access) AppendFile(filename string, data []byte, file_sym_key []byte, file_hmac_key []byte, file_location_bytes []byte) (err error) {
	var userfile File
	userlib.DebugMsg("HOLY")
	file_location, err := uuid.ParseBytes(file_location_bytes)
	if err != nil {
		return errors.New("cannot get file location")
	}

	enc_file_struct_hmac, ok := userlib.DatastoreGet(file_location)
	if !ok {
		return errors.New("could not find a file struct for this filename")
	}

	append_hmac := enc_file_struct_hmac[:64]
	hmac_check , err:= userlib.HMACEval(file_hmac_key, enc_file_struct_hmac[64:])
	if err != nil {
		return errors.New("could not call hmac on file encrypted struct form")
	}
	eq := userlib.HMACEqual(append_hmac, hmac_check)
	if !eq {
		userlib.DebugMsg("hello111")
		return errors.New("hmac is not the same so file struct has been tampered with")
	}

	dec_file_struct := userlib.SymDec(file_sym_key, enc_file_struct_hmac[64:])
	err1 := json.Unmarshal(dec_file_struct, &userfile)
	if err1 != nil {
		return errors.New("could not unmarshal file struct")
	}

	if userfile.Size != len(userfile.Content_UUIDs) || userfile.Size != len(userfile.Content_HMACs) || userfile.Size != len(userfile.IV) {
		return errors.New("parts of our file struct was modified")
	}

	for i := 0; i < userfile.Size; i++ {

		enc_content_slice_hmac, ok := userlib.DatastoreGet(userfile.Content_UUIDs[i])
		if !ok {
			return errors.New("could not get the encrypted file slice")
		}

		append_hmac := enc_content_slice_hmac[:64]
		hmac_check , err2:= userlib.HMACEval(userfile.Content_HMACs[i], enc_content_slice_hmac[64:])
		if err2 != nil {
			return errors.New("could not call hmac on encrypted file content slice form")
		}
		eq := userlib.HMACEqual(append_hmac, hmac_check)
		if !eq {
			userlib.DebugMsg("incorrect")
			return errors.New("hmac is not the same so encrypted file content has been tampered with")
		}
	}

	data_uuid := uuid.New()
	data_iv := userlib.RandomBytes(16)
	data_hmac := userlib.RandomBytes(16)
	enc_data_slice := userlib.SymEnc([]byte(data_uuid.String() + string(data_iv))[:16], userlib.RandomBytes(16), data)
	enc_data_sig, err1 := userlib.HMACEval(data_hmac, enc_data_slice)
	if err1 != nil {
		return errors.New("encrypted data slice could not be signed")
	}
	userlib.DatastoreSet(data_uuid, []byte(string(enc_data_sig) + string(enc_data_slice)))

	userfile.IV = append(userfile.IV, data_iv)
	userfile.Content_UUIDs = append(userfile.Content_UUIDs, data_uuid)
	userfile.Content_HMACs = append(userfile.Content_HMACs, data_hmac)
	userfile.Size++

	// encrypt the file struct it self and sign then place in data store

	file_struct_bytes, err2 := json.Marshal(userfile)
	if err2 != nil {
		return errors.New("could not marshal file struct properly")
	}
	// encrypt then sign
	enc_file_struct := userlib.SymEnc(file_sym_key, userlib.RandomBytes(16), file_struct_bytes)

	file_struct_sig, err3 := userlib.HMACEval(file_hmac_key, enc_file_struct)
	if err3 != nil {
		return errors.New("encrypted file struct could not be signed")
	}

	userlib.DatastoreSet(file_location, []byte (string(file_struct_sig) + string(enc_file_struct)))
	return nil


}

func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	if userdata.Uuid == uuid.Nil || userdata.Username == ""  || userdata.IV == nil{
		return nil, errors.New("load user has not been initialized")
	}
	_, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		return nil, errors.New("load user has not been initialized")
	}
	_, ok = userlib.KeystoreGet(userdata.Username)
	if !ok {
		return nil, errors.New("load user has not been initialized")
	}
	var access Access
	access, err= LoadFileInfo(userdata, filename)
	if err != nil {
		return nil, err
	}
	file_sym_key := access.File_information[0:16]
	file_hmac_key := access.File_information[16:32]
	file_location := access.File_information[32:]

	return access.LoadFile(filename, file_location, file_sym_key, file_hmac_key)
}

func LoadFileInfo(userdata *User, filename string) (access Access, err error) {
	_, ok := userdata.Access_shared_information[filename];
	//if len(shared_info) != 68 {
	//	return access, errors.New("database has been corrupted")
	//}
	if !ok {
		return access, errors.New("You do not have this file")
	}
	access_location_bytes := userdata.Access_shared_information[filename][32:]
	access_location, err := uuid.ParseBytes(access_location_bytes)
	if(err != nil) {
		return access, errors.New("Nothing stored at this access location")
	}

	enc_file_access_hmac, ok := userlib.DatastoreGet(access_location)
	if !ok {
		return access, errors.New("You no longer have access")
	}

	append_hmac := enc_file_access_hmac[:64]
	hmac_check , err:= userlib.HMACEval(userdata.Access_shared_information[filename][16:32], enc_file_access_hmac[64:])
	if err != nil {
		return access, errors.New("could not call hmac on file encrypted struct form")
	}

	eq := userlib.HMACEqual(append_hmac, hmac_check)
	if !eq {
		userlib.DebugMsg("incorrect in access struct ")
		return access, errors.New("hmac is not the same so file struct has been tampered with?")
	}
	dec_access_struct := userlib.SymDec(userdata.Access_shared_information[filename][:16], enc_file_access_hmac[64:])
	err1 := json.Unmarshal(dec_access_struct, &access)

	if err1 != nil {
		return access, errors.New("could not unmarshal file struct")
	}
	return access, nil
}

func (access *Access) LoadFile(filename string, file_location_byte []byte, file_sym_key []byte, file_hmac_key []byte) (data []byte, err error) {
	var userfile File

	file_location, err := uuid.ParseBytes(file_location_byte)
	//userlib.DebugMsg("file location found %v", []byte(file_location.String()))
	if err != nil {
		return nil, errors.New("could not find file")
	}
	enc_file_struct_hmac, ok := userlib.DatastoreGet(file_location)
	if !ok {
		return nil, errors.New("could not find a file struct for this filename")
	}

	append_hmac := enc_file_struct_hmac[:64]
	hmac_check , err:= userlib.HMACEval(file_hmac_key, enc_file_struct_hmac[64:])
	if err != nil {
		return nil, errors.New("could not call hmac on file encrypted struct form")
	}
	eq := userlib.HMACEqual(append_hmac, hmac_check)
	if !eq {
		return nil, errors.New("hmac is not the same so file struct has been tampered with")
	}

	dec_file_struct := userlib.SymDec(file_sym_key, enc_file_struct_hmac[64:])
	err1 := json.Unmarshal(dec_file_struct, &userfile)
	if err1 != nil {
		userlib.DebugMsg("hiiii")

		return nil, errors.New("could not unmarshal file struct")
	}

	if userfile.Size != len(userfile.Content_UUIDs) || userfile.Size != len(userfile.Content_HMACs) || userfile.Size != len(userfile.IV) {
		return nil, errors.New("parts of our file struct was modified")
	}
	for i := 0; i < userfile.Size; i++ {

		enc_content_slice_hmac, ok := userlib.DatastoreGet(userfile.Content_UUIDs[i])
		if !ok {
			return nil, errors.New("could not get the encrypted file slice")
		}

		append_hmac := enc_content_slice_hmac[:64]
		hmac_check , err2:= userlib.HMACEval(userfile.Content_HMACs[i], enc_content_slice_hmac[64:])
		if err2 != nil {
			return nil, errors.New("could not call hmac on encrypted file content slice form")
		}
		eq := userlib.HMACEqual(append_hmac, hmac_check)
		if !eq {
			return nil, errors.New("hmac is not the same so encrypted file content has been tampered with")
		}

		dec_data_slice := userlib.SymDec([]byte ((userfile.Content_UUIDs[i]).String() + string(userfile.IV[i]))[:16], enc_content_slice_hmac[64:])

		data = append(data, dec_data_slice...)
	}

	return data, nil
}

/*
func (userdata *User) ShareFile(filename string, recipient string) (magic_string string, err error) {
	var access Access
	access, err = LoadFileInfo(userdata, filename)
	if err != nil {
		return "", nil
	}
	sign_asym_enc_file_info := access.File_information
	asym_sign := sign_asym_enc_file_info[:256]
	asym_enc_file_info := sign_asym_enc_file_info[256:]
	digit_verify, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		return "",errors.New("could not get sender's digital verify")
	}

	err = userlib.DSVerify(digit_verify, asym_enc_file_info, asym_sign)
	if err != nil {
		return "",errors.New("this was not signed by sender")
	}

	file_info_decrypt, err := userlib.PKEDec(userdata.Secret_key, asym_enc_file_info)
	if err != nil {
		return "",errors.New("could not decrypt  file information")
	}

	pub_enc_key, ok := userlib.KeystoreGet(recipient)
	if !ok  {
		errors.New("Could not get public key")
	}

	file_info_encrypt, err  := userlib.PKEEnc(pub_enc_key,file_info_decrypt)
	if err != nil {
		errors.New("Could not store file information")
	}

	file_info_sign, err := userlib.DSSign(userdata.Sign_key, file_info_encrypt)
	if err != nil {
		errors.New("Could not store file information")
	}
	//save new user state


	enc_access_struct := userlib.SymEnc(userdata.Access_shared_sym_enc[filename][:16], userlib.RandomBytes(16), []byte (string(access_struct_bytes)))
	access_location := userdata.Access_shared_location[filename]

	access_struct_sig, err := userlib.HMACEval(userdata.Access_shared_hmac_key[filename], enc_access_struct)

	if err != nil {
		//errors.New("encrypted user struct could not be signed")
		return
	}
	userlib.DatastoreSet(access_location, []byte (string(access_struct_sig) + string(enc_access_struct)))

	//public key encrypt access location and hmac key of access location
	pkekey, ok := userlib.KeystoreGet(recipient)
	if !ok {
		return "", errors.New("could not get from keystore")
	}
	encrypt_access_struct_info, err := userlib.PKEEnc(pkekey, []byte(string(userdata.Access_shared_sym_enc[filename]) + string(userdata.Access_shared_hmac_key[filename]) + access_location.String()))
	if err != nil {
		return "", errors.New("could not encrypt magic string")
	}
	dig_sign_enc_magic_string, err := userlib.DSSign(userdata.Sign_key, encrypt_access_struct_info)
	if err != nil {
		return "", errors.New("could not sign access information")
	}
	magic_string = string(dig_sign_enc_magic_string) + string(encrypt_access_struct_info)
	return magic_string, nil
}
*/

func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {
	if userdata.Uuid == uuid.Nil || userdata.Username == ""  || userdata.IV == nil{
		return "", errors.New("share user has not been initialized")
	}
	_, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		return "", errors.New("share user has not been initialized")
	}
	_, ok = userlib.KeystoreGet(userdata.Username)
	if !ok {
		return "", errors.New("share user has not been initialized")
	}
	var owner_access Access
	owner_access, err = LoadFileInfo(userdata, filename)
	if err != nil {
		return "", err
	}
	var receiver_Access Access
	receiver_access_loc := uuid.New()
	receiver_sym_enc_key := userlib.RandomBytes(16)
	receiver_hmac_key := userlib.RandomBytes(16)
	receiver_Access.Access_Structs_shared = make(map[string][]byte)

	receiver_Access.File_information = owner_access.File_information
	receiver_Access_struct_bytes, error := json.Marshal(receiver_Access)
	if error != nil {
		return "", errors.New("could not marshal usr struct properly")
	}

	enc_receiver_access_struct := userlib.SymEnc(receiver_sym_enc_key, userlib.RandomBytes(16), []byte (string(receiver_Access_struct_bytes)))

	receiver_access_struct_sig, err := userlib.HMACEval(receiver_hmac_key, enc_receiver_access_struct)
	if err != nil {
		return "", errors.New("encrypted user struct could not be signed")

	}
	userlib.DatastoreSet(receiver_access_loc, []byte (string(receiver_access_struct_sig) + string(enc_receiver_access_struct)))

	owner_access.Access_Structs_shared[recipient] = []byte(string(receiver_sym_enc_key) + string(receiver_hmac_key) + receiver_access_loc.String())

	owner_access_struct_bytes, error := json.Marshal(owner_access)
	if error != nil {
		return "", errors.New("could not marshal usr struct properly")
	}

	owner_enc_access_struct := userlib.SymEnc(userdata.Access_shared_information[filename][:16], userlib.RandomBytes(16), []byte (string(owner_access_struct_bytes)))

	owner_access_struct_sig, err := userlib.HMACEval(userdata.Access_shared_information[filename][16:32], owner_enc_access_struct)
	if err != nil {
		return "", errors.New("encrypted user struct could not be signed")

	}
	owner_access_loc, err := uuid.ParseBytes(userdata.Access_shared_information[filename][32:])
	if err != nil {
		return "", errors.New("Could not get owner location")
	}
	userlib.DatastoreSet(owner_access_loc, []byte (string(owner_access_struct_sig) + string(owner_enc_access_struct)))


	recipient_pke, ok := userlib.KeystoreGet(recipient)
	if !ok {
		return "", errors.New("error getting recipients public key")
	}
	asym_enc_content, err := userlib.PKEEnc(recipient_pke, [] byte (string(receiver_sym_enc_key) + string(receiver_hmac_key) + receiver_access_loc.String()))
	if err != nil {
		return "", errors.New("error asym encrypting magic string")
	}

	digital_sig, err1 := userlib.DSSign(userdata.Sign_key, asym_enc_content)
	if err1 != nil {
		return "", errors.New("could not digitally sign asym information")
	}

	magic_string = string(digital_sig) + string(asym_enc_content)

	return magic_string, nil


}

func (userdata *User) ReceiveFile(filename string, sender string, magic_string string) error {
	if userdata.Uuid == uuid.Nil || userdata.Username == ""  || userdata.IV == nil{
		return errors.New("receive user has not been initialized")
	}
	_, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		return errors.New("receive user has not been initialized")

	}
	_, ok = userlib.KeystoreGet(userdata.Username)
	if !ok {
		return errors.New("receive user has not been initialized")

	}
	_, ok = userdata.Access_shared_information[filename]
	if ok {
		return errors.New("Recipient alread has this file name")
	}
	if magic_string == "" {
		return errors.New("Magic String is not valid")
	}
	magic_byte := [] byte(magic_string)
	appended_ds := magic_byte[:256]
	asym_enc_content := magic_byte[256:]

	digit_verify, ok := userlib.KeystoreGet(sender + "sig")
	if !ok {
		return errors.New("could not get sender's digital verify")
	}

	err := userlib.DSVerify(digit_verify, asym_enc_content, appended_ds)
	if err != nil {
		return errors.New("this was not signed by sender")
	}

	dec_access_information, err1 := userlib.PKEDec(userdata.Secret_key, asym_enc_content)
	if err1 != nil {
		return errors.New("we could not decrypt the magic string")
	}

	userdata.Access_shared_information[filename] = dec_access_information

	usr_struct_bytes, err := json.Marshal(userdata)
	if err != nil {
		return errors.New("could not marshal usr struct properly")
	}

	enc_user_struct := userlib.SymEnc([]byte(userdata.Uuid.String() + userdata.Username)[:16], userlib.RandomBytes(16), []byte (string(usr_struct_bytes)))
	user_struct_sig, err := userlib.HMACEval(userlib.Argon2Key([]byte(userdata.Username + "sad"), []byte("salty"), 16), enc_user_struct)//userdata.Hmackey, usr_struct_bytes)
	if err != nil {
		return errors.New("encrypted user struct could not be signed")
	}
	//userlib.DebugMsg("stored struct %v", userdata.Username, enc_user_struct)
	userlib.DatastoreSet(userdata.Uuid, [] byte (string(user_struct_sig)+string(enc_user_struct)))

	return nil
}


func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	if userdata.Uuid == uuid.Nil || userdata.Username == ""  || userdata.IV == nil{
		return errors.New("revoke user has not been initialized")

	}
	_, ok := userlib.KeystoreGet(userdata.Username + "sig")
	if !ok {
		return errors.New("revoke user has not been initialized")

	}
	_, ok = userlib.KeystoreGet(userdata.Username)
	if !ok {
		return errors.New("revoke user has not been initialized")

	}
	var access Access
	access, err = LoadFileInfo(userdata, filename)
	if err != nil {
		errors.New("cannot find file ")
	}
	revoker_access_info, ok := access.Access_Structs_shared[target_username]
	if !ok {
		return errors.New("owner has not shareed with this person ")
	}
 	return DFSDeleteAccess(revoker_access_info)
}

func DFSDeleteAccess(revoker_access_info []byte) (err error){
	var access Access
	access_location_bytes := revoker_access_info[32:]
	access_location, err := uuid.ParseBytes(access_location_bytes)
	if(err != nil) {
		return errors.New("Nothing stored at this access location")
	}

	enc_file_access_hmac, ok := userlib.DatastoreGet(access_location)
	if !ok {
		return errors.New("You no longer have access")
	}

	append_hmac := enc_file_access_hmac[:64]
	hmac_check , err:= userlib.HMACEval(revoker_access_info[16:32], enc_file_access_hmac[64:])
	if err != nil {
		return errors.New("could not call hmac on access encrypted struct form")
	}

	eq := userlib.HMACEqual(append_hmac, hmac_check)
	if !eq {
		userlib.DebugMsg("incorrect in access struct ")
		return errors.New("hmac is not the same so file struct has been tampered with?")
	}
	dec_access_struct := userlib.SymDec(revoker_access_info[:16], enc_file_access_hmac[64:])
	err1 := json.Unmarshal(dec_access_struct, &access)

	if err1 != nil {
		return errors.New("could not unmarshal file struct")
	}
	for _, recipient_access_info := range access.Access_Structs_shared {
		err = DFSDeleteAccess(recipient_access_info)
		if err != nil {
			errors.New("server data modifierd could not revoke")
		}
	}
	access_loc, err := uuid.ParseBytes(revoker_access_info[32:])
	if err != nil {
		return errors.New("could not parse location")
	}
	userlib.DatastoreDelete(access_loc)
	return nil
}
