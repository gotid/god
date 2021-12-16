package seg

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"git.zc0901.com/go/god/lib/gutil"

	"git.zc0901.com/go/god/lib/container/gset"

	"git.zc0901.com/go/god/lib/stringx"

	"github.com/yanyiwu/gojieba"

	"git.zc0901.com/go/god/lib/seg/clean"
)

var (
	// MainTag 主标签
	MainTag = []string{"空", "物"}
	// DomainTag 领域标签
	DomainTag = []string{"户", "风", "空", "局", "牌", "形", "色", "纹", "材", "特", "物"}
	// SortTag 排序标签
	SortTag = []string{"户", "组", "风", "空", "局", "牌", "形", "色", "纹", "材", "特", "物"}
)

// Segmenter 是一个分词器
type Segmenter struct {
	jieba        *gojieba.Jieba
	stopWords    []string
	combineWords []string
	combineMap   map[string]string
	synonymMap   map[string]string
	cleaner      clean.Cleaner
	groups       [][]string // 领域标签组合方式
	debugMode    bool
}

// Keyword 是一个关键词
type Keyword struct {
	Word     string  `json:"word,omitempty"`
	Tag      string  `json:"tag,omitempty"`
	Weight   float64 `json:"weight,omitempty"`
	Start    int     `json:"start,omitempty"`
	Stop     int     `json:"stop,omitempty"`
	Distance int     `json:"distance,omitempty"`
}

// NewSegmenter 返回一个新的分词器。
func NewSegmenter(dict string,
	stopWords []string,
	combineWords []string, combineMap map[string]string,
	synonymMap map[string]string,
	groups [][]string,
	debug ...bool) *Segmenter {
	s := &Segmenter{
		jieba:        gojieba.NewJieba(gojieba.DICT_PATH, gojieba.HMM_PATH, dict),
		stopWords:    stopWords,
		combineWords: combineWords, combineMap: combineMap,
		synonymMap: synonymMap,
		groups:     groups,
	}
	s.cleaner = clean.NewCleaner(
		"hello world",
		clean.WithStopWords(stopWords...),
		clean.WithCombineWords(combineWords, combineMap),
		clean.WithSynonymWords(synonymMap),
	)
	if len(s.groups) == 0 {
		s.groups = [][]string{
			{"风", "空"},
			{"色", "空"},
			{"材", "局"},
			{"空", "物"},
			{"形", "物"},
			{"色", "物"},
			{"纹", "物"},
			{"材", "物"},
			{"特", "物"},
			{"牌", "物"},
		}
	}
	if len(debug) > 0 {
		s.debugMode = debug[0]
	}
	return s
}

// CutForSearch 搜索分词
func (segmenter *Segmenter) CutForSearch(q string, dist int, domainMode ...bool) []Keyword {
	s := segmenter.cleaner.Query(q).Clean()
	segmenter.debug("查询", s)

	keywords := segmenter.Cut(s)

	// 距离
	if dist < 0 {
		dist = 1
	}

	usedIdxSet := gset.NewIntSet()
	cb := func(usedIdx []int) {
		usedIdxSet.Add(usedIdx...)
	}
	result := make(map[string]Keyword)
	for _, group := range segmenter.groups {
		m := segmenter.Combine(keywords, group[0], group[1], dist, cb)
		for kw, keyword := range m {
			v, exists := result[kw]
			if !exists || keyword.Distance < v.Distance {
				result[kw] = keyword
			}
		}
	}

	// 补词
	if usedIdxSet.Size() > 0 {
		domain := true
		if len(domainMode) > 0 {
			domain = domainMode[0]
		}
		for i, keyword := range keywords {
			if usedIdxSet.Contains(i) {
				continue
			}
			if keyword.Word != "" && keyword.Weight > 0 {
				if domain && !segmenter.isDomainTag(keyword.Tag) {
					continue
				}
				result[keyword.Word] = Keyword{
					Word:   keyword.Word,
					Tag:    keyword.Tag,
					Weight: keyword.Weight,
				}
			}
		}
	}

	// 排词
	ret := segmenter.sort(result)

	return ret
}

// Combine 组词
func (segmenter *Segmenter) Combine(keywords []*Keyword, tag1, tag2 string, maxDist int, cb func(usedIdx []int)) map[string]Keyword {
	// 分组
	var idx1s []int
	var idx2s []int
	var idx3s []int
	for i, kw := range keywords {
		segmenter.debug(i, kw)
		if kw.Tag == tag1 {
			idx1s = append(idx1s, i)
		} else if kw.Tag == tag2 {
			idx2s = append(idx2s, i)
		} else {
			idx3s = append(idx3s, i)
		}
	}
	segmenter.debug(tag1, idx1s, tag2, idx2s, "其他", idx3s)

	// 配对
	idx1idx2sMap := make(map[int][][]int)
	for _, idx1 := range idx1s {
		for _, idx2 := range idx2s {
			dist := int(math.Abs(float64(idx1)-float64(idx2))) - 1
			if dist <= maxDist {
				idx1idx2sMap[idx1] = append(idx1idx2sMap[idx1], []int{idx2, dist})
			}
		}
	}
	segmenter.debug(idx1idx2sMap)

	// 组合
	idx1MinDist := make(map[int]int)
	idx1idx2Map := make(map[int]int)
	for idx1, idx2s := range idx1idx2sMap {
		for _, v := range idx2s {
			idx2, dist := v[0], v[1]
			segmenter.debug("距离", idx1, idx2, dist)
			_, exists := idx1MinDist[idx1]
			if !exists || dist < idx1MinDist[idx1] {
				idx1MinDist[idx1] = dist
				idx1idx2Map[idx1] = idx2
			}
		}
	}

	result := make(map[string]Keyword)
	usedIdxMap := make(map[int]bool)
	for idx1, idx2 := range idx1idx2Map {
		// 判断是否跳过不是紧挨着的组合
		if idx1MinDist[idx1] > 0 {
			// 跳过前1个或后1个紧挨着的是主标签
			skipPrev := idx1 != 0 &&
				segmenter.isMainTag(keywords[idx1-1].Tag) &&
				keywords[idx1-1].Tag != keywords[idx2].Tag
			skipNext := idx1 != len(keywords)-1 &&
				segmenter.isMainTag(keywords[idx1+1].Tag) &&
				keywords[idx1+1].Tag != keywords[idx2].Tag
			if skipPrev || skipNext {
				usedIdxMap[idx1] = true
				continue
			}
		}

		combineWord := keywords[idx1].Word + keywords[idx2].Word
		segmenter.debug("组合词", idx1, idx2, combineWord)
		dist := idx1MinDist[idx1]
		kw, exists := result[combineWord]
		if !exists || dist < kw.Distance {
			result[combineWord] = Keyword{
				Word:     combineWord,
				Tag:      "组",
				Weight:   2,
				Distance: dist,
			}
		}
		usedIdxMap[idx1] = true
		usedIdxMap[idx2] = true
	}

	// 处理索引使用情况
	if cb != nil {
		usedIdx := make([]int, 0, len(usedIdxMap))
		for idx := range usedIdxMap {
			usedIdx = append(usedIdx, idx)
		}
		cb(usedIdx)
	}

	return result
}

// Cut 分词
func (segmenter *Segmenter) Cut(s string) []*Keyword {
	words := segmenter.jieba.Cut(s, true)
	keywords := make([]*Keyword, len(words))
	for i, word := range words {
		keywords[i] = &Keyword{Word: segmenter.cleaner.Synonym(word)}
	}

	tags := segmenter.jieba.Tag(s)
	for i, tag := range tags {
		parts := strings.SplitN(tag, "/", 2)
		keywords[i].Tag = parts[1]
	}

	weights := segmenter.jieba.ExtractWithWeight(s, 10)
	for _, weight := range weights {
		for _, keyword := range keywords {
			if keyword.Word == weight.Word {
				keyword.Weight = weight.Weight
				break
			}
		}
	}

	tokens := segmenter.jieba.Tokenize(s, gojieba.DefaultMode, true)
	for _, token := range tokens {
		for _, keyword := range keywords {
			if keyword.Word == token.Str {
				keyword.Start = token.Start
				keyword.Stop = token.End
				break
			}
		}
	}
	return keywords
}

func (segmenter *Segmenter) isMainTag(tag string) bool {
	return stringx.Contains(MainTag, tag)
}

func (segmenter *Segmenter) isDomainTag(tag string) bool {
	return stringx.Contains(DomainTag, tag)
}

func (segmenter *Segmenter) sort(kws map[string]Keyword) (list []Keyword) {
	segmenter.debug("结果映射", kws)
	list = make([]Keyword, 0, len(kws))
	for _, keyword := range kws {
		list = append(list, keyword)
	}
	sort.Slice(list, func(i, j int) bool {
		idx1 := gutil.IndexOf(list[i].Tag, SortTag)
		idx2 := gutil.IndexOf(list[j].Tag, SortTag)
		if idx1 == -1 {
			idx1 = 999
		}
		if idx2 == -1 {
			idx2 = 999
		}
		return idx1 < idx2
	})
	return list
}

func (segmenter *Segmenter) debug(a ...interface{}) {
	if segmenter.debugMode {
		fmt.Println(a...)
	}
}
