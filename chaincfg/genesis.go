// Copyright (c) 2014-2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chaincfg

import (
	"time"

	"github.com/jadeblaquiere/ctcd/chaincfg/chainhash"
	"github.com/jadeblaquiere/ctcd/ciphrtxt"
	"github.com/jadeblaquiere/ctcd/wire"
)

// ctindigoGenesisCoinbaseTx is the coinbase transaction for the genesis blocks
// for the ciphrtxt indigo network.
var ctindigoGenesisCoinbaseTx = wire.MsgTx{
    Version: 1,
    TxIn: []*wire.TxIn{
        {
            PreviousOutPoint: wire.OutPoint{
                Hash:  chainhash.Hash{},
                Index: 0xffffffff,
            },
            SignatureScript: []byte{
                0x04, 0xff, 0x7f, 0x00, 0x1f, 0x01, 0x04, 0x4c, /* |......L| */
                0x55, 0x4e, 0x65, 0x77, 0x20, 0x59, 0x6f, 0x72, /* |UNew Yor| */
                0x6b, 0x20, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x20, /* |k Times | */
                0x32, 0x32, 0x2f, 0x53, 0x65, 0x70, 0x2f, 0x32, /* |22/Sep/2| */
                0x30, 0x31, 0x36, 0x20, 0x59, 0x61, 0x68, 0x6f, /* |016 Yaho| */
                0x6f, 0x20, 0x53, 0x61, 0x79, 0x73, 0x20, 0x48, /* |o Says H| */
                0x61, 0x63, 0x6b, 0x65, 0x72, 0x73, 0x20, 0x53, /* |ackers S| */
                0x74, 0x6f, 0x6c, 0x65, 0x20, 0x44, 0x61, 0x74, /* |tole Dat| */
                0x61, 0x20, 0x6f, 0x6e, 0x20, 0x35, 0x30, 0x30, /* |a on 500| */
                0x20, 0x4d, 0x69, 0x6c, 0x6c, 0x69, 0x6f, 0x6e, /* | Million| */
                0x20, 0x55, 0x73, 0x65, 0x72, 0x73, 0x20, 0x69, /* | Users i| */
                0x6e, 0x20, 0x32, 0x30, 0x31, 0x34, /* |n 2014| */
            },
            Sequence: 0xffffffff,
        },
    },
    TxOut: []*wire.TxOut{
        {
            Value: 0x17d7840000,
            PkScript: []byte{
                0x41, 0x04, 0xd2, 0x19, 0x60, 0x6a, 0xc1, 0x4a, /* |A...`j.J| */
                0xfe, 0x9a, 0x5a, 0xbd, 0xac, 0x33, 0x06, 0xf4, /* |..Z..3..| */
                0x2d, 0x15, 0xb7, 0x93, 0x77, 0x67, 0xde, 0x40, /* |-...wg.@| */
                0x01, 0x3a, 0x9e, 0x74, 0x25, 0xfe, 0xb7, 0x9f, /* |.:.t%...| */
                0xdf, 0x30, 0x91, 0x16, 0x44, 0x5d, 0x0a, 0xf3, /* |.0..D]..| */
                0xe6, 0x97, 0x96, 0x35, 0x87, 0x70, 0xda, 0x76, /* |...5.p.v| */
                0xbb, 0xe6, 0xd2, 0x95, 0xac, 0x3e, 0x2b, 0x2a, /* |.....>+*| */
                0x19, 0xc7, 0x9e, 0x28, 0x70, 0x67, 0x64, 0x7d, /* |...(pgd}| */
                0xd5, 0x7a, 0xac, /* |.z.| */
            },
        },
    },
    LockTime: 0,
}

// ctindigoGenesisHash is the hash of the first block in the block chain for the
// ciphrtxt indigo network (genesis block).
var ctindigoGenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
    0x18, 0x03, 0x6c, 0xb6, 0x5c, 0x16, 0x67, 0x73, 
    0xbd, 0xe0, 0xb9, 0x91, 0x43, 0x97, 0x9b, 0x22, 
    0xfb, 0x6e, 0xc9, 0x55, 0xa2, 0x9b, 0x18, 0xe8, 
    0x6a, 0xcf, 0xc2, 0x3d, 0xd8, 0x46, 0x00, 0x00, 
})

// ctindigoGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the ciphrtxt indigo network.
var ctindigoGenesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
    0x3e, 0x1c, 0x1a, 0xcd, 0x2a, 0xf5, 0x73, 0xb6, 
    0x22, 0x89, 0xf8, 0x65, 0x74, 0xc3, 0x5f, 0x47, 
    0x51, 0x73, 0xa6, 0x8c, 0x6a, 0xdb, 0x74, 0xec, 
    0x4b, 0x02, 0x47, 0x16, 0x57, 0x53, 0x3c, 0xea, 
})

// ctindigoGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the ciphrtxt indigo network.
var ctindigoGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    101,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: ctindigoGenesisMerkleRoot,        // ea3c53571647024bec74db6a8ca67351475fc37465f88922b673f52acd1a1c3e
		Timestamp:  time.Unix(0x580794a6, 0), // Wed Oct 19 15:43:34 2016
		Bits:       0x1f007fff,               // 520126463 [00007fff00000000000000000000000000000000000000000000000000000000]
     NonceHeaderA: [ciphrtxt.MessageHeaderLengthV2]byte{
         0x4d, 0x02, 0x00, 0x00, 0x58, 0x07, 0x9c, 0x51, /* |M...X..Q| */
         0x58, 0x10, 0xd6, 0xd1, 0x02, 0x53, 0xcf, 0x21, /* |X....S.!| */
         0xb5, 0x07, 0x66, 0x7e, 0x90, 0x39, 0x3e, 0xd0, /* |..f~.9>.| */
         0xf3, 0x29, 0x9f, 0x51, 0xfa, 0x88, 0x75, 0x89, /* |.).Q..u.| */
         0x2b, 0x47, 0x34, 0xa0, 0x4c, 0xe5, 0x84, 0xff, /* |+G4.L...| */
         0x99, 0x61, 0x93, 0x65, 0xd2, 0x03, 0x44, 0x5f, /* |.a.e..D_| */
         0xc8, 0x3e, 0x4c, 0x68, 0x4a, 0x8c, 0x26, 0x30, /* |.>LhJ.&0| */
         0xbf, 0x8f, 0xfd, 0x70, 0x42, 0x73, 0x94, 0x68, /* |...pBs.h| */
         0xfa, 0x48, 0xd9, 0x9f, 0xd7, 0x9b, 0x31, 0xff, /* |.H....1.| */
         0x2a, 0xcb, 0xb6, 0x09, 0x64, 0x84, 0x02, 0x68, /* |*...d..h| */
         0x60, 0x7b, 0xa5, 0x20, 0x93, 0x0f, 0x52, 0x1e, /* |`{. ..R.| */
         0xb8, 0x12, 0x42, 0x44, 0x6b, 0xaf, 0x43, 0xa5, /* |..BDk.C.| */
         0x94, 0xb3, 0xab, 0x61, 0xd5, 0x20, 0x26, 0xfa, /* |...a. &.| */
         0xec, 0xf2, 0xb8, 0xa2, 0xef, 0xa2, 0x8b, 0x9f, /* |........| */
         0x4c, 0x91, 0xc5, 0x12, 0x83, 0xbe, 0xbd, 0x46, /* |L......F| */
         0x98, 0xb6, 0x83, 0xc9, 0x58, 0xc0, 0x85, 0xf9, /* |....X...| */
         0xe1, 0x39, 0xe8, 0x9f, 0x15, 0xe7, 0xb0, 0x6e, /* |.9.....n| */
         0xbb, 0xf0, 0x63, 0x37, 0xeb, 0x5e, 0x63, 0x17, /* |..c7.^c.| */
         0xb2, 0x0b, 0xdb, 0x65, 0x8c, 0x67, 0x5f, 0x4d, /* |...e.g_M| */
         0xfa, 0x6f, 0x9d, 0x05, 0x8e, 0x26, 0xf4, 0x49, /* |.o...&.I| */
         0x09, 0x70, 0x27, 0x4b, 0x30, 0xa6, 0xad, 0x39, /* |.p'K0..9| */
         0x31, 0x78, 0xce, 0xf0, 0x89, 0xce, 0xdf, 0x00, /* |1x......| */
         0x00, 0x00, 0xe5, 0x13, /* |....| */
     },
     NonceHeaderB: [ciphrtxt.MessageHeaderLengthV2]byte{
         0x4d, 0x02, 0x00, 0x00, 0x58, 0x06, 0x4c, 0x8a, /* |M...X.L.| */
         0x58, 0x0f, 0x87, 0x0a, 0x03, 0x33, 0x68, 0x94, /* |X....3h.| */
         0x4b, 0x35, 0x1a, 0x4b, 0xe0, 0x66, 0xe5, 0xdc, /* |K5.K.f..| */
         0xae, 0xbf, 0x51, 0xf1, 0x3d, 0x1c, 0x10, 0xc4, /* |..Q.=...| */
         0x9a, 0x09, 0xf0, 0xce, 0xab, 0xf0, 0xea, 0x45, /* |.......E| */
         0x08, 0x10, 0x1c, 0xcf, 0xb7, 0x03, 0x0d, 0xcd, /* |........| */
         0x35, 0x4f, 0xd1, 0xfb, 0x2b, 0x95, 0x2d, 0x41, /* |5O..+.-A| */
         0x1b, 0xfb, 0x68, 0xc1, 0x97, 0xe0, 0xf0, 0xab, /* |..h.....| */
         0xe8, 0xce, 0x7e, 0x91, 0x83, 0x2f, 0x14, 0x54, /* |..~../.T| */
         0xf1, 0x35, 0xde, 0x35, 0x4b, 0xcf, 0x02, 0x9d, /* |.5.5K...| */
         0x29, 0x2e, 0xfe, 0x34, 0x64, 0xdd, 0xe5, 0x5d, /* |)..4d..]| */
         0xf3, 0x65, 0x9e, 0x8d, 0xcc, 0x4b, 0xde, 0x5c, /* |.e...K.\| */
         0xe8, 0x5b, 0x96, 0x37, 0xd0, 0x72, 0x6b, 0x21, /* |.[.7.rk!| */
         0x29, 0x06, 0x74, 0xcc, 0x41, 0x57, 0xfe, 0x6a, /* |).t.AW.j| */
         0xfe, 0x0f, 0x55, 0x75, 0xe8, 0x78, 0x68, 0xd2, /* |..Uu.xh.| */
         0x8e, 0xc8, 0x43, 0xeb, 0x21, 0x61, 0xa0, 0x31, /* |..C.!a.1| */
         0x67, 0xdc, 0x75, 0x13, 0x49, 0x97, 0xa0, 0xba, /* |g.u.I...| */
         0xc6, 0x55, 0x0a, 0x8a, 0xb2, 0x4b, 0xfd, 0x5c, /* |.U...K.\| */
         0x1a, 0xc8, 0x28, 0x7a, 0x07, 0xa2, 0xbf, 0xf2, /* |..(z....| */
         0xcd, 0xe2, 0xae, 0x43, 0x63, 0x71, 0xa9, 0xe9, /* |...Ccq..| */
         0xfa, 0xc2, 0xe1, 0x25, 0x05, 0xe1, 0x20, 0xab, /* |...%.. .| */
         0xce, 0x13, 0xd6, 0x08, 0xae, 0xff, 0x4f, 0x00, /* |......O.| */
         0x00, 0x00, 0xe2, 0xfe, /* |....| */
     },
	},
	Transactions: []*wire.MsgTx{&ctindigoGenesisCoinbaseTx},
}


// ctredGenesisCoinbaseTx is the coinbase transaction for the genesis blocks
// for the ciphrtxt red network.
var ctredGenesisCoinbaseTx = wire.MsgTx{
    Version: 1,
    TxIn: []*wire.TxIn{
        {
            PreviousOutPoint: wire.OutPoint{
                Hash:  chainhash.Hash{},
                Index: 0xffffffff,
            },
            SignatureScript: []byte{
                0x04, 0xff, 0xff, 0x07, 0x1f, 0x01, 0x04, 0x3f, /* |.......?| */
                0x54, 0x68, 0x65, 0x20, 0x54, 0x69, 0x6d, 0x65, /* |The Time| */
                0x73, 0x20, 0x32, 0x33, 0x2f, 0x41, 0x70, 0x72, /* |s 23/Apr| */
                0x2f, 0x32, 0x30, 0x31, 0x36, 0x20, 0x46, 0x42, /* |/2016 FB| */
                0x49, 0x20, 0x65, 0x6e, 0x64, 0x73, 0x20, 0x73, /* |I ends s| */
                0x74, 0x61, 0x6e, 0x64, 0x2d, 0x6f, 0x66, 0x66, /* |tand-off| */
                0x20, 0x77, 0x69, 0x74, 0x68, 0x20, 0x41, 0x70, /* | with Ap| */
                0x70, 0x6c, 0x65, 0x20, 0x6f, 0x76, 0x65, 0x72, /* |ple over| */
                0x20, 0x69, 0x50, 0x68, 0x6f, 0x6e, 0x65, /* | iPhone| */
            },
            Sequence: 0xffffffff,
        },
    },
    TxOut: []*wire.TxOut{
        {
            Value: 0x17d7840000,
            PkScript: []byte{
                0x41, 0x04, 0xc1, 0x40, 0x4e, 0xaa, 0x79, 0xd6, /* |A..@N.y.| */
                0x4a, 0x1b, 0x81, 0xe5, 0xcd, 0x76, 0x5f, 0xe8, /* |J....v_.| */
                0x2a, 0xfb, 0x6a, 0x33, 0x9a, 0xb2, 0x62, 0x48, /* |*.j3..bH| */
                0x57, 0x1a, 0x83, 0x76, 0x98, 0x48, 0x8b, 0xa6, /* |W..v.H..| */
                0xba, 0xc4, 0xe9, 0x1d, 0x5d, 0x65, 0x4d, 0xa3, /* |....]eM.| */
                0xd0, 0x5b, 0x97, 0x7a, 0x52, 0xd8, 0x6c, 0x4e, /* |.[.zR.lN| */
                0x78, 0x58, 0x92, 0xeb, 0xd9, 0xec, 0xe2, 0xd1, /* |xX......| */
                0xc2, 0xcd, 0x2e, 0xab, 0x42, 0x36, 0x47, 0x7a, /* |....B6Gz| */
                0x78, 0xea, 0xac, /* |x..| */
            },
        },
    },
    LockTime: 0,
}

// ctredGenesisHash is the hash of the first block in the block chain for the
// ciphrtxt red network (genesis block).
var ctredGenesisHash = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
    0x0a, 0x53, 0xa8, 0x70, 0x68, 0x03, 0xee, 0x37, 
    0xbb, 0x12, 0x7c, 0x00, 0x0f, 0x6c, 0x18, 0x9a, 
    0x60, 0x71, 0x96, 0x9e, 0xa3, 0xf7, 0xd1, 0x09, 
    0x51, 0xb0, 0xf7, 0x11, 0x2e, 0x27, 0x06, 0x00, 
})

// ctredGenesisMerkleRoot is the hash of the first transaction in the genesis
// block for the ciphrtxt red network.
var ctredGenesisMerkleRoot = chainhash.Hash([chainhash.HashSize]byte{ // Make go vet happy.
    0xee, 0xc9, 0x63, 0xc1, 0xad, 0xef, 0x6a, 0x6b, 
    0x10, 0x9a, 0x2c, 0x53, 0xb2, 0x20, 0x9b, 0x56, 
    0xe0, 0x2b, 0xac, 0xb9, 0x05, 0xb4, 0xf2, 0xe0, 
    0x6c, 0x84, 0x4b, 0x2d, 0x5f, 0x65, 0x6f, 0x32, 
})

// ctredGenesisBlock defines the genesis block of the block chain which serves
// as the public transaction ledger for the ciphrtxt red network.
var ctredGenesisBlock = wire.MsgBlock{
	Header: wire.BlockHeader{
		Version:    101,
		PrevBlock:  chainhash.Hash{},         // 0000000000000000000000000000000000000000000000000000000000000000
		MerkleRoot: ctredGenesisMerkleRoot,        // 326f655f2d4b846ce0f2b405b9ac2be0569b20b2532c9a106b6aefadc163c9ee
		Timestamp:  time.Unix(0x57fedf94, 0), // Thu Oct 13 01:12:52 2016
		Bits:       0x1f07ffff,               // 520617983 [0007ffff00000000000000000000000000000000000000000000000000000000]
     NonceHeaderA: [ciphrtxt.MessageHeaderLengthV2]byte{
         0x4d, 0x02, 0x00, 0x00, 0x58, 0x07, 0x94, 0xba, /* |M...X...| */
         0x58, 0x10, 0xcf, 0x3a, 0x02, 0x15, 0x7b, 0x74, /* |X..:..{t| */
         0xcc, 0xa9, 0x89, 0x07, 0x61, 0xc2, 0x68, 0x67, /* |....a.hg| */
         0xaa, 0x62, 0x59, 0x86, 0x19, 0xba, 0x4d, 0x82, /* |.bY...M.| */
         0xca, 0x41, 0x6e, 0x38, 0xd7, 0xf7, 0x2b, 0xd4, /* |.An8..+.| */
         0x20, 0x51, 0x5b, 0x82, 0x71, 0x03, 0x16, 0x95, /* | Q[.q...| */
         0x74, 0xcf, 0x0a, 0x92, 0x6a, 0x6f, 0xfd, 0xa0, /* |t...jo..| */
         0xd9, 0xee, 0x17, 0x0b, 0xee, 0x47, 0x3a, 0xc3, /* |.....G:.| */
         0x44, 0x61, 0x19, 0x93, 0xfe, 0xdb, 0x66, 0x75, /* |Da....fu| */
         0xba, 0x59, 0x3d, 0xce, 0xf6, 0x0a, 0x03, 0x5b, /* |.Y=....[| */
         0x69, 0xeb, 0x15, 0x97, 0x7d, 0x6e, 0xda, 0xc4, /* |i...}n..| */
         0xab, 0x1b, 0x55, 0x7e, 0x63, 0x9d, 0xce, 0x02, /* |..U~c...| */
         0x79, 0x99, 0x89, 0x0c, 0xfa, 0x90, 0x3f, 0x39, /* |y.....?9| */
         0x7c, 0x73, 0x34, 0x7d, 0x79, 0xe5, 0x0b, 0xaf, /* ||s4}y...| */
         0x18, 0x16, 0xdf, 0x15, 0x38, 0x33, 0xd2, 0xbe, /* |....83..| */
         0x7f, 0x09, 0xea, 0x78, 0x79, 0x7d, 0x44, 0x7a, /* |..xy}Dz| */
         0x9a, 0x1d, 0x70, 0x39, 0x31, 0xe8, 0x3b, 0xf4, /* |..p91.;.| */
         0x37, 0x0c, 0x47, 0x25, 0xe6, 0x7c, 0x63, 0x89, /* |7.G%.|c.| */
         0xb1, 0x57, 0x98, 0x97, 0x0c, 0xee, 0x6a, 0xb2, /* |.W....j.| */
         0xf6, 0x34, 0x7f, 0x19, 0x3b, 0x20, 0xe7, 0x7f, /* |.4.; .| */
         0x80, 0x4d, 0x60, 0x7b, 0x1f, 0x8b, 0x6d, 0x21, /* |.M`{..m!| */
         0xd4, 0xfc, 0xa2, 0x95, 0xa6, 0x24, 0xbe, 0x00, /* |.....$..| */
         0x00, 0x00, 0x8c, 0x42, /* |...B| */
     },
     NonceHeaderB: [ciphrtxt.MessageHeaderLengthV2]byte{
         0x4d, 0x02, 0x00, 0x00, 0x58, 0x05, 0x67, 0x24, /* |M...X.g$| */
         0x58, 0x0e, 0xa1, 0xa4, 0x03, 0xfc, 0xec, 0x64, /* |X......d| */
         0x93, 0x0e, 0x1b, 0x07, 0x6c, 0xaf, 0x76, 0x80, /* |....l.v.| */
         0xf0, 0xff, 0xaf, 0xc4, 0xb6, 0xf2, 0x49, 0xdd, /* |......I.| */
         0xc8, 0x76, 0xf9, 0x05, 0xc5, 0x54, 0xac, 0x78, /* |.v...T.x| */
         0xb7, 0xc2, 0xc7, 0x64, 0x76, 0x03, 0x9c, 0x31, /* |...dv..1| */
         0xbf, 0x8e, 0xef, 0x1d, 0x7e, 0x5f, 0x1c, 0xb7, /* |....~_..| */
         0x47, 0x6d, 0x90, 0x10, 0x8e, 0xb6, 0x09, 0x70, /* |Gm.....p| */
         0x1f, 0xe1, 0xc2, 0xfb, 0x9d, 0x9d, 0x6a, 0xa9, /* |......j.| */
         0xb8, 0xf8, 0x19, 0x9f, 0x9d, 0xb8, 0x02, 0xc7, /* |........| */
         0x42, 0xd1, 0x5f, 0x83, 0x6c, 0x67, 0x57, 0xc3, /* |B._.lgW.| */
         0x1c, 0x01, 0xf7, 0xbe, 0x54, 0x14, 0x6b, 0x3b, /* |....T.k;| */
         0x1c, 0xb4, 0x12, 0x87, 0xcf, 0xc1, 0xd4, 0x7c, /* |.......|| */
         0xe9, 0x35, 0x3e, 0xbf, 0xd5, 0xc4, 0x9b, 0x21, /* |.5>....!| */
         0x4f, 0x01, 0xfe, 0xce, 0xcc, 0xf0, 0xe1, 0x5e, /* |O......^| */
         0xae, 0x4c, 0x81, 0x86, 0x39, 0x6f, 0x84, 0x9c, /* |.L..9o..| */
         0xc5, 0xc3, 0x74, 0xca, 0xfb, 0xdd, 0x7f, 0xd3, /* |..t....| */
         0x8b, 0x50, 0x99, 0x5a, 0xc2, 0x7b, 0xe9, 0x05, /* |.P.Z.{..| */
         0x6f, 0x01, 0xe8, 0xaa, 0x4c, 0x83, 0x37, 0x87, /* |o...L.7.| */
         0x7e, 0xc0, 0x5f, 0x80, 0xb8, 0xab, 0x85, 0xe2, /* |~._.....| */
         0x87, 0x6f, 0xaa, 0xdc, 0x1f, 0x34, 0x4d, 0xf6, /* |.o...4M.| */
         0x5f, 0x08, 0x61, 0x8e, 0x00, 0x39, 0xb4, 0x00, /* |_.a..9..| */
         0x00, 0x01, 0x7c, 0x9b, /* |..|.| */
     },
	},
	Transactions: []*wire.MsgTx{&ctredGenesisCoinbaseTx},
}
