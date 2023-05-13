package redis

import (
	"fmt"
	"reflect"
	"time"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/str"
	"github.com/redis/go-redis/v9"
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
		for _, v := range cmd.Val() {
			if v == nil {
				return redis.Nil
			}
		}
	}

	for _, d := range dest {
		if err := s.Scan(d); err != nil {
			return err
		}
	}

	return nil
}

func addPrefix(prefix, key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

// mapArticle converts an article to a map. Formats time.Time fields to RFC3339.
func mapArticle(article entity.Article) map[string]string {
	v := reflect.ValueOf(article)
	values := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fType := field.Type()
		fieldName := str.ToLowerSnakeCase(v.Type().Field(i).Name)

		// Format time.Time fields to RFC3339.
		// Redis-go doesn't support time.Time fields scanning.
		if fType == reflect.TypeOf(time.Time{}) {
			values[fieldName] = field.Interface().(time.Time).Format(time.RFC3339)
			continue
		}

		if s := field.String(); s != "" {
			values[fieldName] = s
		}
	}
	return values
}
