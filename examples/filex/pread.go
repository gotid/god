package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"github.com/gotid/god/lib/filex"
	"github.com/gotid/god/lib/fx"
	"github.com/gotid/god/lib/logx"
	"gopkg.in/cheggaaa/pb.v1"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	file        = flag.String("f", "", "输入文件")
	concurrent  = flag.Int("c", runtime.NumCPU(), "并发协程数量")
	wordVecDict TXDictionary
)

type (
	Vector []float64

	TXDictionary struct {
		EmbeddingCount int64
		Dim            int64
		Dict           map[string]Vector
	}

	pair struct {
		key string
		vec Vector
	}
)

func FastLoad(filename string) error {
	if filename == "" {
		return errors.New("缺少有效的字典")
	}

	now := time.Now()
	defer func() {
		logx.Infof("初始化字典耗时 %v", time.Since(now))
	}()

	dictFile, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer dictFile.Close()

	header, err := filex.FirstLine(filename)
	if err != nil {
		return err
	}

	total := strings.Split(header, " ")
	wordVecDict.EmbeddingCount, err = strconv.ParseInt(total[0], 10, 64)
	if err != nil {
		return err
	}

	wordVecDict.Dim, err = strconv.ParseInt(total[1], 10, 64)
	if err != nil {
		return err
	}

	wordVecDict.Dict = make(map[string]Vector, wordVecDict.EmbeddingCount)

	ranges, err := filex.SplitLineChunks(filename, *concurrent)
	if err != nil {
		return err
	}

	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	bar := pb.New64(info.Size()).SetUnits(pb.U_BYTES).Start()
	fx.From(func(source chan<- any) {
		for _, each := range ranges {
			source <- each
		}
	}).Walk(func(item any, pipe chan<- any) {
		offsetRange := item.(filex.OffsetRange)
		scanner := bufio.NewScanner(filex.NewRangeReader(dictFile, offsetRange.Start, offsetRange.Stop))
		scanner.Buffer([]byte{}, 1<<20)
		reader := filex.NewProgressScanner(scanner, bar)
		if offsetRange.Start == 0 {
			// 跳过头部
			reader.Scan()
		}
		for reader.Scan() {
			text := reader.Text()
			elements := strings.Split(text, " ")
			vec := make(Vector, wordVecDict.Dim)
			for i, element := range elements {
				if i == 0 {
					continue
				}

				v, err := strconv.ParseFloat(element, 64)
				if err != nil {
					return
				}

				vec[i-1] = v
			}
			pipe <- pair{
				key: elements[0],
				vec: vec,
			}
		}
	}).ForEach(func(item any) {
		p := item.(pair)
		wordVecDict.Dict[p.key] = p.vec
	})

	return nil
}

func main() {
	flag.Parse()

	start := time.Now()
	if err := FastLoad(*file); err != nil {
		log.Fatal(err)
	}

	fmt.Println(wordVecDict.Dict)
	fmt.Println(time.Since(start))
}
