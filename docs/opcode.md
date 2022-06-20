## Opcodes

| Mnemonic | Opcode | Description                                                              |
|----------|--------|--------------------------------------------------------------------------|
| INIT     | 0x00   | Set X number of register (X from the stack)                              |
| PUSH     | 0x01   | Push value in the stack                                                  |
| POP      | 0x02   | Pop value in the stack to a register                                     |
| PRINT    | 0x03   | Print value from the stack                                               |
| OPEN     | 0x05   | Take a file, a mode and a register as arg                                |
| WRITE    | 0x08   | Pop a value from the stack and write it to a file                        |
| READ     | 0x0d   | Take a register and a number of byte as arg and push result in the stack |
| FSIZE    | 0x15   | Take a register with a file and push the file size into the stack        |

*Note*: Registers are counted backwards, for example r0 = 0xff and r1 = 0xfe