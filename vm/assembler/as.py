#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys

_OPCODES = {
    "PUSH":     (0x01, 1),
    "POP":      (0x02, 1),
    "PRINT":    (0x03, 0),
    "OPEN":     (0x05, 3),
    "WRITE":    (0x08, 1),
    "READ":     (0x0d, 2),
    "FSIZE":    (0x15, 1)
}


def obfuscate_string(s: str) -> list[int]:
    new_str = []

    for letter in s:
        new_str.append(ord(letter) ^ 0x42)

    new_str.append(0 ^ 0x42)
    return new_str


def tokenize(inst: str) -> list[str]:
    result: list[str] = []
    tmp: str = ""
    in_str: bool = False
    
    for token in inst.split():
        if token[0] == '\"' or in_str:
            if token[-1] == '\"':
                result.append(obfuscate_string(tmp[1:]+token[:-1]))
                tmp = ""
                in_str = False
            else:
                in_str = True
                tmp += token + " "
        elif token[0] == "r":
            result.append(0xff - int(token[1:]))
        else:
            result.append(token)
    
    return result

def assemble(s):
    code = []
    for line in s.split('\n'):
        inst = tokenize(line)

        if inst[0].upper() not in _OPCODES:
            raise Exception(f"Unknown opcode {inst[0]}")
        
        if len(inst) - 1 != _OPCODES[inst[0]][1]:
            raise Exception(f"{inst[0]} requires {_OPCODES[inst[0]][1]} arguments, {len(inst)-1} are given")
        
        code.append(_OPCODES[inst[0]][0])
        
        for arg in inst[1:]:
            if type(arg) == str:
                raise Exception(f"Unknown argument {arg}")
            if type(arg) == list:
                code += arg
            else:
                code.append(arg)
    
    src = f"var Bytecode []uint8 = []uint8{{"
    for byte in code:
        src += f"0x{byte:02x}, "

    src = src[:-2] + "}"
    return src
    

if __name__ == "__main__":
    if len(sys.argv) == 1:
        sys.stderr.write("Please provide a filename\n")
        exit(1)

    with open(sys.argv[1]) as f:
        src = open(os.path.join(os.path.dirname(os.path.realpath(__file__)), "..", "impl", "bytecode.go"), "w")
        src.write("package main\n\n")
        src.write(assemble(f.read()))
        src.close()