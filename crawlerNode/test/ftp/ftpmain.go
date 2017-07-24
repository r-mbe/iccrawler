package main

import (
	"fmt"
	//"io/ioutil"
	"os"
	"time"

	ftp "github.com/jlaffaye/ftp"
)

func main() {

	start := time.Now()
	ftpUploadFile("feed.data.ickey.cn:21", "anglia-live", "c6LfZthyVBy45tiB", "2017.csv", "/", "2017.csv")
	dur := time.Since(start).Seconds()
	fmt.Printf("Upload file spend time: %v  seconds\n", dur)

}

func ftpUploadFile(ftpserver, ftpuser, pw, localFile, remoteSavePath, saveName string) {
	ftp, err := ftp.Connect(ftpserver)
	if err != nil {
		fmt.Println(err)
	}
	err = ftp.Login(ftpuser, pw)
	if err != nil {
		fmt.Println(err)
	}
	//注意是 pub/log，不能带“/”开头
	ftp.ChangeDir("pub/log")
	dir, err := ftp.CurrentDir()
	fmt.Println(dir)
	ftp.MakeDir(remoteSavePath)
	ftp.ChangeDir(remoteSavePath)
	dir, _ = ftp.CurrentDir()
	fmt.Println(dir)
	file, err := os.Open(localFile)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	err = ftp.Stor(saveName, file)
	if err != nil {
		fmt.Println(err)
	}
	ftp.Logout()
	ftp.Quit()
	fmt.Println("success upload file:", localFile)
}
