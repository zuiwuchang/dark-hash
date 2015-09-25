package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"os"
	"strings"
)

var g_algorithms map[string]hash.Hash

const (
	BUFFER_SIZE = 1024 * 10
)

func GetAlgorithms() map[string]hash.Hash {
	if g_algorithms == nil {
		g_algorithms = make(map[string]hash.Hash)

		g_algorithms["sha1"] = sha1.New()
		g_algorithms["sha512"] = sha512.New()
		g_algorithms["sha256"] = sha256.New()
		g_algorithms["md5"] = md5.New()
	}

	return g_algorithms
}

func main() {
	h := flag.Bool("h", false, "show help")

	s := flag.String("s", "", "source file path")
	d := flag.String("d", "", "dist file path")

	hash := flag.String("hash", "all", "how to get hash (md5+sha512+...)")

	hashs := flag.Bool("hashs", false, "show algorithm")

	progress := flag.Bool("progress", false, "show progress")

	flag.Parse()
	if *h {
		flag.PrintDefaults()
		return
	} else if *hashs {
		algorithms := GetAlgorithms()
		for hash, _ := range algorithms {
			fmt.Print(hash)
			fmt.Print(" ")
		}
		fmt.Printf("\n")
		return
	}

	getHash(*s, *d, *hash, *progress)
}
func getHash(source, dist, hash string, progress bool) {
	f, err := os.Open(source)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		fmt.Println(err)
		return
	}
	total := info.Size()
	var pos int64

	b := make([]byte, BUFFER_SIZE, BUFFER_SIZE)

	resetAlgorithm(hash)
	pre_percentage := -1
	if progress {
		fmt.Println("0%  10   20   30   40   50   60   70   80   90  100%")
		fmt.Println("|----|----|----|----|----|----|----|----|----|----|")
	}
	for {
		n, err := f.Read(b)
		if n == 0 {
			break
		} else if err != nil {
			fmt.Println(n, err)
			return
		}
		tmp := b[:n]

		algorithmWrite(hash, tmp)

		if !progress {
			continue
		}
		pos += int64(n)

		percentage := int(pos*100/total) / 2
		if pre_percentage == percentage {
			continue
		}
		for i := 0; i < percentage-pre_percentage; i++ {
			fmt.Print("*")
		}

		pre_percentage = percentage
	}

	data := algorithmSum(hash)

	if dist == "" {
		fmt.Println("\n\nhash\n{")
		for k, v := range data {
			fmt.Printf("\t%v:'%v'\n", k, v)
		}
		fmt.Println("}")
	} else {
		f, err := os.OpenFile(dist, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()

		f.WriteString("{\n")
		for k, v := range data {
			f.WriteString(fmt.Sprintf("\t%v:'%v'\n", k, v))
		}
		f.WriteString("}\n")
	}
}
func resetAlgorithm(hash string) {
	algorithms := GetAlgorithms()
	if hash == "all" {
		for _, algorithm := range algorithms {
			algorithm.Reset()
		}
	} else {
		names := strings.Split(hash, "+")
		for _, name := range names {
			if algorithm, ok := algorithms[name]; ok {
				algorithm.Reset()
			}
		}
	}
}
func algorithmWrite(hash string, b []byte) {
	algorithms := GetAlgorithms()
	if hash == "all" {
		for _, algorithm := range algorithms {
			algorithm.Write(b)
		}
	} else {
		names := strings.Split(hash, "+")
		for _, name := range names {
			if algorithm, ok := algorithms[name]; ok {
				algorithm.Write(b)
			}
		}
	}
}
func algorithmSum(hash string) map[string]string {
	rs := make(map[string]string)

	algorithms := GetAlgorithms()
	if hash == "all" {
		for name, algorithm := range algorithms {
			rs[name] = hex.EncodeToString(algorithm.Sum(nil))
		}
	} else {
		names := strings.Split(hash, "+")
		for _, name := range names {
			if algorithm, ok := algorithms[name]; ok {
				rs[name] = hex.EncodeToString(algorithm.Sum(nil))
			}
		}
	}

	return rs
}
