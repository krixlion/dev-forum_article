package gentest

import (
	"encoding/json"
	"math/rand"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev-forum_article/pkg/entity"
)

func RandomString(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	v := make([]rune, length)
	for i := range v {
		v[i] = letters[rand.Intn(len(letters))]
	}
	return string(v)
}

// RandomArticle panics on hardware error.
// It should be used ONLY for testing.
func RandomArticle(titleLen, bodyLen int) entity.Article {
	id := uuid.Must(uuid.NewV4())
	userId := uuid.Must(uuid.NewV4())

	return entity.Article{
		Id:     id.String(),
		UserId: userId.String(),
		Title:  RandomString(titleLen),
		Body:   RandomString(bodyLen),
	}
}

// RandomArticle returns a random article marshaled
// to JSON and panics on error.
// It should be used ONLY for testing.
func RandomJSONArticle(titleLen, bodyLen int) []byte {
	article := RandomArticle(titleLen, bodyLen)
	json, err := json.Marshal(article)
	if err != nil {
		panic(err)
	}
	return json
}
