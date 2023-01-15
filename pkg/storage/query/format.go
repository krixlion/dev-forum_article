package query

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-redis/redis/v9"
	"github.com/krixlion/dev_forum-article/pkg/entity"
)

type scanCmder interface {
	Scan(dst interface{}) error
}

// scan enhances the Scan method of a scanCmder with these features:
//   - it returns the error redis.Nil when the key does not exist. See https://github.com/go-redis/redis/issues/1668
//   - it supports embedded struct better. See https://github.com/go-redis/redis/issues/2005#issuecomment-1019667052
func scan(s scanCmder, dest ...interface{}) error {
	switch cmd := s.(type) {
	case *redis.MapStringStringCmd:
		if len(cmd.Val()) == 0 {
			return redis.Nil
		}
	case *redis.SliceCmd:
		keyExists := false
		for _, v := range cmd.Val() {
			if v != nil {
				keyExists = true
				break
			}
		}
		if !keyExists {
			return redis.Nil
		}
	}

	for _, d := range dest {
		if err := s.Scan(d); err != nil {
			return err
		}
	}

	return nil
}

func addArticlesPrefix(key string) string {
	return fmt.Sprintf("%s:%s", articlesPrefix, key)
}

func toLowerSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func mapArticle(article entity.Article) map[string]string {
	v := reflect.ValueOf(article)
	values := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if s := field.String(); s != "" {
			fieldName := toLowerSnakeCase(v.Type().Field(i).Name)
			values[fieldName] = s
		}
	}
	return values
}
