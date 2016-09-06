// Copyright (c) 2016 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package rpctest

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

<<<<<<< HEAD
	"github.com/jadeblaquiere/ctcd/chaincfg"
	"github.com/jadeblaquiere/ctcd/chaincfg/chainhash"
	"github.com/jadeblaquiere/ctcd/wire"
	"github.com/jadeblaquiere/ctcrpcclient"
	"github.com/jadeblaquiere/ctcutil"
=======
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcrpcclient"
	"github.com/btcsuite/btcutil"
>>>>>>> btcsuite/master
)

var (
	// current number of active test nodes.
	numTestInstances = 0

	// defaultP2pPort is the initial p2p port which will be used by the
	// first created rpc harnesses to listen on for incoming p2p
	// connections.  Subsequent allocated ports for future rpc harness
	// instances will be monotonically increasing odd numbers calculated as
	// such: defaultP2pPort + (2 * harness.nodeNum).
	defaultP2pPort = 18555

	// defaultRPCPort is the initial rpc port which will be used by the
	// first created rpc harnesses to listen on for incoming rpc
	// connections. Subsequent allocated ports for future rpc harness
	// instances will be monotonically increasing even numbers calculated
	// as such: defaultP2pPort + (2 * harness.nodeNum).
	defaultRPCPort = 18556

	// testInstances is a private package-level slice used to keep track of
	// all active test harnesses. This global can be used to perform
	// various "joins", shutdown several active harnesses after a test,
	// etc.
	testInstances = make(map[string]*Harness)

	// Used to protest concurrent access to above declared variables.
	harnessStateMtx sync.RWMutex
)

// HarnessTestCase represents a test-case which utilizes an instance of the
// Harness to exercise functionality.
type HarnessTestCase func(r *Harness, t *testing.T)

// Harness fully encapsulates an active btcd process to provide a unified
// platform for creating rpc driven integration tests involving btcd. The
// active btcd node will typically be run in simnet mode in order to allow for
// easy generation of test blockchains.  The active btcd process is fully
// managed by Harness, which handles the necessary initialization, and teardown
// of the process along with any temporary directories created as a result.
// Multiple Harness instances may be run concurrently, in order to allow for
// testing complex scenarios involving multiple nodes. The harness also
// includes an in-memory wallet to streamline various classes of tests.
type Harness struct {
	// ActiveNet is the parameters of the blockchain the Harness belongs
	// to.
	ActiveNet *chaincfg.Params

<<<<<<< HEAD
	Node     *ctcrpcclient.Client
	node     *node
	handlers *ctcrpcclient.NotificationHandlers
=======
	Node     *btcrpcclient.Client
	node     *node
	handlers *btcrpcclient.NotificationHandlers
>>>>>>> btcsuite/master

	wallet *memWallet

	testNodeDir    string
	maxConnRetries int
	nodeNum        int

	sync.Mutex
}

// New creates and initializes new instance of the rpc test harness.
// Optionally, websocket handlers and a specified configuration may be passed.
// In the case that a nil config is passed, a default configuration will be
// used.
//
// NOTE: This function is safe for concurrent access.
<<<<<<< HEAD
func New(activeNet *chaincfg.Params, handlers *ctcrpcclient.NotificationHandlers,
=======
func New(activeNet *chaincfg.Params, handlers *btcrpcclient.NotificationHandlers,
>>>>>>> btcsuite/master
	extraArgs []string) (*Harness, error) {

	harnessStateMtx.Lock()
	defer harnessStateMtx.Unlock()

	harnessID := strconv.Itoa(int(numTestInstances))
	nodeTestData, err := ioutil.TempDir("", "rpctest-"+harnessID)
	if err != nil {
		return nil, err
	}

	certFile := filepath.Join(nodeTestData, "rpc.cert")
	keyFile := filepath.Join(nodeTestData, "rpc.key")
	if err := genCertPair(certFile, keyFile); err != nil {
		return nil, err
	}

	wallet, err := newMemWallet(activeNet, uint32(numTestInstances))
	if err != nil {
		return nil, err
	}

	miningAddr := fmt.Sprintf("--miningaddr=%s", wallet.coinbaseAddr)
	extraArgs = append(extraArgs, miningAddr)

	config, err := newConfig("rpctest", certFile, keyFile, extraArgs)
	if err != nil {
		return nil, err
	}

	// Generate p2p+rpc listening addresses.
	config.listen, config.rpcListen = generateListeningAddresses()

	// Create the testing node bounded to the simnet.
	node, err := newNode(config, nodeTestData)
	if err != nil {
		return nil, err
	}

	nodeNum := numTestInstances
	numTestInstances++

	if handlers == nil {
<<<<<<< HEAD
		handlers = &ctcrpcclient.NotificationHandlers{}
=======
		handlers = &btcrpcclient.NotificationHandlers{}
>>>>>>> btcsuite/master
	}

	// If a handler for the OnBlockConnected/OnBlockDisconnected callback
	// has already been set, then we create a wrapper callback which
	// executes both the currently registered callback, and the mem
	// wallet's callback.
	if handlers.OnBlockConnected != nil {
		obc := handlers.OnBlockConnected
		handlers.OnBlockConnected = func(hash *chainhash.Hash, height int32, t time.Time) {
			wallet.IngestBlock(hash, height, t)
			obc(hash, height, t)
		}
	} else {
		// Otherwise, we can claim the callback ourselves.
		handlers.OnBlockConnected = wallet.IngestBlock
	}
	if handlers.OnBlockDisconnected != nil {
		obd := handlers.OnBlockConnected
		handlers.OnBlockDisconnected = func(hash *chainhash.Hash, height int32, t time.Time) {
			wallet.UnwindBlock(hash, height, t)
			obd(hash, height, t)
		}
	} else {
		handlers.OnBlockDisconnected = wallet.UnwindBlock
	}

	h := &Harness{
		handlers:       handlers,
		node:           node,
		maxConnRetries: 20,
		testNodeDir:    nodeTestData,
		ActiveNet:      activeNet,
		nodeNum:        nodeNum,
		wallet:         wallet,
	}

	// Track this newly created test instance within the package level
	// global map of all active test instances.
	testInstances[h.testNodeDir] = h

	return h, nil
}

// SetUp initializes the rpc test state. Initialization includes: starting up a
// simnet node, creating a websockets client and connecting to the started
// node, and finally: optionally generating and submitting a testchain with a
// configurable number of mature coinbase outputs coinbase outputs.
//
// NOTE: This method and TearDown should always be called from the same
// goroutine as they are not concurrent safe.
func (h *Harness) SetUp(createTestChain bool, numMatureOutputs uint32) error {
	// Start the btcd node itself. This spawns a new process which will be
	// managed
	if err := h.node.start(); err != nil {
		return err
	}
	if err := h.connectRPCClient(); err != nil {
		return err
	}

	h.wallet.Start()

	// Ensure the btcd properly dispatches our registered call-back for
	// each new block. Otherwise, the memWallet won't function properly.
	if err := h.Node.NotifyBlocks(); err != nil {
		return err
	}

	// Create a test chain with the desired number of mature coinbase
	// outputs.
	if createTestChain && numMatureOutputs != 0 {
		numToGenerate := (uint32(h.ActiveNet.CoinbaseMaturity) +
			numMatureOutputs)
		_, err := h.Node.Generate(numToGenerate)
		if err != nil {
			return err
		}
	}

	// Block until the wallet has fully synced up to the tip of the main
	// chain.
	_, height, err := h.Node.GetBestBlock()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(time.Millisecond * 100)
out:
	for {
		select {
		case <-ticker.C:
			walletHeight := h.wallet.SyncedHeight()
			if walletHeight == height {
				break out
			}
		}
	}

	return nil
}

// TearDown stops the running rpc test instance. All created processes are
// killed, and temporary directories removed.
//
// NOTE: This method and SetUp should always be called from the same goroutine
// as they are not concurrent safe.
func (h *Harness) TearDown() error {
	if h.Node != nil {
		h.Node.Shutdown()
	}

	if err := h.node.shutdown(); err != nil {
		return err
	}

	if err := os.RemoveAll(h.testNodeDir); err != nil {
		return err
	}

	delete(testInstances, h.testNodeDir)

	return nil
}

// connectRPCClient attempts to establish an RPC connection to the created btcd
// process belonging to this Harness instance. If the initial connection
// attempt fails, this function will retry h.maxConnRetries times, backing off
// the time between subsequent attempts. If after h.maxConnRetries attempts,
// we're not able to establish a connection, this function returns with an
// error.
func (h *Harness) connectRPCClient() error {
<<<<<<< HEAD
	var client *ctcrpcclient.Client
=======
	var client *btcrpcclient.Client
>>>>>>> btcsuite/master
	var err error

	rpcConf := h.node.config.rpcConnConfig()
	for i := 0; i < h.maxConnRetries; i++ {
<<<<<<< HEAD
		if client, err = ctcrpcclient.New(&rpcConf, h.handlers); err != nil {
=======
		if client, err = btcrpcclient.New(&rpcConf, h.handlers); err != nil {
>>>>>>> btcsuite/master
			time.Sleep(time.Duration(i) * 50 * time.Millisecond)
			continue
		}
		break
	}

	if client == nil {
		return fmt.Errorf("connection timeout")
	}

	h.Node = client
	h.wallet.SetRPCClient(client)
	return nil
}

// NewAddress returns a fresh address spendable by the Harness' internal
// wallet.
//
// This function is safe for concurrent access.
func (h *Harness) NewAddress() (btcutil.Address, error) {
	return h.wallet.NewAddress()
}

// ConfirmedBalance returns the confirmed balance of the Harness' internal
// wallet.
//
// This function is safe for concurrent access.
func (h *Harness) ConfirmedBalance() btcutil.Amount {
	return h.wallet.ConfirmedBalance()
}

// SendOutputs creates, signs, and finally broadcasts a transaction spending
// the harness' available mature coinbase outputs creating new outputs
// according to targetOutputs.
//
// This function is safe for concurrent access.
func (h *Harness) SendOutputs(targetOutputs []*wire.TxOut,
	feeRate btcutil.Amount) (*chainhash.Hash, error) {

	return h.wallet.SendOutputs(targetOutputs, feeRate)
}

// CreateTransaction returns a fully signed transaction paying to the specified
// outputs while observing the desired fee rate. The passed fee rate should be
// expressed in satoshis-per-byte. Any unspent outputs selected as inputs for
// the crafted transaction are marked as unspendable in order to avoid
// potential double-spends by future calls to this method. If the created
// transaction is cancelled for any reason then the selected inputs MUST be
// freed via a call to UnlockOutputs. Otherwise, the locked inputs won't be
// returned to the pool of spendable outputs.
//
// This function is safe for concurrent access.
func (h *Harness) CreateTransaction(targetOutputs []*wire.TxOut,
	feeRate btcutil.Amount) (*wire.MsgTx, error) {

	return h.wallet.CreateTransaction(targetOutputs, feeRate)
}

// UnlockOutputs unlocks any outputs which were previously marked as
// unspendabe due to being selected to fund a transaction via the
// CreateTransaction method.
//
// This function is safe for concurrent access.
func (h *Harness) UnlockOutputs(inputs []*wire.TxIn) {
	h.wallet.UnlockOutputs(inputs)
}

// RPCConfig returns the harnesses current rpc configuration. This allows other
// potential RPC clients created within tests to connect to a given test
// harness instance.
<<<<<<< HEAD
func (h *Harness) RPCConfig() ctcrpcclient.ConnConfig {
=======
func (h *Harness) RPCConfig() btcrpcclient.ConnConfig {
>>>>>>> btcsuite/master
	return h.node.config.rpcConnConfig()
}

// GenerateAndSubmitBlock creates a block whose contents include the passed
// transactions and submits it to the running simnet node. For generating
// blocks with only a coinbase tx, callers can simply pass nil instead of
// transactions to be mined. Additionally, a custom block version can be set by
// the caller. A blockVersion of -1 indicates that the current default block
// version should be used. An uninitialized time.Time should be used for the
// blockTime parameter if one doesn't wish to set a custom time.
//
// This function is safe for concurrent access.
func (h *Harness) GenerateAndSubmitBlock(txns []*btcutil.Tx, blockVersion int32,
	blockTime time.Time) (*btcutil.Block, error) {

	h.Lock()
	defer h.Unlock()

	if blockVersion == -1 {
		blockVersion = wire.BlockVersion
	}

	prevBlockHash, prevBlockHeight, err := h.Node.GetBestBlock()
	if err != nil {
		return nil, err
	}
	prevBlock, err := h.Node.GetBlock(prevBlockHash)
	if err != nil {
		return nil, err
	}
	prevBlock.SetHeight(prevBlockHeight)

	// Create a new block including the specified transactions
	newBlock, err := createBlock(prevBlock, txns, blockVersion,
		blockTime, h.wallet.coinbaseAddr, h.ActiveNet)
	if err != nil {
		return nil, err
	}

	// Submit the block to the simnet node.
	if err := h.Node.SubmitBlock(newBlock, nil); err != nil {
		return nil, err
	}

	return newBlock, nil
}

// generateListeningAddresses returns two strings representing listening
// addresses designated for the current rpc test. If there haven't been any
// test instances created, the default ports are used. Otherwise, in order to
// support multiple test nodes running at once, the p2p and rpc port are
// incremented after each initialization.
func generateListeningAddresses() (string, string) {
	var p2p, rpc string
	localhost := "127.0.0.1"

	if numTestInstances == 0 {
		p2p = net.JoinHostPort(localhost, strconv.Itoa(defaultP2pPort))
		rpc = net.JoinHostPort(localhost, strconv.Itoa(defaultRPCPort))
	} else {
		p2p = net.JoinHostPort(localhost,
			strconv.Itoa(defaultP2pPort+(2*numTestInstances)))
		rpc = net.JoinHostPort(localhost,
			strconv.Itoa(defaultRPCPort+(2*numTestInstances)))
	}

	return p2p, rpc
}
