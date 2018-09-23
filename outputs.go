package main

import (
	"fmt"
	"io"

	"github.com/davecgh/go-spew/spew"
)

type Outputter interface {
	Print(nodes []*RedisInstance)
}

type debugOutputter struct {
	writer io.Writer
}

func newDebugOutputter(writer io.Writer) Outputter {
	return &debugOutputter{
		writer: writer,
	}
}

func (d *debugOutputter) Print(nodes []*RedisInstance) {
	spew.Dump(nodes)
}

type graphvizOutputter struct {
	writer io.Writer
}

func newGraphvizOutputter(writer io.Writer) Outputter {
	return &graphvizOutputter{
		writer: writer,
	}
}

func (d *graphvizOutputter) Print(nodes []*RedisInstance) {
	d.writer.Write([]byte("digraph redis {\n"))

	for _, node := range nodes {
		switch node.Role {
		case Master:
			fmt.Fprintf(d.writer, "\t%q[color=red];\n", node.Address)
		case Slave:
			fmt.Fprintf(d.writer, "\t%q[color=yellow];\n", node.Address)
		case Sentinel:
			fmt.Fprintf(d.writer, "\t%q[color=blue];\n", node.Address)
		}

		if node.Master != nil {
			fmt.Fprintf(d.writer, "\t%q -> %q[label=\"REPLICAOF\"];\n", node.Address, node.Master.Address)
		}

		for _, sentinelNode := range node.SentinelMasters {
			fmt.Fprintf(d.writer, "\t%q -> %q[label=\"WATCHES\"];\n", node.Address, sentinelNode.Address)
		}
	}

	d.writer.Write([]byte("}\n"))
}
