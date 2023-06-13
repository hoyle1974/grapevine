package main

import (
	"fmt"

	"github.com/rivo/tview"
)

type LogAdapter struct {
	out *tview.TextView
}

func (l LogAdapter) Write(p []byte) (n int, err error) {

	batch := l.out.BatchWriter()
	defer batch.Close()

	fmt.Fprint(batch, string(p))

	return len(p), nil
}
