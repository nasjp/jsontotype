package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/nasjp/jsontotype"
)

const exampleJson = `{
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

const exampleJson2 = `"hoge"`

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	r := strings.NewReader(exampleJson)

	str, err := jsontotype.Exec(r, "user", "User")
	if err != nil {
		return err
	}
	fmt.Println(str)
	return nil
}
