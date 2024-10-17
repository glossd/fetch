package main

import (
	"fmt"
	"github.com/glossd/fetch"
)

type IPet interface {
	Name() string
}

type Pet struct {
	Name string
}

type I interface {
	Get() int
}

type S struct {
	X int
}

func (s S) Get() int {
	return s.X
}

func main() {
	j, err := fetch.Unmarshal[fetch.J](`{
    "name": "Jason",
    "category": {
        "name":"dogs"
    },
    "tags": [
        {"name":"briard"}
    ]
}`)
	if err != nil {
		panic(err)
	}

	fmt.Println("Print the whole json:", j)
	fmt.Println("Pet's name is", j.Q(".name"))
	fmt.Println("Pet's category name is", j.Q(".category.name"))
	fmt.Println("First tag's name is", j.Q(".tags[0].name"))

	category := j.Q(".category")
	fmt.Println("Pet's category object", category)
	fmt.Println("Pet's category name is", category.Q(".name"))

	//var a any
	//err := json.Unmarshal([]byte(`[1, "2"]`), &a)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//arr := a.([]any)
	//fmt.Println(reflect.TypeOf(arr[0]))
}

//i := reflect.ValueOf(p).Interface()
//ip := reflect.ValueOf(&i).Elem()
//ivp := reflect.ValueOf(&iv).Elem()
//fmt.Println(ip.IsNil(), ivp.IsNil())
