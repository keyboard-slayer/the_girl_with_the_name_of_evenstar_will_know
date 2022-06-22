#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import sys

_OPCODES = {
    "INIT": (0x00, 0),
    "PUSH": (0x01, 1),
    "POP": (0x02, 1),
    "PRINT": (0x03, 0),
    "OPEN": (0x05, 2),
    "WRITE": (0x08, 1),
    "READ": (0x0D, 2),
    "FSIZE": (0x15, 2),
    "SYSTEM": (0x22, 0),
    "PUSHREG": (0x37, 1),
    "APPEND": (0x59, 1),
    "MKDIR": (0x90, 0),
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
        if token[0] == '"' or in_str:
            if token[-1] == '"':
                if tmp:
                    result.append(obfuscate_string(tmp[1:] + token[:-1]))
                    tmp = ""
                else:
                    result.append(obfuscate_string(token[1:-1]))
                in_str = False
            else:
                in_str = True
                tmp += token + " "
        elif token[0] == "r":
            result.append(0xFF - int(token[1:]))
        else:
            result.append(token)

    return result


def macro(line):
    inst = line[1:].split()
    ret = ""

    match inst[0].upper():
        case "EMBED":
            assert len(inst) == 2
            with open(inst[1], "r") as f:
                ret = ["PUSH", obfuscate_string(f.read())]

    return ret


def assemble(s):
    code = []
    for line in s.split("\n"):
        if not line or line[0] == ";":
            continue
        if line[0] == "#":
            inst = macro(line)
        else:
            inst = tokenize(line)

        if inst[0].upper() not in _OPCODES:
            raise Exception(f"Unknown opcode {inst[0]}")

        if len(inst) - 1 != _OPCODES[inst[0]][1]:
            raise Exception(
                f"{inst[0]} requires {_OPCODES[inst[0]][1]} arguments, {len(inst)-1} are given"
            )

        code.append(_OPCODES[inst[0]][0])

        for arg in inst[1:]:
            if type(arg) == str:
                raise Exception(f"Unknown argument {arg}")
            if type(arg) == list:
                code += arg
            else:
                code.append(arg)

    src = f"var Bytecode []byte = []byte{{"
    for byte in code:
        src += f"0x{byte:02x}, "

    src = src[:-2] + "}"
    return src


if __name__ == "__main__":
    if len(sys.argv) == 1:
        sys.stderr.write("Please provide a filename\n")
        exit(1)

    with open(sys.argv[1]) as f:
        src = open(
            os.path.join(
                os.path.dirname(os.path.realpath(__file__)), "..", "impl", "bytecode.go"
            ),
            "w",
        )
        src.write("package main\n\n")
        src.write(assemble(f.read()))
        src.close()
