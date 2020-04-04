package jsontotype_test

import (
	"io"
	"strings"
	"testing"

	"github.com/nasjp/jsontotype"
)

func TestExec(t *testing.T) {
	type args struct {
		r        io.Reader
		pkgName  string
		typeName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"String", args{r: strings.NewReader(`"hoge"`), pkgName: "hoge", typeName: "Hoge"}, `package hoge

type Hoge string
`, false},
		{"int64", args{r: strings.NewReader(`1`), pkgName: "hoge", typeName: "Hoge"}, `package hoge

type Hoge int64
`, false},
		{"float64", args{r: strings.NewReader(`0.1`), pkgName: "hoge", typeName: "Hoge"}, `package hoge

type Hoge float64
`, false},
		{"bool", args{r: strings.NewReader(`true`), pkgName: "hoge", typeName: "Hoge"}, `package hoge

type Hoge bool
`, false},
		{"obj", args{r: strings.NewReader(`{"given_name": "bob"}`), pkgName: "hoge", typeName: "Hoge"}, "package hoge\n\ntype Hoge struct {\n	GivenName string `json:\"given_name\"`\n}\n", false},
		{"array", args{r: strings.NewReader(`["name"]`), pkgName: "hoge", typeName: "Hoge"}, `package hoge

type Hoge []string
`, false},
		{"toCamelCase - snake", args{r: strings.NewReader(`{"csv_id": 1}`), pkgName: "hoge", typeName: "Hoge"}, "package hoge\n\ntype Hoge struct {\n	CSVID int64 `json:\"csv_id\"`\n}\n", false},
		{"toCamelCase - camel", args{r: strings.NewReader(`{"csvId": 1}`), pkgName: "hoge", typeName: "Hoge"}, "package hoge\n\ntype Hoge struct {\n	CSVID int64 `json:\"csvId\"`\n}\n", false},
		{"null", args{r: strings.NewReader(`{"status": null}`), pkgName: "hoge", typeName: "Hoge"}, "", true},
		{"fail - invalid syntax", args{r: strings.NewReader(`{`), pkgName: "hoge", typeName: "Hoge"}, "", true},
		{"fail - empty object", args{r: strings.NewReader(`{}`), pkgName: "hoge", typeName: "Hoge"}, "", true},
		{"fail - empty array", args{r: strings.NewReader(`[]`), pkgName: "hoge", typeName: "Hoge"}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsontotype.Exec(tt.args.r, tt.args.pkgName, tt.args.typeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Exec() error = %v`, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Exec() = \n`%v`, want \n`%v`", got, tt.want)
			}
		})
	}
}
