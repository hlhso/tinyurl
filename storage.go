package main

//Storage:存储功能封装接口
type Storage interface {
	Shorten(url string, exp int64) (string, error)
	ShortLinkInfo(eid string) (interface{}, error)
	UnShorten(eid string) (string, error)
}
