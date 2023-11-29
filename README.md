# Vi Router

[![pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit)](https://github.com/pre-commit/pre-commit)
[![Go Reference](https://pkg.go.dev/badge/github.com/diontr00/vi.svg)](https://pkg.go.dev/github.com/diontr00/vi)
![ci workflow](https://github.com/diontr00/vi/actions/workflows/ci.yml/badge.svg)
[![codecov](https://codecov.io/gh/diontr00/vi/graph/badge.svg?token=bPz6VDXHae)](https://codecov.io/gh/diontr00/vi)

"This project draws inspiration from **gorilla/mux**, and I appreciate
the ease of working with the library.However, its performance
falls short of my expectations. After conducting research, my goal
is to develop a high-performance HTTP router that not only
outperforms but also retains the convenient API of mux,
enhanced with additional support for regex. I welcome any feedback
as this will be my first open source projects."

## Installation

```go
go get -u github.com/diontr00/vi
```

## TestRegex

If you want to make sure whether the regex matcher work as you expected , you can use **TestMatcher** function in the init function

```go
func init() {
    RegisterHelper("ip", `((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|
    [1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])`)
    RegisterHelper("phone", `([+]?[\s0-9]+)?(\d{3}|[(]?[0-9]+[)])?([-]?[\s]?[0-9])+`)

    // Expect match
    errs := TestMatcher(true, "/user/{id:[0-9]+}/:ip/:phone?", map[vi.TestUrl]vi.TestResult{
        "/user/101/192.168.0.1/999-9999999":
        { "id" : "101" , "ip": "192.168.0.1", "phone": "999-9999999"},
        "/user/102/192.168.0.2/+48(12)504-203-260":
        { "id" : "102" , "ip": "192.168.0.2", "phone": "+48(12)504-203-260"},
        "/user/103/192.168.0.3":
        { "id" : "103" , "ip": "192.168.0.3"},
        "/user/104/192.168.0.4/555-5555-555" :
        { "id" : "104" , "ip": "192.168.0.4" ,"phone": "555-5555-555"}
 })

    for i := range errs {
        fmt.Println(errs[i])
    }

    // Expect not match
    errs := TestMatcher(false , ....)
}
```

## Simple Usage

```go
package main

import (
    "github.com/diontr00/vi"
)

func main(){
    customNotFoundHandler  := func(w http.ResponseWriter , r *http.Request) {
   w.WriteHeader(http.StatusNotFound)
   w.Write([]byte(customNotFoundMsg))
    }

    mux := vi.New(&vi.Config{Banner : false ,  NotFoundHandler: customNotFoundHandler})
    mux.RegisterHelper("ip" , `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`)
    mux.Get("/location/:ip", func(w http.ResponseWriter , r *http.Request){
            ip := r.Context().Value("ip").(string)
            msg := fmt.Sprintf("You have search ip addres %s \n", ip)
            w.Write([]byte(msg))
    })

    srv :=  &http.Server{
        Addre: ":8080" ,
        Handler: vi,
    }

    done := make(chan os.Signal, 1)
 signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

 go func() {
  if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
   log.Fatalf("listen: %s\n", err)
  }
 }()
 log.Print("Server Started")

 <-done
 log.Print("Server Stopped")

 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer func() {
  // extra handling here
  cancel()
 }()

 if err := srv.Shutdown(ctx); err != nil {
  log.Fatalf("Server Shutdown Failed:%+v", err)
 }
 log.Print("Server Exited Properly")
}




```

## Matching Rule

- **Named parameter**

  - **Syntax:** :name
  - **Example:** /student/:name
  - **Explain:** Match name as word

- **Name with regex pattern**

  - **Syntax:** {name:regex-pattern}
  - **Example:** /student/{id:[0-9]+}
  - **Explain:** Match id as number

- **Helper pattern**
  - **:id** : short for **/student/{id:[0-9]+}**
  - **:name** : short for **/{name:[0-9a-zA-Z]+}**

## Benchmark

Run benchmark and test with ginkgo:

```
ginkgo
```

![](https://i.imgur.com/sxkEBvu.png)
