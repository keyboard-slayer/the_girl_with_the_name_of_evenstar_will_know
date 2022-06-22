package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	XOR = 0x42
)

type Pair struct {
	a, b interface{}
}

type Reader struct {
	buf    []byte
	offset int
}

func (r Reader) peek() byte {
	return r.buf[r.offset]
}

func (r *Reader) next() byte {
	out := r.peek()
	r.offset++
	return out
}

func (r Reader) isEof() bool {
	return r.offset == len(r.buf)
}

type Cpu struct {
	stack []byte
	regs  []interface{}
	sp    int
}

func (cpu *Cpu) push(r *Reader) {
	cpu.stack[cpu.sp] = XOR
	cpu.sp -= 1

	for r.peek() != XOR {
		cpu.stack[cpu.sp] = r.peek()
		cpu.sp -= 1
		r.next()
	}
}

func revertBytes(bytes []byte) []byte {
	if len(bytes) == 0 {
		return bytes
	}
	return append(revertBytes(bytes[1:]), bytes[0])
}

func (cpu *Cpu) pop() []byte {
	var data []byte

	for {
		b := cpu.stack[cpu.sp+1]
		cpu.sp += 1

		if b == XOR {
			break
		} else {
			data = append(data, b^XOR)
		}
	}

	return revertBytes(data)
}

func (cpu *Cpu) open(r *Reader) {
	var mode string
	var err error

	filename := string(cpu.pop())
	reg := 0xff - r.next()

	for {
		b := r.next()

		if b == XOR {
			break
		}

		mode += string(b ^ XOR)
	}

	if strings.HasPrefix(filename, "~") {
		usr, _ := user.Current()
		filename = filepath.Join(usr.HomeDir, filename[2:])
	}

	if mode == "r" {
		cpu.regs[reg], err = os.OpenFile(filename, os.O_RDONLY, 0755)
	} else if mode == "w" {
		cpu.regs[reg], err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0755)
	} else if mode == "rw" {
		cpu.regs[reg], err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
	} else if mode == "ra" {
		cpu.regs[reg], err = os.OpenFile(filename, os.O_APPEND|os.O_RDONLY, 0755)
	} else if mode == "a" {
		cpu.regs[reg], err = os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		fmt.Println(mode)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cpu *Cpu) fsize(r *Reader) {
	src := 0xff - r.next()
	dst := 0xff - r.next()

	file := cpu.regs[src].(fs.File)
	info, _ := file.Stat()

	cpu.regs[dst] = info.Size()
}

func (cpu *Cpu) imm_push(data []uint8) {
	cpu.stack[cpu.sp] = XOR
	cpu.sp -= 1

	for _, b := range data {
		cpu.stack[cpu.sp] = b ^ XOR
		cpu.sp -= 1
	}
}

func (cpu *Cpu) read(r *Reader) {
	src := 0xff - r.next()
	size := 0xff - r.next()

	file := cpu.regs[src].(*os.File)
	data := make([]byte, cpu.regs[size].(int64))

	file.Read(data)
	cpu.imm_push(data)
}

func (cpu *Cpu) write(r *Reader) {
	dst := 0xff - r.next()
	file := cpu.regs[dst].(*os.File)
	data := cpu.pop()
	_, err := file.Write(data)

	if err != nil {
		fmt.Println(err)
	}
}

func (cpu *Cpu) pop_instr(r *Reader) {
	dst := 0xff - r.next()
	cpu.regs[dst] = cpu.pop()
}

func (cpu *Cpu) system(r *Reader) {
	args := string(cpu.pop())
	cmd := exec.Command("bash", "-c", args)
	stdout, err := cmd.Output()

	if err != nil {
		os.Exit(1)
	}

	cpu.imm_push([]byte(strings.TrimSuffix(string(stdout), "\n")))
}

func (cpu *Cpu) append(r *Reader) {
	register := 0xff - r.next()
	var str []byte = cpu.regs[register].([]byte)
	var to_append []byte = cpu.pop()

	for _, b := range to_append {
		str = append(str, b)
	}

	cpu.imm_push(str)
}

func (cpu *Cpu) pushreg(r *Reader) {
	register := 0xff - r.next()
	cpu.imm_push(cpu.regs[register].([]byte))
}

func (cpu *Cpu) mkdir(r *Reader) {
	dir := string(cpu.pop())
	os.MkdirAll(dir, 0755)
}

var (
	INIT    byte = 0x00
	PUSH    byte = 0x01
	POP     byte = 0x02
	PRINT   byte = 0x03
	OPEN    byte = 0x05
	WRITE   byte = 0x08
	READ    byte = 0x0d
	FSIZE   byte = 0x15
	SYSTEM  byte = 0x22
	PUSHREG byte = 0x37
	APPEND  byte = 0x59
	MKDIR   byte = 0x90
)

func main() {
	cpu := Cpu{}

	cpu.sp = 0x1000

	for i := 0; i < cpu.sp+1; i++ {
		cpu.stack = append(cpu.stack, 0)
	}

	r := Reader{Bytecode, 0}
	for !r.isEof() {
		switch r.next() {
		case INIT:
			n, _ := strconv.Atoi(string(cpu.pop()[0]))

			for i := 0; i < n; i++ {
				cpu.regs = append(cpu.regs, 0)
			}
		case PUSH:
			cpu.push(&r)
		case POP:
			cpu.pop_instr(&r)
		case PRINT:
			fmt.Println(string(cpu.pop()))
		case OPEN:
			cpu.open(&r)
		case FSIZE:
			cpu.fsize(&r)
		case READ:
			cpu.read(&r)
		case WRITE:
			cpu.write(&r)
		case SYSTEM:
			cpu.system(&r)
		case APPEND:
			cpu.append(&r)
		case PUSHREG:
			cpu.pushreg(&r)
		case MKDIR:
			cpu.mkdir(&r)
		}
	}
}
