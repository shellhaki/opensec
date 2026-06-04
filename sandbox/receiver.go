package main

import (
	"net"
	"encoding/binary"
	"io"
)
/*
receiver opens tcp socket and listens for senders
 */

func main(){
	listener,err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}

	// accept connections with listener.Accept()
	conn,err := listener.Accept()
	if err != nil {
		panc
	}
	defer conn.Close()
	// declare variable to hold the value of the filename length to store the file name as bytes

	var namelen uint32
	binary.Read(conn,binary.BigEndian,&namelen)

	// now construct or obtain the file name
	nameBuf := make([]byte,namelen)
	// this creates a buffer that will be used to store the file name with the appropriate file size
	// read the full file name next
	io.ReadFull(conn,nameBuf)
	filename := string(nameBuf)

	// create a variable to store file size of type int64
	var fileSize int64

	// use binary read to write the file size into the variable
	binary.Read(conn,binary.BigEndian,&fileSize)

	//then create the file with os.create
	out,err := os.Create("recv_" + filename)
	if err != nil{
		panic(err)
	}
	defer out.Close()
	fmt.Println("receiving a file...")

	// receive file in chunks
	written,err := io.copyN(out,conn,fileSize)
	if err != nil && err != io.EOF {
		panic(err)
	}
	fmt.Println("done, bytes: " + written)
}
