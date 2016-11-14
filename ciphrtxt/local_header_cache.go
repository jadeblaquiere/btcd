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
    //"net/http"
    //"io/ioutil"
    "encoding/hex"
    //"encoding/json"
    "fmt"
    "errors"
    "github.com/syndtr/goleveldb/leveldb"
    "github.com/syndtr/goleveldb/leveldb/util"
    "math/rand"
    "strconv"
    "sync"
    "time"
)

const lhcRefreshMinDelay = 10

type peerCache struct {
    hc    *HeaderCache
    lastRefresh   uint32
}

type LocalHeaderCache struct {
    basepath string
    db *leveldb.DB
    syncMutex sync.Mutex
    serverTime uint32
    lastRefresh uint32
    Count int
    Peers []*peerCache
}

// NOTE : if dbpath is empty ("") header cache will be in-memory only

func OpenLocalHeaderCache(filepath string) (lhc *LocalHeaderCache, err error) {
    lhc = new(LocalHeaderCache)
    lhc.basepath = filepath
    
    dbpath := filepath + "/localdb"
    
    if len(dbpath) == 0 {
        return nil, errors.New("refusing to open empty db path")
    }
    
    lhc.db, err = leveldb.OpenFile(dbpath, nil)
    if err != nil {
        return nil, err
    }
    
    err = lhc.recount()
    if err != nil {
        return nil, err
    }
    
    fmt.Printf("LocalHeaderCache open, found %d message headers\n", lhc.Count)
    return lhc, nil
}

func (lhc *LocalHeaderCache) recount() (err error) {
    emptyMessage := "000000000000000000000000000000000000000000000000000000000000000000"
    expiredBegin, err := hex.DecodeString("E" + "00000000" + emptyMessage + "0")
    if err != nil {
        return err
    }
    expiredEnd, err := hex.DecodeString("E" + "FFFFFFFF" + emptyMessage + "0")
    if err != nil {
        return err
    }
    
    iter := lhc.db.NewIterator(&util.Range{Start: expiredBegin,Limit: expiredEnd}, nil)

    count := int(0)
        
    for iter.Next() {
        count += 1
    }
    iter.Release()

    lhc.Count = count
    
    return nil
}

func (lhc *LocalHeaderCache) Close() {
    for _, p := range lhc.Peers {
        if p.hc != nil {
            p.hc.Close()
            p.hc = nil
        }
    }
    
    if lhc.db != nil {
        lhc.db.Close()
        lhc.db = nil
    }
}

func (lhc *LocalHeaderCache) Insert(h *RawMessageHeader) (insert bool, err error) {
    servertime := uint32(time.Now().Unix())
    dbk, err := h.dbKeys(servertime)
    if err != nil {
        return false, err
    }
    _, err = lhc.db.Get(dbk.I, nil)
    if err == nil {
        return false, nil
    }
    value := []byte(h.Serialize())
    //value := h.Serialize()[:]
    batch := new(leveldb.Batch)
    batch.Put(dbk.date, value)
    batch.Put(dbk.servertime, value)
    batch.Put(dbk.expire, value)
    batch.Put(dbk.I, value)
    err = lhc.db.Write(batch, nil)
    if err != nil {
        return false, err
    }
    lhc.Count += 1
    return true, nil
}

func (lhc *LocalHeaderCache) Remove(h *RawMessageHeader) (err error) {
    I, err := hex.DecodeString(h.I)
    if err != nil {
        return err
    }
    value, err := lhc.db.Get(I, nil)
    if err != nil {
        return err
    }
    servertime := deserializeUint32(value[MessageHeaderLengthV2:MessageHeaderLengthV2+4])
    dbk, err := h.dbKeys(servertime)
    if err != nil {
        return err
    }
    batch := new(leveldb.Batch)
    batch.Delete(dbk.date)
    batch.Delete(dbk.servertime)
    batch.Delete(dbk.expire)
    batch.Delete(dbk.I)
    lhc.Count -= 1
    return lhc.db.Write(batch, nil)
}

func (lhc *LocalHeaderCache) FindByI (I []byte) (h *RawMessageHeader, err error) {
    lhc.Sync()

    value, err := lhc.db.Get(I, nil)
    if err != nil {
        return nil, err
    }
    h = new(RawMessageHeader)
    if h.Deserialize(string(value)) == nil {
        return nil, errors.New("retreived invalid header from database")
    }
    return h, nil
}

func (lhc *LocalHeaderCache) FindSince (tstamp uint32) (hdrs []RawMessageHeader, err error) {
    lhc.Sync()

    emptyMessage := "000000000000000000000000000000000000000000000000000000000000000000"
    tag1 := fmt.Sprintf("C%08X%s0", tstamp, emptyMessage)
    tag2 := "C" + "FFFFFFFF" + emptyMessage + "0"
    
    bin1, err := hex.DecodeString(tag1)
    if err != nil {
        return nil, err
    }
    bin2, err := hex.DecodeString(tag2)
    if err != nil {
        return nil, err
    }
    
    iter := lhc.db.NewIterator(&util.Range{Start: bin1, Limit: bin2}, nil)
    
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

func (lhc *LocalHeaderCache) findSector (seg ShardSector) (hdrs []RawMessageHeader, err error) {
    var tag1, tag2, tag3, tag4 string
    var bin1, bin2, bin3, bin4 []byte
    
    start := seg.Start
    ring := seg.Ring

    lhc.Sync()
    
    //fmt.Printf("LocalHeaderCache.findSector %04x, %d\n", start, ring)

    if ((start < 0x0200) || (start > 0x03ff)) {
        return nil, fmt.Errorf("LocalHeaderCache.findSector start value out of range")
    }

    if ((ring < 0) || (ring > 9)) {
        return nil, fmt.Errorf("LocalHeaderCache.findSector ring value out of range")
    }

    ringsz := 512 >> ring
    end := start + ringsz
    
    emptyMessage := "00000000000000000000000000000000000000000000000000000000000000"
    tag1 = fmt.Sprintf("%04X%s", start, emptyMessage)
    
    if end > 0x400 {
        tag2 = fmt.Sprintf("0400%s", emptyMessage)
        tag3 = fmt.Sprintf("0000%s", emptyMessage)
        tag4 = fmt.Sprintf("%04X%s", ((end & 0x03FF) | (0x0200)) , emptyMessage)
        
        //fmt.Printf("fs: tag1 = %s\n", tag1)
        //fmt.Printf("fs: tag2 = %s\n", tag2)
        //fmt.Printf("fs: tag3 = %s\n", tag3)
        //fmt.Printf("fs: tag4 = %s\n", tag4)
        
        bin3, err = hex.DecodeString(tag3)
        if err != nil {
            return nil, err
        }
        bin4, err = hex.DecodeString(tag4)
        if err != nil {
            return nil, err
        }
    } else {
        tag2 = fmt.Sprintf("%04X%s", end, emptyMessage)
        
        //fmt.Printf("fs: tag1 = %s\n", tag1)
        //fmt.Printf("fs: tag2 = %s\n", tag2)
    }
    
    bin1, err = hex.DecodeString(tag1)
    if err != nil {
        return nil, err
    }
    bin2, err = hex.DecodeString(tag2)
    if err != nil {
        return nil, err
    }
    
    iter := lhc.db.NewIterator(&util.Range{Start: bin1, Limit: bin2}, nil)
    
    hdrs = make([]RawMessageHeader, 0)
    for iter.Next() {
        h := new(RawMessageHeader)
        if h.Deserialize(string(iter.Value())) == nil {
            return nil, errors.New("error parsing message")
        }
        hdrs = append(hdrs, *h)
    }
    
    if (end > 0x400) {
        iter := lhc.db.NewIterator(&util.Range{Start: bin3, Limit: bin4}, nil)
    
        for iter.Next() {
            h := new(RawMessageHeader)
            if h.Deserialize(string(iter.Value())) == nil {
                return nil, errors.New("error parsing message")
            }
            hdrs = append(hdrs, *h)
        }
    }
    
    //fmt.Printf("found %d headers\n", len(hdrs))
    return hdrs, nil
}

func (lhc *LocalHeaderCache) FindExpiringAfter (tstamp uint32) (hdrs []RawMessageHeader, err error) {
    lhc.Sync()

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
    
    iter := lhc.db.NewIterator(&util.Range{Start: bin1, Limit: bin2}, nil)
    
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

func (lhc *LocalHeaderCache) getTime() (serverTime uint32, err error) {
    lhc.serverTime = uint32(time.Now().Unix())
    return lhc.serverTime, nil
}

func (lhc *LocalHeaderCache) pruneExpired() (err error) {
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

    iter := lhc.db.NewIterator(&util.Range{Start: expiredBegin,Limit: expiredEnd}, nil)
    batch := new(leveldb.Batch)
    hdr := new(RawMessageHeader)
    
    delCount := int(0)
        
    for iter.Next() {
        value := iter.Value()
        if hdr.Deserialize(string(value)) == nil {
            return errors.New("unable to parse database value")
        }
        servertime := deserializeUint32(value[MessageHeaderLengthV2:MessageHeaderLengthV2+4])
        dbk, err := hdr.dbKeys(servertime)
        if err != nil {
            return err
        }
        batch.Delete(dbk.date)
        batch.Delete(dbk.servertime)
        batch.Delete(dbk.expire)
        batch.Delete(dbk.I)
        delCount += 1
    }
    iter.Release()
    
    err = lhc.db.Write(batch, nil)
    if err == nil {
        lhc.Count -= delCount
        fmt.Printf("LocalHeaderCache: dropping %d message headers\n", delCount)
    }
    
    return err
}

func (lhc *LocalHeaderCache) Sync() (err error) {
    // if "fresh enough" (refreshMinDelay) then simply return
    now := uint32(time.Now().Unix())
    
    if (now - lhc.lastRefresh) < lhcRefreshMinDelay {
        return nil
    }
    
    //should only have a single goroutine sync'ing at a time
    lhc.syncMutex.Lock()
    defer lhc.syncMutex.Unlock()
    
    err = lhc.pruneExpired()
    if err != nil {
        return err
    }
    
    insCount := int(0)
    
    ordinal := rand.Perm(len(lhc.Peers))
    for i := 0; i < len(lhc.Peers) ; i++ {
        p := lhc.Peers[ordinal[i]]
        
        p.hc.Sync()
        
        if p.hc.lastRefresh > p.lastRefresh {
            mhdrs, err := p.hc.FindSince(p.lastRefresh)
            if err != nil {
                return err
            }
    
            insCount := int(0)

            for _, mh := range mhdrs {
                insert, err := lhc.Insert(&mh)
                if err != nil {
                    return err
                }
                if insert {
                    insCount += 1
                }
            }

            fmt.Printf("LocalHeaderCache: inserted %d message headers\n", insCount)    
        }
    }

    lhc.lastRefresh = now

    lhc.Count += insCount
    fmt.Printf("LocalHeaderCache: insert %d message headers\n", insCount)
    
    fmt.Printf("LocalHeaderCache: %d active message headers\n", lhc.Count)

    return nil
}

func (lhc *LocalHeaderCache) AddPeer(host string, port uint16) (err error) {
    dbpath := lhc.basepath + "/remote/" + host + "_" + strconv.Itoa(int(port)) + "/hdb"
    
    pc := new(peerCache)
    
    rhc, err := OpenHeaderCache(host, port, dbpath)
    if err != nil {
        return err
    }
    
    err = rhc.Sync()
    if err != nil {
        return err
    }
    
    mhdrs, err := rhc.FindSince(0)
    if err != nil {
        return err
    }
    
    pc.hc = rhc
    pc.lastRefresh = rhc.lastRefresh
    lhc.Peers = append(lhc.Peers, pc)

    insCount := int(0)

    for _, mh := range mhdrs {
        insert, err := lhc.Insert(&mh)
        if err != nil {
            return err
        }
        if insert {
            insCount += 1
        }
    }
    
    fmt.Printf("LocalHeaderCache: inserted %d message headers\n", insCount)

    lhc.recount()
    
    fmt.Printf("LocalHeaderCache: %d active message headers\n", lhc.Count)

    return nil
}

