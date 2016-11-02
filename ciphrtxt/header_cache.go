// Copyright (c) 2016, Joseph deBlaquiere <jadeblaquiere@yahoo.com>
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice,
//   this list of conditions and the following disclaimer in the documentation
//   and/or other materials provided with the distribution.
//
// * Neither the name of ciphrtxt nor the names of its
//   contributors may be used to endorse or promote products derived from
//   this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
// FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
// DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
// SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
// CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
// OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package ciphrtxt

import (
    "net/http"
    "io/ioutil"
    "encoding/hex"
    "encoding/json"
    "fmt"
    "errors"
    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/util"
    "strconv"
    "sync"
    "time"
)

const apiStatus string = "api/status/"
const apiTime string = "api/time/"
const apiHeadersSince string = "api/header/list/since/"

const refreshMinDelay = 10

// {"pubkey": "030b5a7b432ec22920e20063cb16eb70dcb62dfef28d15eb19c1efeec35400b34b", "storage": {"max_file_size": 268435456, "capacity": 137438953472, "messages": 6252, "used": 17828492}}

type StatusStorageResponse struct {
    Messages int `json:"messages"`
    Maxfilesize int `json:"max_file_size"`
    Capacity int `json:"capacity"`
    Used int `json:"used"`
}

type StatusResponse struct {
    Pubkey string `json:"pubkey"`
    Status StatusStorageResponse `json:"storage"`
}

type TimeResponse struct {
    Time int `json:"time"`
}

type HeaderListResponse struct {
    Headers []string `json:"header_list"`
}

type HeaderCache struct {
    baseurl string
    db *leveldb.DB
    syncMutex sync.Mutex
    status StatusResponse
    serverTime uint32
    lastRefresh uint32
    Count int
}

// NOTE : if dbpath is empty ("") header cache will be in-memory only

func OpenHeaderCache(host string, port uint16, dbpath string) (hc *HeaderCache, err error) {
    hc = new(HeaderCache)
    hc.baseurl = fmt.Sprintf("http://%s:%d/", host, port)
    
    res, err := http.Get(hc.baseurl + apiStatus)
    if err != nil {
        return nil, err
    }
    
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }
    
    err = json.Unmarshal(body, &hc.status)
    if err != nil {
        return nil, err
    }
    
    if len(dbpath) == 0 {
        return nil, errors.New("refusing to open empty db path")
    }
    
    hc.db, err = leveldb.OpenFile(dbpath, nil)
    if err != nil {
        return nil, err
    }
    
    emptyMessage := "000000000000000000000000000000000000000000000000000000000000000000"
    expiredBegin, err := hex.DecodeString("E" + "00000000" + emptyMessage + "0")
    if err != nil {
        return nil, err
    }
    expiredEnd, err := hex.DecodeString("E" + "FFFFFFFF" + emptyMessage + "0")
    if err != nil {
        return nil, err
    }
    
    iter := hc.db.NewIterator(&util.Range{Start: expiredBegin,Limit: expiredEnd}, nil)

    count := int(0)
        
    for iter.Next() {
        count += 1
    }
    iter.Release()

    hc.Count = count
    fmt.Printf("open, found %d message headers\n", count)
    return hc, nil
}

func (hc *HeaderCache) Close() {
    if hc.db != nil {
        hc.db.Close()
        hc.db = nil
    }
}

type dbkeys struct {
    date []byte
    expire []byte
    I []byte
}

func (h *RawMessageHeader) dbKeys() (dbk *dbkeys, err error) {
    dbk = new(dbkeys)
    dbk.date, err = hex.DecodeString(fmt.Sprintf("D%08X%s0", h.time, h.I))
    if err != nil {
        return nil, err
    }
    dbk.expire, err = hex.DecodeString(fmt.Sprintf("E%08X%s0", h.expire, h.I))
    if err != nil {
        return nil, err
    }
    dbk.I, err = hex.DecodeString(h.I)
    if err != nil {
        return nil, err
    }
    return dbk, err
}

func (hc *HeaderCache) Insert(h *RawMessageHeader) (insert bool, err error) {
    dbk, err := h.dbKeys()
    if err != nil {
        return false, err
    }
    _, err = hc.db.Get(dbk.I, nil)
    if err == nil {
        return false, nil
    }
    value := []byte(h.Serialize())
    //value := h.Serialize()[:]
    batch := new(leveldb.Batch)
    batch.Put(dbk.date, value)
    batch.Put(dbk.expire, value)
    batch.Put(dbk.I, value)
    err = hc.db.Write(batch, nil)
    if err != nil {
        return false, err
    }
    return true, nil
}

func (hc *HeaderCache) Remove(h *RawMessageHeader) (err error) {
    dbk, err := h.dbKeys()
    if err != nil {
        return err
    }
    batch := new(leveldb.Batch)
    batch.Delete(dbk.date)
    batch.Delete(dbk.expire)
    batch.Delete(dbk.I)
    return hc.db.Write(batch, nil)
}

func (hc *HeaderCache) FindByI (I []byte) (h *RawMessageHeader, err error) {
    hc.Sync()

    value, err := hc.db.Get(I, nil)
    if err != nil {
        return nil, err
    }
    h = new(RawMessageHeader)
    if h.Deserialize(string(value)) == nil {
        return nil, errors.New("retreived invalid header from database")
    }
    return h, nil
}

func (hc *HeaderCache) FindSince (tstamp uint32) (hdrs []RawMessageHeader, err error) {
    hc.Sync()

    emptyMessage := "000000000000000000000000000000000000000000000000000000000000000000"
    tag1 := fmt.Sprintf("D%08X%s0", tstamp, emptyMessage)
    tag2 := "D" + "FFFFFFFF" + emptyMessage + "0"
    
    bin1, err := hex.DecodeString(tag1)
    if err != nil {
        return nil, err
    }
    bin2, err := hex.DecodeString(tag2)
    if err != nil {
        return nil, err
    }
    
    iter := hc.db.NewIterator(&util.Range{Start: bin1, Limit: bin2}, nil)
    
    hdrs = make([]RawMessageHeader, 0)
    for iter.Next() {
        h := new(RawMessageHeader)
        if h.Deserialize(string(iter.Value())) == nil {
            return nil, errors.New("error parsing message")
        }
        hdrs = append(hdrs, *h)
    }
    return hdrs, nil
}

func (hc *HeaderCache) FindExpiringAfter (tstamp uint32) (hdrs []RawMessageHeader, err error) {
    hc.Sync()

    emptyMessage := "000000000000000000000000000000000000000000000000000000000000000000"
    tag1 := fmt.Sprintf("E%08X%s0", tstamp, emptyMessage)
    tag2 := "E" + "FFFFFFFF" + emptyMessage + "0"
    
    bin1, err := hex.DecodeString(tag1)
    if err != nil {
        return nil, err
    }
    bin2, err := hex.DecodeString(tag2)
    if err != nil {
        return nil, err
    }
    
    iter := hc.db.NewIterator(&util.Range{Start: bin1, Limit: bin2}, nil)
    
    hdrs = make([]RawMessageHeader, 0)
    for iter.Next() {
        h := new(RawMessageHeader)
        if h.Deserialize(string(iter.Value())) == nil {
            return nil, errors.New("error parsing message")
        }
        hdrs = append(hdrs, *h)
    }
    return hdrs, nil
}

func (hc *HeaderCache) getTime() (serverTime uint32, err error) {
    var tr TimeResponse

    res, err := http.Get(hc.baseurl + apiTime)
    if err != nil {
        return 0, err
    }
    
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return 0, err
    }
    
    err = json.Unmarshal(body, &tr)
    if err != nil {
        return 0, err
    }
    
    hc.serverTime = uint32(tr.Time)
    return hc.serverTime, nil
}

func (hc *HeaderCache) getHeadersSince(since uint32) (mh []RawMessageHeader, err error) {
    res, err := http.Get(hc.baseurl + apiHeadersSince + strconv.FormatInt(int64(since),10))
    if err != nil {
        return nil, err
    }
    
    body, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }
    
    s := new(HeaderListResponse)
    err = json.Unmarshal(body, &s)
    if err != nil {
        return nil, err
    }
    
    mh = make([]RawMessageHeader, 0)
    for _, hdr := range s.Headers {
        h := new(RawMessageHeader)
        if h.Deserialize(hdr) == nil {
            return nil, errors.New("error parsing message")
        }
        mh = append(mh, *h)
    }
    return mh, nil
}

func (hc *HeaderCache) pruneExpired() (err error) {
    emptyMessage := "000000000000000000000000000000000000000000000000000000000000000000"
    expiredBegin, err := hex.DecodeString("E" + "00000000" + emptyMessage + "0")
    if err != nil {
        return err
    }
    now := strconv.FormatUint(uint64(time.Now().Unix()),16)
    expiredEnd, err := hex.DecodeString("E" + now + emptyMessage + "0")
    if err != nil {
        return err
    }

    iter := hc.db.NewIterator(&util.Range{Start: expiredBegin,Limit: expiredEnd}, nil)
    batch := new(leveldb.Batch)
    hdr := new(RawMessageHeader)
    
    delCount := int(0)
        
    for iter.Next() {
        if hdr.Deserialize(string(iter.Value())) == nil {
            return errors.New("unable to parse database value")
        }
        dbk, err := hdr.dbKeys()
        if err != nil {
            return err
        }
        batch.Delete(dbk.date)
        batch.Delete(dbk.expire)
        batch.Delete(dbk.I)
        delCount += 1
    }
    iter.Release()
    
    err = hc.db.Write(batch, nil)
    if err == nil {
        hc.Count -= delCount
        fmt.Printf("dropping %d message headers\n", delCount)
    }
    
    return err
}

func (hc *HeaderCache) Sync() (err error) {
    // if "fresh enough" (refreshMinDelay) then simply return
    now := uint32(time.Now().Unix())
    
    if (now - hc.lastRefresh) < refreshMinDelay {
        return nil
    }
    
    //should only have a single goroutine sync'ing at a time
    hc.syncMutex.Lock()
    defer hc.syncMutex.Unlock()
    
    serverTime, err := hc.getTime()
    if err != nil {
        return err
    }
    
    err = hc.pruneExpired()
    if err != nil {
        return err
    }
    
    mhdrs, err := hc.getHeadersSince(hc.lastRefresh)
    if err != nil {
        return err
    }
    
    insCount := int(0)

    for _, mh := range mhdrs {
        insert, err := hc.Insert(&mh)
        if err != nil {
            return err
        }
        if insert {
            insCount += 1
        }
    }

    hc.lastRefresh = serverTime

    hc.Count += insCount
    fmt.Printf("insert %d message headers\n", insCount)

    return nil
}

