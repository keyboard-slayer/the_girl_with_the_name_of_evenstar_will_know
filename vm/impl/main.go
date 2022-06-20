package main

import "fmt"

type Pair struct {
	a, b interface{}
}

type Reader struct {
	buf    []uint8
	offset int
}

func (r Reader) peek() uint8 {
	return r.buf[r.offset]
}

func (r *Reader) next() uint8 {
	out := r.peek()
	r.offset++
	return out
}

func (r Reader) isEof() bool {
	return r.offset == len(r.buf)
}

type Cpu struct {
	stack []uint8
	regs  []uint64
	pc    uint64
	sp    int
}

func (cpu *Cpu) push(r Reader) {
	cpu.stack[cpu.sp] = 0x42
	cpu.sp -= 1

	for r.peek() != 0x42 {
		cpu.stack[cpu.sp] = r.peek()
		cpu.sp -= 1
		r.next()
	}
}

func revertBytes(bytes []uint8) []uint8 {
	if len(bytes) == 0 {
		return bytes
	}
	return append(revertBytes(bytes[1:]), bytes[0])
}

func (cpu *Cpu) pop(r Reader) []uint8 {
	var data []uint8

	for {
		b := cpu.stack[cpu.sp+1]
		cpu.sp += 1

		if b == 0x42 {
			break
		} else {
			data = append(data, b^0x42)
		}
	}

	return revertBytes(data)
}

var (
	INIT  uint8 = 0x00
	PUSH  uint8 = 0x01
	POP   uint8 = 0x02
	PRINT uint8 = 0x03
	OPEN  uint8 = 0x05
	WRITE uint8 = 0x08
	READ  uint8 = 0x0d
	FSIZE uint8 = 0x15
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
			// UNUSED
		case PUSH:
			cpu.push(r)
		case PRINT:
			fmt.Println(string(cpu.pop(r)))
		}
	}
}
