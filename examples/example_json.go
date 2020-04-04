package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/nasjp/jsontotype"
)

const (
	exampleJson = `{
  "id": 1,
  "name": "bob",
  "age": 23,
  "score": 3.7,
  "favoriteFoods": [
    "gohan",
    "pan"
  ],
  "credentials": {
    "id": 1,
    "password": "123456789"
  }
}`
	packageName = "user"
	typeName    = "User"
)

func main() {
	r := strings.NewReader(exampleJson)
	result, err := jsontotype.Exec(r, packageName, typeName)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

/* fmt.Println(result)
package user

type User struct {
        ID            int64    `json:"id"`
        Name          string   `json:"name"`
        Age           int64    `json:"age"`
        Score         float64  `json:"score"`
        FavoriteFoods []string `json:"favoriteFoods"`
        Credentials   struct {
                ID       int64  `json:"id"`
                Password string `json:"password"`
        } `json:"credentials"`
}
*/
