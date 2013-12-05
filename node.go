package main

// Node interface
type INode interface {
	// Node name
	name() string

    // Idx to use when sending value to domoticz
    domoticzIdx() string

	// Handle data received from that node
	handleData(data []byte)

    // Dump node data to STDOUT
	dumpData()

    // Dump node data as plain text in a string
    textData() string

    // Value that can be sent to domoticz
    domoticzValue() string
}

// Base node
type Node struct {
	Id          byte
	Name        string
    DomoticzIdx string
}

func (node *Node) name() string {
	return node.Name;
}

func (node *Node) domoticzIdx() string {
	return node.DomoticzIdx;
}
