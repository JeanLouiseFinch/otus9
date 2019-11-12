package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/cheggaaa/pb/v3"
)


const BUFSIZE = 4096 * 1024


type input struct {
	offset, limit    int64 
	fromFile, toFile string  
	sourceSize       int64    
	fSrc             *os.File 
	fDst             *os.File 
}

var core = input{}

func init() {
	flag.StringVar(&core.fromFile, "from", "", "Source file from Copy")
	flag.StringVar(&core.toFile, "to", "", "Destination file to Copy")
	flag.Int64Var(&core.offset, "offset", 0, "Offset in source file to begin copying")
	flag.Int64Var(&core.limit, "limit", 0, "Copy limit bytes to Destination File")
	flag.Parse()
}

func main() {
	core.check()
	core.Copy()
}

func (i *input) check() {
	if i.toFile == "" {
		flag.Usage()
		log.Fatalf("Wrong argument values to=%v", i.toFile)
	}

	_, err := os.Stat(i.toFile)
	if !os.IsNotExist(err) {
		log.Fatal("Sorry, Destination file exists!")
	}

	if i.fromFile == "" {
		flag.Usage()
		log.Fatalf("Wrong argument values from=%v", i.fromFile)
	}

	fileSource, err := os.Stat(i.fromFile)
	if os.IsNotExist(err) {
		log.Fatal(err)
	}
	i.sourceSize = fileSource.Size()

	if i.limit <= 0 {
		i.limit = i.sourceSize
	}
}

func (i *input) getSizeCopy() int64 {
	if i.sourceSize < (i.limit + i.offset) {
		return i.sourceSize - i.offset
	}
	return i.limit
}

func (i *input) Open() {

	file, err := os.Open(i.fromFile)
	if err != nil {
		log.Fatal(err)
	}
	if _, err = file.Seek(i.offset, 0); err != nil {
		log.Fatal(err)
	}
	i.fSrc = file
}

func (i *input) Create() {
	file, err := os.Create(i.toFile)
	if err != nil {
		log.Fatal(err)
	}
	i.fDst = file
}

func (i *input) Write() error {
	buf := make([]byte, BUFSIZE)
	mrReader := io.LimitReader(i.fSrc, i.limit)

	//progress bar
	bar := pb.Start64(i.getSizeCopy())
	bar.SetWidth(150)
	bar.Set(pb.Bytes, true)

	for {
		c, err := mrReader.Read(buf)

		if c > 0 {
			if _, err := i.fDst.Write(buf[:c]); err != nil {
				return err
			}
			//progress bar
			bar.Add(c)
		}

		if err != nil {
			bar.Finish()
			return err
		}
	}
}


func (i *input) Copy() {
	i.Open()
	defer i.fSrc.Close()

	i.Create()
	defer i.fDst.Close()

	if err := i.Write(); err != io.EOF {
		log.Fatal(err)
	}
}