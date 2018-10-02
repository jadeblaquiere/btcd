// Copyright (c) 2018 The ciphrtxt developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jadeblaquiere/ctclient/ctgo"
)

func executeRequest(req *http.Request, router *mux.Router) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestNewCtRestServer(t *testing.T) {
	dirname, err := ioutil.TempDir("", "ctmstest")
	if err != nil {
		t.Fatalf("Failed to create temp dir, error: %s", err.Error())
	}

	cfg := new(restServerConfig)
	cfg.restListenerPort = ""
	cfg.params = nil
	cfg.ms, err = ctgo.OpenMessageStore(dirname)
	if err != nil {
		t.Fatalf("Failed to open MessageStore, error: %s", err.Error())
	}

	ctrs := NewCtRestServer(cfg)
	if ctrs == nil {
		t.Fatalf("Failed to create CtRestServer")
	}

	numiter := int(10)
	sK := make([]*ctgo.SecretKey, numiter)
	pK := make([]*ctgo.PublicKey, numiter)
	m := make([]*ctgo.Message, len(sK)*len(sK))

	for i := 0; i < numiter; i++ {
		sK[i] = ctgo.NewSecretKey(0, 0, 0, 0, 0)
		pK[i] = sK[i].PublicKey()
	}

	prate := ctgo.NewPostageRate(10, 65536, 1, (64 * 16777216))
	ptxt := []byte("Hello, Alice")

	for i := 0; i < len(sK); i++ {
		for j := 0; j < len(sK); j++ {
			if i == j {
				continue
			}
			m[(i*len(sK))+j] = ctgo.EncryptMessage(pK[j], sK[i], time.Now(), time.Duration(7*24*time.Hour), "", ptxt, prate)
			ctxt := m[(i*len(sK))+j].Ciphertext()
			bodybuf := bytes.NewBuffer(ctxt[:])
			req, _ := http.NewRequest("POST", "/messages/", bodybuf)
			response := executeRequest(req, ctrs.Router)

			checkResponseCode(t, http.StatusOK, response.Code)
		}
	}

	for i := 0; i < len(sK); i++ {
		for j := 0; j < len(sK); j++ {
			if i == j {
				continue
			}
			mhash := m[(i*len(sK))+j].PayloadHash()
			req, _ := http.NewRequest("GET", "/messages/"+hex.EncodeToString(mhash), nil)
			response := executeRequest(req, ctrs.Router)

			checkResponseCode(t, http.StatusOK, response.Code)

			bodybuf := bytes.NewBuffer([]byte{})
			io.Copy(bodybuf, response.Body)
			ctxtcp := bodybuf.Bytes()
			if bytes.Compare(ctxtcp, m[(i*len(sK))+j].Ciphertext()) != 0 {
				t.Errorf("Ciphertext mismatch for i,j = %d,%d\n", i, j)
			}
		}
	}

	req, _ := http.NewRequest("GET", "/messages/", nil)
	response := executeRequest(req, ctrs.Router)

	checkResponseCode(t, http.StatusOK, response.Code)

	jsonbytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("List all messages: error reading response\n")
	}

	fmt.Println("Response : " + string(jsonbytes))

	cfg.ms.Close()
	os.RemoveAll(dirname)
}
