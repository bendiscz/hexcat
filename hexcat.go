// Copyright (c) 2019 Martin Benda
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const (
	exitUsage = 64
	exitIOErr = 74
)

func main() {
	if len(os.Args) != 2 {
		_, _ = fmt.Fprintf(os.Stderr, "usage: %v <address>\n", os.Args[0])
		os.Exit(exitUsage)
	}

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		exit(err)
	}

	go func() {
		var buf [1024]byte
		for {
			n, err := conn.Read(buf[:])
			if err == io.EOF {
				_ = conn.Close()
				os.Exit(0)
			}

			if err != nil {
				exit(err)
			}

			fmt.Print(encode(buf[:n]))
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			_ = conn.Close()
			break
		}

		if err != nil {
			exit(err)
		}

		_, err = conn.Write(decode(line))
		if err != nil {
			exit(err)
		}
	}
}

func exit(err error) {
	_, _ = fmt.Fprintln(os.Stderr, "error: ", err)
	os.Exit(exitIOErr)
}

func encode(buf []byte) string {
	const block = 8
	var out bytes.Buffer

	for i := 0; i < len(buf); {
		s := i + block
		if s > len(buf) {
			s = len(buf)
		}

		out.WriteString(fmt.Sprintf("%04x  ", i))

		for j := i; j < s; j++ {
			out.WriteString(fmt.Sprintf("%02x ", buf[j]))
		}

		for j := s - i; j < block; j++ {
			out.WriteString("   ")
		}

		out.WriteString(" |")

		for j := i; j < s; j++ {
			ch := buf[j]
			if ch >= 32 && ch < 128 {
				out.WriteByte(ch)
			} else {
				out.WriteByte('.')
			}
		}

		out.WriteString("|\n")

		i = s
	}

	return out.String()
}

func decode(line string) []byte {
	data, err := hex.DecodeString(strings.TrimSpace(line))
	if err != nil {
		log.Println(err)
	}

	return data
}
