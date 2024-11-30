<div align="center">
    <img src=".github/fetch.svg" width="250px" height="60" alt="FETCH">
</div>

<div align="center">Go HTTP client. Inspired by the simplicity of <a href="https://github.com/axios/axios">axios</a> and JSON handling in JS. Improved with generics.</div>

## Reasons
1. I was tired of writing `json:"fieldName,omitempty"` alongside each field. This package de-capitalizes the public fields in JSON parsing and omits empty ones unless `json` tag is specified.
2. I always forget all the boilerplate code to make an HTTP request. This package provides a simple one-function call approach.


## Installing
This is a zero-dependency package. It requires Go version 1.21 or above.
```shell
go get github.com/glossd/fetch
```

## Structure
Functions of the `fetch` package match HTTP methods. Each function is generic and its generic type is the response.  

$${\color{lightblue}fetch. \color{lightgreen}Method \color{lightblue}[ \color{cyan}ResponseType \color{lightblue}](url \space string, \space ...) \space returns \space \color{cyan}ResponseType}$$

## Examples

This is the `Pet` object from https://petstore.swagger.io/
```json
{
  "id": 1,
  "name": "Buster",
  "tags": [
    {
      "name": "beagle"
    }
  ]
}
```
### GET request
#### Print response
```go
str, err := fetch.Get[string]("https://petstore.swagger.io/v2/pet/1")
if err != nil {
    //handle err
}
fmt.Println(str)
```

#### Dynamically typed
`fetch.J` is an interface representing arbitrary JSON.
```go
j, err := fetch.Get[fetch.J]("https://petstore.swagger.io/v2/pet/1")
if err != nil {
    panic(err)
}
// access nested values using jq-like patterns
fmt.Println("Pet's name is ", j.Q(".name"))
fmt.Println("First tag's name is ", j.Q(".tags[0].name"))
```
[More about jq-like patterns](#jq-like-queries)
#### Statically typed
```go
type Tag struct {
    Name string
}
type Pet struct {
    Name string
    Tags []Tag
}

pet, err := fetch.Get[Pet]("https://petstore.swagger.io/v2/pet/1")
if err != nil {
    panic(err)
}
fmt.Println("Pet's name is ", pet.Name)
fmt.Println("First tag's name is ", pet.Tags[0].Name) // handle index
```

### POST request
Post, Put and others have an additional argument for the request body of `any` type. 
```go
type Pet struct {
    Name string
}
type IdObj struct {
    Id int
}
obj, err := fetch.Post[IdObj]("https://petstore.swagger.io/v2/pet", Pet{Name: "Lola"})
if err != nil {
    panic(err)
}
fmt.Println("Posted pet's ID ", obj.Id)
```
*Passing `string` or `[]byte` type variable as the second argument will directly add its value to the request body.

### HTTP response status, headers and other attributes
If you need to check the status or headers of the response, you can wrap your response type with `fetch.Response`.
```go
type Pet struct {
    Name string
}
resp, err := fetch.Get[fetch.Response[Pet]]("https://petstore.swagger.io/v2/pet/1")
if err != nil {
    panic(err)
}
if resp.Status == 200 {
    fmt.Println("Found pet with id 1")
    // Response.Body is the Pet object.
    fmt.Println("Pet's name is ", resp.Body.Name)
    fmt.Println("Response headers", resp.Headers)
}
```
If you don't need the HTTP body you can use `fetch.ResponseEmpty`
```go
res, err := fetch.Delete[fetch.ResponseEmpty]("https://petstore.swagger.io/v2/pet/10")
if err != nil {
    panic(err)
}
fmt.Println("Status:", res.Status)
fmt.Println("Headers:", res.Headers)
```
#### Error handling
Any **non-2xx** response status is treated as an **error**!
If the error isn't `nil` it can be safely cast to `*fetch.Error` which will contain the status and other HTTP attributes. 
```go
_, err := fetch.Get[string]("https://petstore.swagger.io/v2/pet/-1")
if err != nil {
    fmt.Printf("Get pet failed: %s\n", err)
    ferr := err.(*fetch.Error)
    fmt.Printf("HTTP status=%d, headers=%v, body=%s", ferr.Status, ferr.Headers, ferr.Body)
}
```
### Make request with Go Context
Request with 5 seconds timeout: 
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
_, err := fetch.Get[string]("https://petstore.swagger.io/v2/pet/1", fetch.Config{Ctx: ctx})
if err != nil {
    panic(err)
}
```

### Request with headers
```go
headers := map[string]string{"Content-type": "text/plain"}
_, err := fetch.Get[string]("https://petstore.swagger.io/v2/pet/1", fetch.Config{Headers: headers})
if err != nil {
    panic(err)
}
```

### Capitalized fields
If you want this package to parse the public fields as capitalized into JSON, you need to add the `json` tag:
```go
type Pet struct {
    Name string `json:"Name"`
}
```

### Arrays
Simple, just pass an array as the response type.
```go
type Pet struct {
    Name string
}
pets, err := fetch.Get[[]Pet]("https://petstore.swagger.io/v2/pet/findByStatus?status=sold")
if err != nil {
    panic(err)
}
fmt.Println("First sold pet ", pets[0]) // handle index out of range
```
Or you can use `fetch.J`
```go
j, err := fetch.Get[fetch.J]("https://petstore.swagger.io/v2/pet/findByStatus?status=sold")
if err != nil {
    panic(err)
}
fmt.Println("First sold pet ", j.Q(".[0]"))
```

### JQ-like queries
`fetch.J` is an interface with `Q` method which provides easy access to any field.   
Method `fetch.J#String()` returns JSON formatted string of the value. 
```go
j := fetch.Parse(`{
    "name": "Jason",
    "category": {
        "name":"dogs"
    },
    "tags": [
        {"name":"briard"}
    ]
}`)

fmt.Println("Print the whole json:", j)
fmt.Println("Pet's name is", j.Q(".name"))
fmt.Println("Pet's category name is", j.Q(".category.name"))
fmt.Println("First tag's name is", j.Q(".tags[0].name"))
``` 
Method `fetch.J#Q` returns `fetch.J`. You can use the method `Q` on the result as well.
```go
category := j.Q(".category")
fmt.Println("Pet's category object", category)
fmt.Println("Pet's category name is", category.Q(".name"))
```

To convert `fetch.J` to a basic value use one of `As*` methods

| J Method  | Return type    |
|-----------|----------------|
| AsObject  | map[string]any |
| AsArray   | []any          |
| AsNumber  | float64        |
| AsString  | string         |
| AsBoolean | bool           |

E.g.
```go
n, ok := fetch.Parse(`{"price": 14.99}`).Q(".price").AsNumber()
if !ok {
    // not a number
}
fmt.Printf("Price: %.2f\n", n) // n is a float64
```

Use `IsNil` to check the value on presence.
```go
if fetch.Parse("{}").Q(".price").IsNil() {
    fmt.Println("key 'price' doesn't exist")
}
// fields of unknown values are nil as well.
if fetch.Parse("{}").Q(".price.cents").IsNil() {
    fmt.Println("'cents' of undefined is fine.")
}
```

## JSON handling
I have patched `encoding/json` package and attached to the `internal` folder, but you can use these functions.
### Marhsal
Use it to convert any object into a string. It's the same as `json.Marshal`
but it treats public struct fields as de-capitalized and omits empty fields by default unless `json` tag is specified. 
```go
str, err := fetch.Marhsal(map[string]string{"key":"value"})
```
### Unmarshalling
Unmarshal will parse the input into the generic type.
```go
type Pet struct {
    Name string
}
p, err := fetch.Unmarshal[Pet](`{"name":"Jason"}`)
if err != nil {
    panic(err)
}
fmt.Println(p.Name)
```

`fetch.Parse` unmarshalls JSON string into `fetch.J`, returning `fetch.Nil` instead of an error,
which allows you to write one-liners. 
```go
fmt.Println(fetch.Parse(`{"name":"Jason"}`).Q(".name"))
```

### Global Setters
You can set base URL path for all requests. 
```go
fetch.SetBaseURL("https://petstore.swagger.io/v2")
pet := fetch.Get[string]("/pets/1")
// you can still call other URLs by passing URL with protocol.
fetch.Get[string]("https://www.google.com")
```
You can set the http.Client for all requests
```go
fetch.SetHttpClient(&http.Client{Timeout: time.Minute})
```

### fetch.Config
Each HTTP method has the configuration option. 
```go
type Config struct {
    // Defaults to context.Background()
    Ctx context.Context
    // Defaults to GET
    Method  string
    Body    string
    Headers map[string]string
}
```

## HTTP handlers
`fetch.ToHandlerFunc` allows you to convert `func(in) out, error` signature function into `http.HandlerFunc`.
It unmarshals the HTTP request body into the function argument, then marshals the returned value into the HTTP response body.
```go
// accepts Pet object and returns Pet object
http.HandleFunc("POST /pets", fetch.ToHandlerFunc(func(in Pet) Pet, error {
    pet, err := savePet(pet)
    if err != nil {
        log.Println("Couldn't create a pet", err)
        return nil, err
    }
    return pet, nil
}))
http.ListenAndServe(":8080", nil)
```
If you don't need request or response body, use `fetch.Empty` to fit the function signature.
```go
http.HandleFunc("GET /pets/1", fetch.ToHandlerFunc(func(_ fetch.Empty) (Pet, error) {
    return Pet{Id: 1}, nil
}))
```
The error format can be customized with the `fetch.SetRespondErrorFormat` global setter.  
To log http errors with your logger call `SetDefaultHandlerConfig`
```go
fetch.SetDefaultHandlerConfig(fetch.HandlerConfig{ErrorHook: func(err error) {
    mylogger.Errorf("fetch http error: %s", err)
}})
```

