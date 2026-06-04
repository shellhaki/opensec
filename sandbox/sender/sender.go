package main


import (

	"fmt"
	"net"
	"os"
	"encoding/binary"
	"io"
)

func main(){
	filePath := "/home/haki/opensec/sandbox/sender/test.txt"

	conn,err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	file,err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	info,_ := file.Stat()
	filename := []byte(info.Name())
	// sendfilename length
	binary.Write(conn,binary.BigEndian,uint32(len(filename)))

	//send file name
	conn.Write(filename)

	//send file size
	binary.Write(conn,binary.BigEndian,info.Size())

	//then finally send file data
	_,err = io.Copy(conn,file)
	if err != nil {
		panic(err)
	}
	fmt.Println("sent files")
}