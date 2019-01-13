package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

func IntToByte(num int64)[]byte  {
	var buffer bytes.Buffer
	err := binary.Write(&buffer, binary.BigEndian, num)
	CheckErr(err, "int64转换切片失败")
	return buffer.Bytes();
}

func CheckErr(err error, msg string)  {
	if err !=nil{
		fmt.Println(msg,", err=", err)
		os.Exit(1)
	}
}
