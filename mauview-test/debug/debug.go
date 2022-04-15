// gomuks - A terminal Matrix client written in Go.
// Copyright (C) 2018 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package debug

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"
)

var writer io.Writer

func init() {
	var err error
	writer, err = os.OpenFile("/tmp/mauview-debug.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
}

func Printf(text string, args ...interface{}) {
	if writer != nil {
		fmt.Fprintf(writer, time.Now().Format("[2006-01-02 15:04:05] "))
		fmt.Fprintf(writer, text+"\n", args...)
	}
}

func Print(text ...interface{}) {
	if writer != nil {
		fmt.Fprintf(writer, time.Now().Format("[2006-01-02 15:04:05] "))
		fmt.Fprintln(writer, text...)
	}
}

func PrintStack() {
	if writer != nil {
		data := debug.Stack()
		writer.Write(data)
	}
}
