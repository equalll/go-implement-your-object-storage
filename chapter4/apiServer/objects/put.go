package objects

import (
	"../../lib/es"
	"../heartbeat"
	"../locate"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func storeData(w http.ResponseWriter, r *http.Request, object string) {
	s := heartbeat.ChooseRandomDataServer()
	if s == "" {
		log.Println("cannot find any dataServer")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	request, e := http.NewRequest("PUT", "http://"+s+"/objects/"+object, r.Body)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	client := http.Client{}
	nr, e := client.Do(request)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(nr.StatusCode)
	io.Copy(w, nr.Body)
}

func put(w http.ResponseWriter, r *http.Request) {
	digest := r.Header.Get("digest")
	if len(digest) < 9 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if digest[:8] != "SHA-256=" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hash := digest[8:]

	object := url.PathEscape(hash)
	s := locate.Locate(object)
	if s == "" {
		storeData(w, r, object)
	}

	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	version, _, e := es.SearchLatestVersion(name)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	version += 1
	size, e := strconv.ParseInt(r.Header.Get("content-length"), 0, 64)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	e = es.PutVersion(name, version, size, hash)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}