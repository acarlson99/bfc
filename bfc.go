// brainfuck compiler to LLVM IR
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/llir/llvm/ir"
	"github.com/llir/llvm/ir/constant"
	"github.com/llir/llvm/ir/enum"
	"github.com/llir/llvm/ir/types"
)

type BlockSet struct {
	inner, outer, head *ir.Block
}

func CI8(n int64) *constant.Int {
	return constant.NewInt(types.I8, n)
}

func CI32(n int64) *constant.Int {
	return constant.NewInt(types.I32, n)
}

func CI64(n int64) *constant.Int {
	return constant.NewInt(types.I64, n)
}

const (
	arrLen = 0xFFFFF
)

func main() {
	var arch string
	var outputFile string

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  Compiler for Brainfuck targetting LLVM.\n")
		flag.PrintDefaults()
	}
	flag.StringVar(&arch, "arch", "x86_64-pc-linux-gnu", "Target architecture")
	flag.StringVar(&outputFile, "o", "", "Output file path")
	flag.Parse()
	args := flag.Args()

	var scanner *bufio.Scanner
	if len(args) > 0 {
		filePath := args[0]
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	} else {
		scanner = bufio.NewScanner(os.Stdin)
	}

	m := compile(scanner)
	m.TargetTriple = arch

	if outputFile != "" {
		output, err := os.Create(outputFile)
		if err != nil {
			fmt.Println("Error creating output file:", err)
			return
		}
		defer output.Close()
		fmt.Fprintln(output, m)
	} else {
		fmt.Println(m)
	}
}

func compile(scanner *bufio.Scanner) *ir.Module {
	m := ir.NewModule()

	putchar := m.NewFunc("putchar", types.I32, ir.NewParam("ch", types.I8))
	getchar := m.NewFunc("safe_getchar", types.I8)
	// getchar := m.NewFunc("getchar", types.I8)
	memset := m.NewFunc("memset", types.Void, ir.NewParam("ptr", types.I8Ptr), ir.NewParam("val", types.I8), ir.NewParam("len", types.I64))

	main := m.NewFunc("main", types.I32)
	block := main.NewBlock("outer")

	arrType := types.NewArray(arrLen, types.I8)
	arr := block.NewAlloca(arrType)
	{
		ptr := block.NewGetElementPtr(arrType, arr, CI64(0), CI64(0))
		block.NewCall(memset, ptr, CI8(0), CI64(arrLen))
	}

	pc := block.NewAlloca(types.I64)
	block.NewStore(CI64(0), pc)

	n := 0

	loadArrPtr := func() *ir.InstGetElementPtr {
		// loadArrPtr := func(block *ir.Block, arr, idx value.Value) *ir.InstGetElementPtr {
		return block.NewGetElementPtr(arrType, arr, CI64(0), block.NewLoad(types.I64, pc))
	}

	// stack := []BlockSet{{inner: block, outer: end, head: block}}
	stack := []BlockSet{}
	for scanner.Scan() {
		for _, c := range scanner.Text() {
			switch c {
			case '+':
				// inc bytes[dp]
				ptr := loadArrPtr()
				counter := block.NewAdd(block.NewLoad(types.I8, ptr), CI8(1))
				block.NewStore(counter, ptr)
			case '-':
				// dec bytes[dp]
				ptr := loadArrPtr()
				counter := block.NewAdd(block.NewLoad(types.I8, ptr), CI8(-1))
				block.NewStore(counter, ptr)
			case '>':
				// dp++
				block.NewStore(block.NewAdd(block.NewLoad(types.I64, pc), CI8(1)), pc)
			case '<':
				// dp--
				block.NewStore(block.NewAdd(block.NewLoad(types.I64, pc), CI8(-1)), pc)
			case '.':
				// print
				ptr := loadArrPtr()
				block.NewCall(putchar, block.NewLoad(types.I8, ptr))
			case ',':
				// read
				ch := block.NewCall(getchar)
				ptr := loadArrPtr()
				block.NewStore(ch, ptr)
			case '[':
				// if bytes[dp]==0 then jump to end of block, else advance instruction ptr
				ptr := loadArrPtr()
				counter := block.NewLoad(types.I8, ptr)

				cmpResult := block.NewICmp(enum.IPredNE, counter, constant.NewInt(types.I8, 0))

				bs := BlockSet{
					inner: main.NewBlock(fmt.Sprintf("__inner_%d", n)),
					outer: main.NewBlock(fmt.Sprintf("__outer_%d", n)),
				}
				stack = append(stack, bs)
				n++

				block.NewCondBr(cmpResult, bs.inner, bs.outer)
				block = bs.inner
			case ']':
				// jnz block start
				if len(stack) == 0 {
					log.Fatal("unexpected closing bracket")
				}
				bs := stack[len(stack)-1]
				stack = stack[:len(stack)-1]

				ptr := loadArrPtr()
				counter := block.NewLoad(types.I8, ptr)

				cmpResult := block.NewICmp(enum.IPredNE, counter, constant.NewInt(types.I8, 0))

				block.NewCondBr(cmpResult, bs.inner, bs.outer)
				block = bs.outer
			default:
			}
		}
	}

	if len(stack) != 0 {
		log.Fatal("unclosed '['")
	}
	block.NewRet(CI32(0))
	return m
}
