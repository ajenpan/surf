package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func DirExists(path string) (bool, error) {
	s, err := os.Stat(path)
	if err == nil {
		return s.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func FileExist(path string) (bool, error) {
	s, err := os.Stat(path)
	if err == nil {
		return !s.IsDir(), nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func PresistDir(path string) error {
	if exist, err := DirExists(path); err != nil {
		return err
	} else if exist {
		return nil
	}
	return os.MkdirAll(path, os.ModePerm)
}

func Sha1Bytes(raw []byte) string {
	hashInBytes := sha1.Sum(raw)
	return hex.EncodeToString(hashInBytes[:])
}

func Sha1File(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnSHA1String string

	//Open the filepath passed by the argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnSHA1String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new SHA1 hash interface to write to
	hash := sha1.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnSHA1String, err
	}

	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]

	//Convert the bytes to a string
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil
}

func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func MD5(v []byte) string {
	temp := md5.Sum(v)
	return hex.EncodeToString(temp[:])
}

func JoinURL(base string, paths ...string) string {
	p := path.Join(paths...)
	return fmt.Sprintf("%s/%s", strings.TrimRight(base, "/"), strings.TrimLeft(p, "/"))
}

func RemoveFile(dst string) error {
	if b, err := FileExist(dst); err != nil {
		return err
	} else if b {
		return os.Remove(dst)
	}
	return nil
}
