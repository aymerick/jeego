package main

// Base node
type Node struct {
	Id   byte
	Name string
}

// Node interface
type INode interface {
    // Handle data received from that node
	handleData(data []byte)

    // Dump node data to STDOUT
	dumpData()

    // Dump node data as plain text in a string
    textData() string
}
