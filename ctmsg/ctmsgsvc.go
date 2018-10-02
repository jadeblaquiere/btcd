// Copyright (c) 2018 The ciphrtxt developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package ctmsg

import (
	"errors"
	//"fmt"
	//"runtime/pprof"

	"github.com/jadeblaquiere/ctclient/ctgo"
)

var (
	//ErrNoRootDir is used to indicate that MessageStoreRootDir must be
	//specified in the configuration.
	ErrNoRootDir = errors.New("Config: Dial cannot be nil")
)

type Config struct {
	MessageStoreRootDir string
	//SectorRing          uint
	//SectorStart         uint
}

type CiphrtxtMsgSvc struct {
	MStore *ctgo.MessageStore
}

func (ctms *CiphrtxtMsgSvc) Close() {
	log.Info("closing ciphrtxt message store database")
	ctms.MStore.Close()
}

func New(cfg *Config) (*CiphrtxtMsgSvc, error) {
	if len(cfg.MessageStoreRootDir) == 0 {
		return nil, ErrNoRootDir
	}
	ms, err := ctgo.OpenMessageStore(cfg.MessageStoreRootDir)
	if err != nil {
		return nil, errors.New("ctmsg:New OpenMessageStore failed : " + err.Error())
	}
	ctms := new(CiphrtxtMsgSvc)
	ctms.MStore = ms
	log.Info("ciphrtxt message store database opened")
	return ctms, nil
}
