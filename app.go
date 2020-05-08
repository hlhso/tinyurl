package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"log"
	"net/http"
)

type App struct {
	Router      *mux.Router //路由，路径和handle绑定
	Middlewares *Middleware //请求预处理中间件
	Config      *Env        //用户环境变量
}

type shortenReq struct {
	//短地址
	URL string `json:"url" validate:"required"`
	//过期时间
	Expira int64 `json:"expiration_in_minutes" validate:"min=0"`
}

type shortlinkResp struct {
	Shortlink string `json:"short_link"`
}

func (a *App) Initapp(e *Env) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	a.Config = e
	a.Router = mux.NewRouter()
	a.Middlewares = &Middleware{}
	a.initRoutes()
}

func (a *App) initRoutes() {
	/*a.Router.HandleFunc("/api/shorten",a.createShortlink).Methods("POST")
	a.Router.HandleFunc("/api/info",a.getShortlinkInfo).Methods("GET")
	a.Router.HandleFunc("/{shortlink:[a-zA-Z0-9]{1,11}}",a.redirect).Methods("GET")*/
	m := alice.New(a.Middlewares.LoggingHandler, a.Middlewares.RecoverHandler)

	a.Router.Handle("/api/shorten", m.ThenFunc(a.createShortlink)).Methods("POST")
	a.Router.Handle("/api/info", m.ThenFunc(a.getShortlinkInfo)).Methods("GET")
	a.Router.Handle("/{shortlink:[a-zA-Z0-9]{1,11}}", m.ThenFunc(a.redirect)).Methods("GET")
}

//创建短地址函数
func (a *App) createShortlink(w http.ResponseWriter, r *http.Request) {
	var req shortenReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, StatusError{
			Code: http.StatusBadRequest,                            //错误状态码
			Err:  fmt.Errorf("parse parameters failed %v", r.Body), //json解析错误
		})
		return
	}
	if err := validator.New().Struct(req); err != nil {
		respondWithError(w, StatusError{
			Code: http.StatusBadRequest,                            //错误状态码
			Err:  fmt.Errorf("validate parameters failed %v", req), //参数验证错误
		})
		return
	}
	defer r.Body.Close()
	//fmt.Printf("%+v\n",req)
	s, err := a.Config.S.Shorten(req.URL, req.Expira)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusCreated, shortlinkResp{Shortlink: s})
	}
}

func (a *App) getShortlinkInfo(w http.ResponseWriter, r *http.Request) {
	vals := r.URL.Query()
	s := vals.Get("shortlink")
	//fmt.Printf("%s\n", s)
	d, err := a.Config.S.ShortLinkInfo(s)
	if err != nil {
		respondWithError(w, err)
	} else {
		respondWithJSON(w, http.StatusOK, d) //200ok
	}
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//fmt.Printf("%s\n", vars["shortlink"])
	u, err := a.Config.S.UnShorten(vars["shortlink"])
	if err != nil {
		respondWithError(w, err)
	} else {
		http.Redirect(w, r, u, http.StatusTemporaryRedirect) //307临时重定向
		return
	}
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}

//响应错误函数
func respondWithError(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case Error: //如果是是自定义的错误类型
		log.Printf("HTTP %d - %s", e.Status(), e) //状态码和错误描述
		respondWithJSON(w, e.Status(), e.Error())
	default: //默认的Internal错误
		respondWithJSON(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

//返回给客户端json信息
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	resp, _ := json.Marshal(payload)
	w.WriteHeader(code) //写入错误码
	w.Header().Set("Context-Type", "application/json")
	w.Write(resp) //写入序列化后的错误描述
}
