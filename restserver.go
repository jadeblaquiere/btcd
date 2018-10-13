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
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
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

	// restAuthTimeoutSeconds is the number of seconds a connection to the
	// REST server is allowed to stay open without authenticating before it
	// is closed.
	restAuthTimeoutSeconds = 10
)

var (
	b64encoding base64.Encoding
)

type restServerConfig struct {
	// Listeners defines a slice of listeners for which the REST server will
	// take ownership of and accept connections.  Since the REST server takes
	// ownership of these listeners, they will be closed when the REST server
	// is stopped.
	Listeners []net.Listener

	MStore *ctgo.MessageStore
}

type ctRestServer struct {
	Router   *mux.Router
	cfg      *restServerConfig
	wg       sync.WaitGroup
	shutdown int32
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

func (ctrs *ctRestServer) getMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	mhash, err := hex.DecodeString(vars["msgid"])
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	mf := ctrs.cfg.MStore.GetMessage(mhash)
	cfile, err := mf.CiphertextFile()
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Message Not Found")
		return
	}
	defer cfile.Close()
	respondWithFileOctetStream(w, http.StatusOK, cfile)
}

func (ctrs *ctRestServer) postMessage(w http.ResponseWriter, r *http.Request) {
	file, err := ioutil.TempFile("", "MStoreg")
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

	err = ctrs.cfg.MStore.IngestMessageFile(mf)
	if err != nil {
		// path error on Stat() ? No recovery
		respondWithError(w, http.StatusInternalServerError, "Error receiving messages")
		return
	}

	w.WriteHeader(http.StatusOK)
	//fall through respond
}

func (ctrs *ctRestServer) listMessages(w http.ResponseWriter, r *http.Request) {
	hlist, err := ctrs.cfg.MStore.ListHashesForInterval(ctgo.UTimeToTime(0), time.Now())
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

func (ctrs *ctRestServer) initializeRoutes() {
	ctrs.Router.HandleFunc("/api/v1/messages/{msgid:[0-9abcdefABCDEF]+}", ctrs.getMessage).Methods("GET")
	ctrs.Router.HandleFunc("/api/v1/messages/", ctrs.listMessages).Methods("GET")
	ctrs.Router.HandleFunc("/api/v1/messages/", ctrs.postMessage).Methods("POST")
}

func newCtRESTServer(cfg *restServerConfig) (ctrs *ctRestServer, err error) {
	ctrs = new(ctRestServer)
	ctrs.cfg = cfg
	ctrs.Router = mux.NewRouter()
	return ctrs, nil
}

func (ctrs *ctRestServer) Start() {
	rpcsLog.Trace("Starting ciphrtxt REST API server")

	httpServer := &http.Server{
		Handler: ctrs.Router,

		// Timeout connections which don't complete the initial
		// handshake within the allowed timeframe.
		ReadTimeout: time.Second * restAuthTimeoutSeconds,
	}

	ctrs.initializeRoutes()

	for _, listener := range ctrs.cfg.Listeners {
		ctrs.wg.Add(1)
		go func(listener net.Listener) {
			rpcsLog.Infof("REST server listening on %s", listener.Addr())
			httpServer.Serve(listener)
			rpcsLog.Tracef("REST listener done for %s", listener.Addr())
			ctrs.wg.Done()
		}(listener)
	}
}

// Stop is used by server.go to stop the rpc listener.
func (ctrs *ctRestServer) Stop() error {
	if atomic.AddInt32(&ctrs.shutdown, 1) != 1 {
		restLog.Infof("REST server is already in the process of shutting down")
		return nil
	}
	restLog.Warnf("REST server shutting down")
	for _, listener := range ctrs.cfg.Listeners {
		err := listener.Close()
		if err != nil {
			rpcsLog.Errorf("Problem shutting down REST listener: %v", err)
			return err
		}
	}
	//close(s.quit)
	ctrs.wg.Wait()
	restLog.Infof("REST server shutdown complete")
	return nil
}
