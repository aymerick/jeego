package main

// Base node
type Node struct {
	Id   byte
	Name string
}

// Node interface
type INode interface {
	handleData(data []byte)
	dumpData()
}
