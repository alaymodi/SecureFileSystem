package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	_ "encoding/hex"
	_ "encoding/json"
	_ "errors"
	"github.com/cs161-staff/userlib"
	"github.com/google/uuid"
	"reflect"
	_ "reflect"
	_ "strconv"
	_ "strings"
	"testing"
)


func TestInit(t *testing.T) {
	t.Log("Initialization TEST test")
	t.Log("EHLLOOOE")

	// You may want to turn it off someday
	userlib.SetDebugStatus(true)
	//someUsefulThings()  //  Don't call someUsefulThings() in the autograder in case a student removes it
	// userlib.SetDebugStatus(false)
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	//_,err = InitUser("alice", "fubar")
	//if err != nil {
	//	t.Error("Failed to initialize user", err)
	//	return
	//}
	// t.Log() only produces output if you run with "go test -v"
	//t.Log("Got user", u)
	// If you want to comment the line above,
	// write _ = u here to make the compiler happy
	// You probably want many more tests here.
}

func TestStorage(t *testing.T) {
	// And some more tests, because
	u, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	//t.Log("Loaded user", u)

	_, err = GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	//t.Log("Loaded user", u2)
	v := []byte("This is a test")
	u.StoreFile("file1", v)

	//t.Log("well ??", u)

	err4 := u.AppendFile("file1", []byte("asldkfjlaksd"))
	t.Log(err4)

	err4 = u.AppendFile("file1", []byte("s"))
	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	t.Log("v2!!", string(v2))
	v3 := []byte("This is a testasldkfjlaksds")
	if !reflect.DeepEqual(v3, v2) {
		t.Error("Downloaded file is not the same", v3, v2)
		return
	}


}

func TestShare(t *testing.T) {
	u, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	//t.Log("hi", u)
	u2, err2 := InitUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	//t.Log("hi", u2)

	var v , v2 []byte
	var magic_string string

	v, err = u.LoadFile("file1")
	if err != nil {
		t.Error("Failed to download the file from alice\n", err)
		return
	}
	t.Log("vvhfsssds", v)

	magic_string, err = u.ShareFile("file1", "bob")
	if err != nil {
		t.Error("Failed to share the a file", err)
		return
	}
	t.Log("magic!! ", []byte(magic_string))
	err = u2.ReceiveFile("file2", "alice", magic_string)
	if err != nil {
		t.Error("Failed to receive the share message", err)
		return
	}
//
	v2, err = u2.LoadFile("file2")
	if err != nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	t.Log("wait actually", string(v))
	t.Log("but really tho..", string(v2))
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
}

func TestRevoke(t *testing.T) {
	u, err := GetUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to reload user", err)
		return
	}
	v,err := u.LoadFile("file1")
	u2, err2 := GetUser("bob", "foobar")
	if err2 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	u3, err3 := InitUser("charlie", "foobar")

	if err3 != nil {
		t.Error("Failed to initialize bob", err2)
		return
	}
	u4, err4 := InitUser("connor", "foobar")

	if err4 != nil {
		t.Error("Failed to initialize bob", err3)
		return
	}
	magic_string, err := u.ShareFile("file1", "connor")
	u4.ReceiveFile("file4", "alice", magic_string)
	magic_string, err = u2.ShareFile("file2", "charlie")
	u3.ReceiveFile("file3", "bob", magic_string)
	u.RevokeFile("file1", "bob")
	v2, err := u.LoadFile("file1")
	_, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("user access not revoked", err)
	}
	_, err = u3.LoadFile("file3")
	if err == nil {
		t.Error("Failed to download the file after sharing", err)
		return
	}
	if !reflect.DeepEqual(v, v2) {
		t.Error("Shared file is not the same", v, v2)
		return
	}
	v4, err := u4.LoadFile("file4")
	if !reflect.DeepEqual(v, v4) {
		t.Error("Shared file is not the same", v, v2)
		return
	}

}

func TestUser_AppendFile2(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	magic_string, err2 := u.ShareFile("file1", "bob")
	if err2 != nil {
		t.Error("Failed to share the a file", err2)
		return
	}
	err3 := u1.ReceiveFile("file2", "alice", magic_string)
	if err3 != nil {
		t.Error("Receive the share message when you shouldn't since bob did not sign it" , err3)
		return
	}

	v2 := []byte("yes a test")
	err2 = u1.AppendFile("file2", v2)
	if err2 != nil {
		t.Error("Failed to append to file1", err2)
	}

	v3, err3 := u.LoadFile("file1")
	if err3 != nil {
		t.Error("Failed to upload and download", err3)
		return
	}
	v4 := append(v, v2...)
	if !reflect.DeepEqual(v3, v4) {
		t.Error("Downloaded file is not the same", v3, v4)
		return
	}
}

func TestInitUser2(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
	}
	t.Log("Create user alice", u)
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err2 := GetUser("alice", "fubar")
	if err2 == nil {
		t.Error("Created when it shouldn't exist ", err2)
	}
}

// see if it errors properly when loading a file that doesn't exist
func TestLoading_Invalid(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
	}
	t.Log("Create user alice", u)

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v1, err2 := u.LoadFile("file")
	if err2 == nil {
		t.Log("loaded a file when it shouldn't have", v1)
	}
}

func TestUser_AppendFile(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2 := []byte("yes a test")
	err2 := u.AppendFile("file1", v2)
	if err2 != nil {
		t.Error("Failed to append to file1", err2)
	}

	v3, err3 := u.LoadFile("file1")
	if err3 != nil {
		t.Error("Failed to upload and download", err3)
		return
	}
	v4 := append(v, v2...)
	if !reflect.DeepEqual(v3, v4) {
		t.Error("Downloaded file is not the same", v3, v4)
		return
	}
}

func TestStoreOveride(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)

	v := []byte("This is a test")
	u.StoreFile("file1", v)

	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	z := [] byte("This is not a test")
	u.StoreFile("file1", z)

	v3, err3 := u.LoadFile("file1")
	if err3 != nil {
		t.Error("Failed to upload and download", err3)
		return
	}
	if reflect.DeepEqual(v2, v3) {
		t.Error("Downloaded file is not the same", v2, v3)
		return
	}
}

//makes sure revoke works properly
func TestUser_RevokeFile(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	u2, err2 := InitUser("james", "barfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string2, err5 := u1.ShareFile("file2", "james")
	if err5 != nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	err6 := u2.ReceiveFile("file3", "bob", magic_string2)
	if err6!= nil {
		t.Error("Failed to receive the share message", err6)
		return
	}

	err7 := u.RevokeFile("file1", "bob")
	if err7 != nil {
		t.Error("Failed to revoke", err7)
		return
	}

	v, err8 := u2.LoadFile("file3")
	if err8 == nil {
		t.Error("james still has access when he shouldn't", err8)
		return
	}
	t.Log("Successfully revoked", v)
}

//// checks that anyone that is not the owner cannot revoke the file
//func TestUser_RevokeFile2(t *testing.T) {
//	userlib.KeystoreClear()
//	userlib.DatastoreClear()
//	u, err := InitUser("alice", "fubar")
//	if err != nil {
//		// t.Error says the test fails
//		t.Error("Failed to initialize user", err)
//		return
//	}
//
//	u1, err1 := InitUser("bob", "foobar")
//	if err1 != nil {
//		// t.Error says the test fails
//		t.Error("Failed to initialize user", err1)
//		return
//	}
//
//	v := []byte("This is a test")
//	u.StoreFile("file1", v)
//
//	magic_string, err2 := u.ShareFile("file1", "bob")
//	if err2 != nil {
//		t.Error("Failed to share the a file", err2)
//		return
//	}
//	err3 := u1.ReceiveFile("file2", "alice", magic_string)
//	if err3 != nil {
//		t.Error("Failed to receive the share message", err3)
//		return
//	}
//
//	err4 := u1.RevokeFile("file2", "alice")
//	if err4 == nil {
//		t.Error("revoke when we should not be able to as bob is not the owner", err4)
//		return
//	}
//}

// ensure's receiving from a different sender errors
func TestUser_ReceiveFile2(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	magic_string, err2 := u.ShareFile("file1", "bob")
	if err2 != nil {
		t.Error("Failed to share the a file", err2)
		return
	}
	err3 := u1.ReceiveFile("file2", "alic", magic_string)
	if err3 == nil {
		t.Error("Receive the share message when you shouldn't since alic does not exist" , err3)
		return
	}
}

// test that it fails because we verify with another user that exist but did not sign it
func TestUser_ReceiveFile3(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	magic_string, err2 := u.ShareFile("file1", "bob")
	if err2 != nil {
		t.Error("Failed to share the a file", err2)
		return
	}
	err3 := u1.ReceiveFile("file2", "bob", magic_string)
	if err3 == nil {
		t.Error("Receive the share message when you shouldn't since bob did not sign it" , err3)
		return
	}
}

func TestUser_RevokeFile3(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	u2, err2 := InitUser("james", "barfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string2, err5 := u1.ShareFile("file2", "james")
	if err5 != nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	err6 := u2.ReceiveFile("file3", "bob", magic_string2)
	if err6!= nil {
		t.Error("Failed to receive the share message", err6)
		return
	}

	err7 := u.RevokeFile("file1", "bob")
	if err7 != nil {
		t.Error("Failed to revoke", err7)
		return
	}

	magic_string3, err9 := u1.ShareFile("file2", "james")
	if err9 == nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	t.Log("magic string ", magic_string3)
}

func TestUser_RevokeFile4(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	u2, err2 := InitUser("james", "barfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string2, err5 := u1.ShareFile("file2", "james")
	if err5 != nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	err6 := u2.ReceiveFile("file3", "bob", magic_string2)
	if err6!= nil {
		t.Error("Failed to receive the share message", err6)
		return
	}

	err7 := u.RevokeFile("file0", "bob")
	if err7 == nil {
		t.Error("Revoke a file that doesnt exist", err7)
		return
	}
}

func TestUser_RevokeFile5(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	u2, err2 := InitUser("james", "barfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string2, err5 := u1.ShareFile("file2", "james")
	if err5 != nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	err6 := u2.ReceiveFile("file3", "bob", magic_string2)
	if err6!= nil {
		t.Error("Failed to receive the share message", err6)
		return
	}

	err7 := u.RevokeFile("file1", "bobs")
	if err7 == nil {
		t.Error("Revoke a file that doesnt exist", err7)
		return
	}
}

// 19 change share a file that doesnt exist
func TestUser_RevokeFile6(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	u2, err2 := InitUser("james", "barfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string2, err5 := u1.ShareFile("file3", "james")
	if err5 == nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	err6 := u2.ReceiveFile("file3", "bob", magic_string2)
	if err6 == nil {
		t.Error("Failed to receive the share message", err6)
		return
	}

	err7 := u.RevokeFile("file1", "bob")
	if err7 != nil {
		t.Error("Failed to revoke", err7)
		return
	}

	v, err8 := u2.LoadFile("file3")
	if err8 == nil {
		t.Error("james still has access when he shouldn't", err8)
		return
	}
	t.Log("Successfully revoked", v)
}

func TestUser_RevokeFile7(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bo")
	if err3 == nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	t.Log("magic", magic_string)

}

// store to a user that doesnt exist any longer didnt help at all
func TestStore1(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)

	userlib.KeystoreClear()
	userlib.DatastoreClear()
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	v2, err2 := u.LoadFile("file1")
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	t.Log("idk", v2)
}

//	recipient has a file with that name now try and receive should throw an error
func TestUser_ReceiveFile10(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	s := [] byte("pls end this misery")
	u1.StoreFile("file1", s)
	magic_string, err2 := u.ShareFile("file1", "bob")
	if err2 != nil {
		t.Error("we should be erroring", err2)
		return
	}
	err3 := u1.ReceiveFile("file1", "alice", magic_string)
	if err3 == nil {
		t.Error("Receive the share message when you shouldn't since alic does not exist" , err3)
		return
	}
}
// share file when u revoke access
func TestUser_RevokeFile8(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}
	u.RevokeFile("file1", "bob")
	_, err5 := u1.ShareFile("file3", "james")
	if err5 == nil {
		t.Error("Failed to share the a file", err5)
		return
	}

}

// two users same password
func TestInit1(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err := InitUser("alic", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	if reflect.DeepEqual(u, u1) {
		t.Error("should not be same", u, u1)
		return
	}
}

//append to file that doesn't exist
func TestUser_AppendFile3(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)

	v2 := []byte("yes a test")
	err2 := u.AppendFile("file1", v2)
	if err2 == nil {
		t.Error(" appending to file1 which doesnt exist", err2)
	}
}


func TestUser_AppendFile4(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)
	u2, err := InitUser("bob", "foobar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u2.StoreFile("file2", v)
	v2 := []byte("yes a test")

	err2 := u2.AppendFile("file1", v2)
	if err2 == nil {
		t.Error(" appending to file1 which you have noaccess", err2)
	}
}

func TestInitUser8(t *testing.T) {
	u, err := InitUser("alice", "fubar")
	if err != nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)
	userlib.DatastoreClear()
	u, err = InitUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to create user", err)
		return
	}
	t.Log("Loaded user", u)
}

func TestUser_RevokeFile13(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	u1, err1 := InitUser("bob", "foobar")
	if err1 != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err1)
		return
	}

	u2, err2 := InitUser("james", "barfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	u3, err2 := InitUser("connor", "jarfoo")
	if err2 != nil {
		t.Error("Failed to initialize user", err2)
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	u.AppendFile("file1", []byte(" work please"))
	magic_string, err3 := u.ShareFile("file1", "bob")
	if err3 != nil {
		t.Error("Failed to share the a file", err3)
		return
	}
	err4 := u1.ReceiveFile("file2", "alice", magic_string)
	if err4!= nil {
		t.Error("Failed to receive the share message", err)
		return
	}

	magic_string2, err5 := u1.ShareFile("file2", "james")
	if err5 != nil {
		t.Error("Failed to share the a file", err5)
		return
	}
	err6 := u2.ReceiveFile("file3", "bob", magic_string2)
	if err6!= nil {
		t.Error("Failed to receive the share message", err6)
		return
	}
	magic_string2, err10 := u.ShareFile("file1", "connor")
	if err10 != nil {
		t.Error("Failed to share the a file", err10)
		return
	}
	err11 := u3.ReceiveFile("file4", "alice", magic_string2)
	if err11!= nil {
		t.Error("Failed to receive the share message", err11)
		return
	}

	v4,err12 := u3.LoadFile("file4")
	v = []byte("This is a test work please")

	if err12 != nil {
		t.Error("Connor should be able to load file ", err12)
	}
	if !reflect.DeepEqual(v, v4) {
		t.Error("Downloaded file is not the same", v, v4)
		return
	}
	err7 := u.RevokeFile("file1", "bob")
	if err7 != nil {
		t.Error("Failed to revoke", err7)
		return
	}

	v, err8 := u2.LoadFile("file3")

	if err8 == nil {
		t.Error("james still has access when he shouldn't", err8)
		return
	}
	t.Log("Successfully revoked", v)
}
// corrupt database and get user
func TestInit21(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()

	// You may want to turn it off someday
	userlib.SetDebugStatus(true)
	//someUsefulThings()  //  Don't call someUsefulThings() in the autograder in case a student removes it
	// userlib.SetDebugStatus(false)
	_, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	usr_argon2key := userlib.Argon2Key([]byte("alice" + "fubar"), []byte("nosalt"), 16)
	usr_uuid, err := uuid.FromBytes(usr_argon2key)
	datastore := userlib.DatastoreGetMap()
	datastore[usr_uuid] = [] byte ("1231")
	_, err = GetUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to reload user", err)
		return
	}
}

func TestInit22(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()

	// You may want to turn it off someday
	userlib.SetDebugStatus(true)
	//someUsefulThings()  //  Don't call someUsefulThings() in the autograder in case a student removes it
	// userlib.SetDebugStatus(false)
	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)
	usr_argon2key := userlib.Argon2Key([]byte("alice" + "fubar"), []byte("nosalt"), 16)
	usr_uuid, err := uuid.FromBytes(usr_argon2key)
	datastore := userlib.DatastoreGetMap()
	datastore[usr_uuid] = [] byte ("1231")
	_, err = GetUser("alice", "fubar")
	if err == nil {
		t.Error("Failed to reload user", err)
		return
	}

}

func TestR1(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()
	var usr User
	v := [] byte ("idk")
	usr.StoreFile("file", v)
	usr.LoadFile("file")
}

func TestInit23(t *testing.T) {
	userlib.KeystoreClear()
	userlib.DatastoreClear()

	u, err := InitUser("alice", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	vs := []byte("This is a test")
	u.StoreFile("file1", vs)
	datastore := userlib.DatastoreGetMap()

	for k, v := range datastore {
		t.Log("fjkhdsafhsdfadsfhkfgk hi ", k, v)
		datastore[k] = v[:len(v) - 1]
	}
	_, err10:= u.LoadFile("file1")
	if err10 == nil {
		t.Error("error load file should be corrupted")
		return
	}
}