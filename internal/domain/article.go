package domain

import "time"

type Article struct {
	Id      int64
	Title   string
	Content string
	Author  Author
	Status  ArticleStatus
	Ctime   time.Time
	Utime   time.Time
}

func (a Article) Abstract() string {
	// 摘要我们取前几句。
	// 要考虑一个中文问题
	cs := []rune(a.Content)
	if len(cs) < 100 {
		return a.Content
	}
	// 英文怎么截取一个完整的单词，我的看法是……不需要纠结，就截断拉到
	// 词组、介词，往后找标点符号
	return string(cs[:100])
}

type Author struct {
	Id   int64
	Name string
}

type ArticleStatus uint8

const (
	// ArticleStatusUnknown 为了避免零值之类的问题
	ArticleStatusUnknown ArticleStatus = iota
	ArticleStatusUnpublished
	ArticleStatusPublished
	ArticleStatusPrivate
)

func (s ArticleStatus) ToUint8() uint8 {
	return uint8(s)
}

func (s ArticleStatus) NonPublished() bool {
	return s != ArticleStatusPublished
}

func (s ArticleStatus) String() string {
	switch s {
	case ArticleStatusPrivate:
		return "private"
	case ArticleStatusUnpublished:
		return "unpublished"
	case ArticleStatusPublished:
		return "published"
	default:
		return "unknown"
	}
}
