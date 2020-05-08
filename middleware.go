package main

import (
	"log"
	"net/http"
	"time"
)

type Middleware struct{
	
}

//记录请求消耗的时间
func (m Middleware)LoggingHandler(next http.Handler)http.Handler  {
	fn:= func(w http.ResponseWriter,r *http.Request) {
		t1:=time.Now()
		next.ServeHTTP(w,r)
		t2:=time.Now()
		log.Printf("[%s] %q %v",r.Method,r.URL.String(),t2.Sub(t1))
	}
	return http.HandlerFunc(fn)
}

//从panic中恢复过来
func (m Middleware)RecoverHandler(next http.Handler) http.Handler  {
	fn:= func(w http.ResponseWriter,r *http.Request) {
		defer func(){
			if err :=recover();err!=nil{
				log.Printf("Recover from panic: %+v",err)
				http.Error(w ,http.StatusText(500),500)
			}
		}()
		next.ServeHTTP(w,r)
	}
	return http.HandlerFunc(fn)
}