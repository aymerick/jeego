package main

type Node struct {
	Id   byte
	Name string
}

type INode interface {
	handleData(data []byte)
	dump()
}
