// Copyright (c) 2018 The ciphrtxt developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	//"fmt"
	"io"
	"io/ioutil"
	//"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jadeblaquiere/ctclient/ctgo"
)

// API version constants
const (
	restapiSemverString = "1.0.0"
	restapiSemverMajor  = 1
	restapiSemverMinor  = 0
	restapiSemverPatch  = 0
)

var (
	b64encoding base64.Encoding
)

type restServerConfig struct {
	restListenerPort string
	params           *params
	ms               *ctgo.MessageStore
}

type CtRestServer struct {
	Router *mux.Router
	cfg    *restServerConfig
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithFileOctetStream(w http.ResponseWriter, code int, bfile *os.File) {
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.WriteHeader(code)
	io.Copy(w, bfile)
}

func (ctrs *CtRestServer) getMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mhash, err := hex.DecodeString(vars["msgid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	mf := ctrs.cfg.ms.GetMessage(mhash)
	cfile, err := mf.CiphertextFile()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Message Not Found")
		return
	}
	defer cfile.Close()
	respondWithFileOctetStream(w, http.StatusOK, cfile)
}

func (ctrs *CtRestServer) postMessage(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.TempFile("", "ctmsg")
	if err != nil {
		// failed to create a temp file ? No recovery
		respondWithError(w, http.StatusInternalServerError, "Error receiving messages")
		return
	}

	// todo: preview header and validate, disconnect on header error

	io.Copy(file, r.Body)
	finfo, err := file.Stat()
	if err != nil {
		// can't stat temp file ? No recovery
		respondWithError(w, http.StatusInternalServerError, "Error receiving messages")
		return
	}

	filename := finfo.Name()

	file.Close()
	mf, err := ctgo.NewMessageFile(os.TempDir() + "/" + filename)
	if err != nil {
		// failed to import as MessageFile - bad data
		respondWithError(w, http.StatusBadRequest, "Invalid Message File")
		return
	}

	err = ctrs.cfg.ms.IngestMessageFile(mf)
	if err != nil {
		// path error on Stat() ? No recovery
		respondWithError(w, http.StatusInternalServerError, "Error receiving messages")
		return
	}

	w.WriteHeader(http.StatusOK)
	//fall through respond
}

func (ctrs *CtRestServer) listMessages(w http.ResponseWriter, r *http.Request) {
	hlist, err := ctrs.cfg.ms.ListHashesForInterval(ctgo.UTimeToTime(0), time.Now())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retreiving messages")
		return
	}
	lmr := make([]string, len(hlist))
	for i, h := range hlist {
		lmr[i] = hex.EncodeToString(h)
	}
	respondWithJSON(w, http.StatusOK, lmr)
}

func (ctrs *CtRestServer) initializeRoutes() {
	ctrs.Router.HandleFunc("/messages/{msgid:[0-9abcdefABCDEF]+}", ctrs.getMessage).Methods("GET")
	ctrs.Router.HandleFunc("/messages/", ctrs.listMessages).Methods("GET")
	ctrs.Router.HandleFunc("/messages/", ctrs.postMessage).Methods("POST")
}

func NewCtRestServer(cfg *restServerConfig) (ctrs *CtRestServer) {
	ctrs = new(CtRestServer)
	ctrs.cfg = cfg
	ctrs.Router = mux.NewRouter()
	return ctrs
}

func (ctrs *CtRestServer) Start() {
	rpcsLog.Trace("Starting ciphrtxt REST API server")
	ctrs.initializeRoutes()

}
