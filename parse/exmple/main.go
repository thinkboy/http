package main

import (
	"fmt"
	"github.com/thinkboy/http/parse"
	"net/url"
)

type Param struct {
	Num  string
	Num1 int    `param:"omitempty,range[1:]"`
	Num2 string `param:"range[3]"`
	Num3 []byte `param:"range[16]"`
}

func main() {
	pa := Param{}

	val := make(url.Values)
	val.Set("num", "11")
	val.Set("num1", "22")
	val.Set("num2", "223")
	val.Set("num3", "e8313ee1b99d4f96b5451a0209b1099e")

	err := parse.ParseUrlParam(val, &pa)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(pa)
}
