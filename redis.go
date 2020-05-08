package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"time"
)

const (
	URLIdKey           = "next.url.id"         // 全局自增器
	ShortlinkKey       = "shortlink:%s:url"    // 短地址和原地址的映射
	URLHashKey         = "urlhash:%s:url"      // 地址hash和短地址的映射
	ShortLinkDetailKey = "shortlink:%s:detail" // 短地址和详情的映射
)

type RedisCli struct {
	Cli *redis.Client
}

//短地址详细信息
type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

func NewRedisCli(addr string, passwd string, db int) *RedisCli {

	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})
	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}
	return &RedisCli{Cli: c}
}

func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	//计算长地址的哈希值，便于储存长地址作为key
	urlHash := toSha1(url)
	var id int64

	d, err := r.Cli.Get(fmt.Sprintf(URLHashKey, urlHash)).Result()
	if err == redis.Nil {
		// URLHashKey不存在
		// 1. 用自增量取得id
		id1, err1 := r.Cli.Incr(URLIdKey).Result()
		id = id1
		if err1 != nil {
			return "", err1
		}
	} else if err != nil {
		return "", err
	} else {
		if d != "{}" {
			id = Decode(d)
			//return d, nil
		}
	}

	// 把这个id转成62进制
	encodeId := Encode(id)

	// 2.存短地址和长地址的映射
	err = r.Cli.Set(fmt.Sprintf(ShortlinkKey, encodeId), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// 3. 存长地址的哈希值和短地址的映射
	err = r.Cli.Set(fmt.Sprintf(URLHashKey, urlHash), encodeId, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	detail, err := json.Marshal(&URLDetail{
		URL:                 url,
		CreatedAt:           time.Now().String(),
		ExpirationInMinutes: time.Duration(exp),
	})
	if err != nil {
		return "", err
	}

	// 4. 存短地址和详情的映射，创建时间过期时间等
	err = r.Cli.Set(fmt.Sprintf(ShortLinkDetailKey, encodeId), detail, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", nil
	}
	return encodeId, nil
}

//获取短地址详细信息
func (r *RedisCli) ShortLinkInfo(encodeId string) (interface{}, error) {
	detail, err := r.Cli.Get(fmt.Sprintf(ShortLinkDetailKey, encodeId)).Result()
	if err == redis.Nil {
		return "", StatusError{
			Code: 404,
			Err:  errors.New("unknown short URL"),
		}
	} else if err != nil {
		return "", err
	}
	return detail, nil
}

//获取长地址
func (r *RedisCli) UnShorten(encodeId string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(ShortlinkKey, encodeId)).Result()
	if err == redis.Nil {
		return "", StatusError{
			Code: 404,
			Err:  errors.New("unknown short link"),
		}
	} else if err != nil {
		return "", err
	}
	return url, nil
}

func toSha1(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
