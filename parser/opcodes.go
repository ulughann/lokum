package parser

type Opcode = byte

const (
	OpConstant Opcode = iota
	OpBComplement
	OpPop
	OpTrue
	OpFalse
	OpEqual
	OpNotEqual
	OpMinus
	OpLNot
	OpJumpFalsy
	OpAndJump
	OpOrJump
	OpJump
	OpNull
	OpArray
	OpMap
	OpError
	OpImmutable
	OpIndex
	OpSliceIndex
	OpCall
	OpReturn
	OpGetGlobal
	OpSetGlobal
	OpSetSelGlobal
	OpGetLocal
	OpSetLocal
	OpDefineLocal
	OpSetSelLocal
	OpGetFreePtr
	OpGetFree
	OpSetFree
	OpGetLocalPtr
	OpSetSelFree
	OpGetBuiltin
	OpClosure
	OpIteratorInit
	OpIteratorNext
	OpIteratorKey
	OpIteratorValue
	OpBinaryOp
	OpSuspend
)

var OpcodeNames = [...]string{
	OpConstant:      "CONST",
	OpPop:           "POP",
	OpTrue:          "TRUE",
	OpFalse:         "FALSE",
	OpBComplement:   "NEG",
	OpEqual:         "EQL",
	OpNotEqual:      "NEQ",
	OpMinus:         "NEG",
	OpLNot:          "NOT",
	OpJumpFalsy:     "JMPF",
	OpAndJump:       "ANDJMP",
	OpOrJump:        "ORJMP",
	OpJump:          "JMP",
	OpNull:          "NULL",
	OpGetGlobal:     "GETG",
	OpSetGlobal:     "SETG",
	OpSetSelGlobal:  "SETSG",
	OpArray:         "ARR",
	OpMap:           "MAP",
	OpError:         "ERROR",
	OpImmutable:     "IMMUT",
	OpIndex:         "INDEX",
	OpSliceIndex:    "SLICE",
	OpCall:          "CALL",
	OpReturn:        "RET",
	OpGetLocal:      "GETL",
	OpSetLocal:      "SETL",
	OpDefineLocal:   "DEFL",
	OpSetSelLocal:   "SETSL",
	OpGetBuiltin:    "BUILTIN",
	OpClosure:       "CLOSURE",
	OpGetFreePtr:    "GETFP",
	OpGetFree:       "GETF",
	OpSetFree:       "SETF",
	OpGetLocalPtr:   "GETLP",
	OpSetSelFree:    "SETSF",
	OpIteratorInit:  "ITER",
	OpIteratorNext:  "ITNXT",
	OpIteratorKey:   "ITKEY",
	OpIteratorValue: "ITVAL",
	OpBinaryOp:      "BINARYOP",
	OpSuspend:       "SUSPEND",
}

var OpcodeOperands = [...][]int{
	OpConstant:      {2},
	OpPop:           {},
	OpTrue:          {},
	OpFalse:         {},
	OpBComplement:   {},
	OpEqual:         {},
	OpNotEqual:      {},
	OpMinus:         {},
	OpLNot:          {},
	OpJumpFalsy:     {2},
	OpAndJump:       {2},
	OpOrJump:        {2},
	OpJump:          {2},
	OpNull:          {},
	OpGetGlobal:     {2},
	OpSetGlobal:     {2},
	OpSetSelGlobal:  {2, 1},
	OpArray:         {2},
	OpMap:           {2},
	OpError:         {},
	OpImmutable:     {},
	OpIndex:         {},
	OpSliceIndex:    {},
	OpCall:          {1, 1},
	OpReturn:        {1},
	OpGetLocal:      {1},
	OpSetLocal:      {1},
	OpDefineLocal:   {1},
	OpSetSelLocal:   {1, 1},
	OpGetBuiltin:    {1},
	OpClosure:       {2, 1},
	OpGetFreePtr:    {1},
	OpGetFree:       {1},
	OpSetFree:       {1},
	OpGetLocalPtr:   {1},
	OpSetSelFree:    {1, 1},
	OpIteratorInit:  {},
	OpIteratorNext:  {},
	OpIteratorKey:   {},
	OpIteratorValue: {},
	OpBinaryOp:      {1},
	OpSuspend:       {},
}

func ReadOperands(numOperands []int, ins []byte) (operands []int, offset int) {
	for _, width := range numOperands {
		switch width {
		case 1:
			operands = append(operands, int(ins[offset]))
		case 2:
			operands = append(operands, int(ins[offset+1])|int(ins[offset])<<8)
		}
		offset += width
	}
	return
}
