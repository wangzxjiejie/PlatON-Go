package vm_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PlatONnetwork/PlatON-Go/common"
	"github.com/PlatONnetwork/PlatON-Go/common/hexutil"
	"github.com/PlatONnetwork/PlatON-Go/consensus/ethash"
	"github.com/PlatONnetwork/PlatON-Go/core"
	"github.com/PlatONnetwork/PlatON-Go/core/ppos"
	"github.com/PlatONnetwork/PlatON-Go/core/state"
	"github.com/PlatONnetwork/PlatON-Go/core/vm"
	"github.com/PlatONnetwork/PlatON-Go/ethdb"
	"github.com/PlatONnetwork/PlatON-Go/p2p/discover"
	"github.com/PlatONnetwork/PlatON-Go/params"
	"github.com/PlatONnetwork/PlatON-Go/rlp"
	"math/big"
	"time"

	"testing"
)

func TestCandidatePoolOverAll(t *testing.T) {

	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(7000)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")

	// CandidateDetails(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateDetails input==>", "nodeId: ", nodeId.String())
	resByte, err := candidateContract.CandidateDetails(nodeId)
	if nil != err {
		fmt.Println("CandidateDetails fail", "err", err)
		return
	}
	if nil == resByte {
		fmt.Println("The candidate info is null")
		return
	}
	fmt.Println("The candidate info is: ", vm.ResultByte2Json(resByte))

	// CandidateApplyWithdraw(nodeId discover.NodeID, withdraw *big.Int) ([]byte, error)
	withdraw := big.NewInt(100)
	fmt.Println("CandidateApplyWithdraw input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err = candidateContract.CandidateApplyWithdraw(nodeId, withdraw)
	if nil != err {
		fmt.Println("CandidateApplyWithdraw fail", "err", err)
	}
	fmt.Println("CandidateApplyWithdraw success")

	// CandidateWithdraw(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateWithdraw input==>", "nodeId: ", nodeId.String())
	_, err = candidateContract.CandidateWithdraw(nodeId)
	if nil != err {
		fmt.Println("CandidateWithdraw fail", "err", err)
	}
	fmt.Println("CandidateWithdraw success")

	// CandidateWithdrawInfos(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateWithdrawInfos input==>", "nodeId: ", nodeId.String())
	resByte, err = candidateContract.CandidateWithdrawInfos(nodeId)
	if nil != err {
		fmt.Println("CandidateWithdrawInfos fail", "err", err)
		return
	}
	if nil == resByte {
		fmt.Println("The CandidateWithdrawInfos is null")
		return
	}
	fmt.Println("The CandidateWithdrawInfos is: ", vm.ResultByte2Json(resByte))

	// CandidateList() ([]byte, error)
	resByte, err = candidateContract.CandidateList()
	if nil != err {
		fmt.Println("CandidateList fail", "err", err)
	}
	if nil == resByte {
		fmt.Println("The candidate list is null")
		return
	}
	fmt.Println("The candidate list is: ", vm.ResultByte2Json(resByte))
}

func newContract() *vm.Contract {
	callerAddress := vm.AccountRef(common.HexToAddress("0x12"))
	contract := vm.NewContract(callerAddress, callerAddress, big.NewInt(1000), uint64(1))
	return contract
}

func newEvm() *vm.EVM {
	state, _ := newChainState()
	candidatePool, ticketPool := newPool()
	evm := &vm.EVM{
		StateDB:       state,
		CandidatePool: candidatePool,
		TicketPool:    ticketPool,
	}
	context := vm.Context{
		BlockNumber: big.NewInt(7),
	}
	evm.Context = context
	return evm
}

func newChainState() (*state.StateDB, error) {
	var (
		db      = ethdb.NewMemDatabase()
		genesis = new(core.Genesis).MustCommit(db)
	)
	fmt.Println("genesis", genesis)
	// Initialize a fresh chain with only a genesis block
	blockchain, _ := core.NewBlockChain(db, nil, params.AllEthashProtocolChanges, ethash.NewFaker(), vm.Config{}, nil)

	var state *state.StateDB
	if statedb, err := blockchain.State(); nil != err {
		return nil, errors.New("reference statedb failed" + err.Error())
	} else {
		state = statedb
	}
	return state, nil
}

func newPool() (*pposm.CandidatePool, *pposm.TicketPool) {
	configs := params.PposConfig{
		Candidate: &params.CandidateConfig{
			MaxChair:          1,
			MaxCount:          3,
			RefundBlockNumber: 1,
		},
		TicketConfig: &params.TicketConfig{
			MaxCount:          100,
			ExpireBlockNumber: 2,
		},
	}
	return pposm.NewCandidatePool(&configs), pposm.NewTicketPool(&configs)
}

func TestCandidateDeposit(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}

	// CandidateDeposit(nodeId discover.NodeID, owner common.Address, fee uint64, host, port, extra string) ([]byte, error)
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(7000)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")
}

func TestCandidateDetails(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(7000)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")

	// CandidateDetails(nodeId discover.NodeID) ([]byte, error)
	// nodeId = discover.MustHexID("0x11234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	fmt.Println("CandidateDetails input==>", "nodeId: ", nodeId.String())
	resByte, err := candidateContract.CandidateDetails(nodeId)
	if nil != err {
		fmt.Println("CandidateDetails fail", "err", err)
		return
	}
	if nil == resByte {
		fmt.Println("The candidate info is null")
		return
	}
	fmt.Println("The candidate info is: ", vm.ResultByte2Json(resByte))
}

func TestGetBatchCandidateDetail(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(7000)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit1 success")

	nodeId = discover.MustHexID("0x11234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner = common.HexToAddress("0x12")
	fee = uint64(7000)
	host = "10.0.0.2"
	port = "8548"
	extra = "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err = candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit2 success")

	// GetBatchCandidateDetail(nodeIds []discover.NodeID) ([]byte, error)
	nodeIds := []discover.NodeID{discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345"), discover.MustHexID("0x11234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")}
	input, _ := json.Marshal(nodeIds)
	fmt.Println("GetBatchCandidateDetail input==>", "nodeIds: ", string(input))
	resByte, err := candidateContract.GetBatchCandidateDetail(nodeIds)
	if nil != err {
		fmt.Println("GetBatchCandidateDetail fail", "err", err)
	}
	if nil == resByte {
		fmt.Println("The candidate info is null")
		return
	}
	fmt.Println("The batch candidate info is: ", vm.ResultByte2Json(resByte))
}

func TestCandidateList(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(7000)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit1 success")

	nodeId = discover.MustHexID("0x11234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner = common.HexToAddress("0x12")
	fee = uint64(6800)
	host = "10.0.0.2"
	port = "8548"
	extra = "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err = candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit2 success")

	// CandidateList() ([]byte, error)
	resByte, err := candidateContract.CandidateList()
	if nil != err {
		fmt.Println("CandidateList fail", "err", err)
	}
	if nil == resByte {
		fmt.Println("The candidate list is null")
		return
	}
	fmt.Println("The candidate list is: ", vm.ResultByte2Json(resByte))
}

func TestSetCandidateExtra(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(1)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")

	// SetCandidateExtra(nodeId discover.NodeID, extra string) ([]byte, error)
	extra = "this node is powerful"
	fmt.Println("SetCandidateExtra input=>", "nodeId: ", nodeId.String(), "extra: ", extra)
	_, err = candidateContract.SetCandidateExtra(nodeId, extra)
	if nil != err {
		fmt.Println("SetCandidateExtra fail", "err", err)
	}
	fmt.Println("SetCandidateExtra success")

	// CandidateDetails(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateDetails input==>", "nodeId: ", nodeId.String())
	resByte, err := candidateContract.CandidateDetails(nodeId)
	if nil != err {
		fmt.Println("CandidateDetails fail", "err", err)
		return
	}
	if nil == resByte {
		fmt.Println("The candidate info is null")
		return
	}
	fmt.Println("The candidate info is: ", vm.ResultByte2Json(resByte))
}

func TestCandidateApplyWithdraw(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(1)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")

	// CandidateApplyWithdraw(nodeId discover.NodeID, withdraw *big.Int) ([]byte, error)
	withdraw := big.NewInt(100)
	fmt.Println("CandidateApplyWithdraw input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err = candidateContract.CandidateApplyWithdraw(nodeId, withdraw)
	if nil != err {
		fmt.Println("CandidateApplyWithdraw fail", "err", err)
	}
	fmt.Println("CandidateApplyWithdraw success")
}

func TestCandidateWithdraw(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(7000)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")

	// CandidateApplyWithdraw(nodeId discover.NodeID, withdraw *big.Int) ([]byte, error)
	withdraw := big.NewInt(100)
	fmt.Println("CandidateApplyWithdraw input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err = candidateContract.CandidateApplyWithdraw(nodeId, withdraw)
	if nil != err {
		fmt.Println("CandidateApplyWithdraw fail", "err", err)
	}
	fmt.Println("CandidateApplyWithdraw success")

	// CandidateWithdraw(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateWithdraw input==>", "nodeId: ", nodeId.String())
	_, err = candidateContract.CandidateWithdraw(nodeId)
	if nil != err {
		fmt.Println("CandidateWithdraw fail", "err", err)
	}
	fmt.Println("CandidateWithdraw success")

	// CandidateDetails(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateDetails input==>", "nodeId: ", nodeId.String())
	resByte, err := candidateContract.CandidateDetails(nodeId)
	if nil != err {
		fmt.Println("CandidateDetails fail", "err", err)
		return
	}
	if nil == resByte {
		fmt.Println("The candidate info is null")
		return
	}
	fmt.Println("The candidate info is: ", vm.ResultByte2Json(resByte))
}

func TestCandidateWithdrawInfos(t *testing.T) {
	candidateContract := vm.CandidateContract{
		newContract(),
		newEvm(),
	}
	nodeId := discover.MustHexID("0x01234567890121345678901123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345")
	owner := common.HexToAddress("0x12")
	fee := uint64(1)
	host := "10.0.0.1"
	port := "8548"
	extra := "extra data"
	fmt.Println("CandidateDeposit input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err := candidateContract.CandidateDeposit(nodeId, owner, fee, host, port, extra)
	if nil != err {
		fmt.Println("CandidateDeposit fail", "err", err)
	}
	fmt.Println("CandidateDeposit success")

	// CandidateApplyWithdraw(nodeId discover.NodeID, withdraw *big.Int) ([]byte, error)
	withdraw := big.NewInt(100)
	fmt.Println("CandidateApplyWithdraw input==>", "nodeId: ", nodeId.String(), "owner: ", owner.Hex(), "fee: ", fee, "host: ", host, "port: ", port, "extra: ", extra)
	_, err = candidateContract.CandidateApplyWithdraw(nodeId, withdraw)
	if nil != err {
		fmt.Println("CandidateApplyWithdraw fail", "err", err)
	}
	fmt.Println("CandidateApplyWithdraw success")

	// CandidateWithdraw(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateWithdraw input==>", "nodeId: ", nodeId.String())
	_, err = candidateContract.CandidateWithdraw(nodeId)
	if nil != err {
		fmt.Println("CandidateWithdraw fail", "err", err)
	}
	fmt.Println("CandidateWithdraw success")

	// CandidateWithdrawInfos(nodeId discover.NodeID) ([]byte, error)
	fmt.Println("CandidateWithdrawInfos input==>", "nodeId: ", nodeId.String())
	resByte, err := candidateContract.CandidateWithdrawInfos(nodeId)
	if nil != err {
		fmt.Println("CandidateWithdrawInfos fail", "err", err)
		return
	}
	if nil == resByte {
		fmt.Println("The CandidateWithdrawInfos is null")
		return
	}
	fmt.Println("The CandidateWithdrawInfos is: ", vm.ResultByte2Json(resByte))
}

func TestTime(t *testing.T) {
	fmt.Printf("时间戳（毫秒）：%v;\n", time.Now().UnixNano()/1e6)
}

func TestCandidatePoolEncode(t *testing.T) {

	nodeId := []byte("0xb6c8c9f99bfebfa4fb174df720b9385dbd398de699ec36750af3f38f8e310d4f0b90447acbef64bdf924c4b59280f3d42bb256e6123b53e9a7e99e4c432549d6")
	owner := []byte("0x740ce31b3fac20dac379db243021a51e80ad00d7")
	extra := "{\"nodeName\": \"Platon-Beijing\", \"nodePortrait\": \"\",\"nodeDiscription\": \"PlatON-引力区\",\"nodeDepartment\": \"JUZIX\",\"officialWebsite\": \"https://www.platon.network/\",\"time\":1546503651190}"
	// CandidateDeposit(nodeId discover.NodeID, owner common.Address, fee uint64, host, port, extra string)
	var CandidateDeposit [][]byte
	CandidateDeposit = make([][]byte, 0)
	CandidateDeposit = append(CandidateDeposit, uint64ToBytes(1001))
	CandidateDeposit = append(CandidateDeposit, []byte("CandidateDeposit"))
	CandidateDeposit = append(CandidateDeposit, nodeId)
	CandidateDeposit = append(CandidateDeposit, owner)
	CandidateDeposit = append(CandidateDeposit, uint64ToBytes(8000))
	CandidateDeposit = append(CandidateDeposit, []byte("192.168.9.183"))
	CandidateDeposit = append(CandidateDeposit, []byte("16789"))
	CandidateDeposit = append(CandidateDeposit, []byte(extra))
	bufDeposit := new(bytes.Buffer)
	err := rlp.Encode(bufDeposit, CandidateDeposit)
	if err != nil {
		fmt.Println(err)
		t.Errorf("CandidateDeposit encode rlp data fail")
	} else {
		fmt.Println("CandidateDeposit data rlp: ", hexutil.Encode(bufDeposit.Bytes()))
	}

	// CandidateApplyWithdraw(nodeId discover.NodeID, withdraw *big.Int)
	var CandidateApplyWithdraw [][]byte
	CandidateApplyWithdraw = make([][]byte, 0)
	CandidateApplyWithdraw = append(CandidateApplyWithdraw, uint64ToBytes(1002))
	CandidateApplyWithdraw = append(CandidateApplyWithdraw, []byte("CandidateApplyWithdraw"))
	CandidateApplyWithdraw = append(CandidateApplyWithdraw, nodeId)
	withdraw, ok := new(big.Int).SetString("14d1120d7b160000", 16)
	if !ok {
		t.Errorf("big int setstring fail")
	}
	CandidateApplyWithdraw = append(CandidateApplyWithdraw, withdraw.Bytes())
	bufApply := new(bytes.Buffer)
	err = rlp.Encode(bufApply, CandidateApplyWithdraw)
	if err != nil {
		fmt.Println(err)
		t.Errorf("CandidateApplyWithdraw encode rlp data fail")
	} else {
		fmt.Println("CandidateApplyWithdraw data rlp: ", hexutil.Encode(bufApply.Bytes()))
	}

	// CandidateWithdraw(nodeId discover.NodeID)
	var CandidateWithdraw [][]byte
	CandidateWithdraw = make([][]byte, 0)
	CandidateWithdraw = append(CandidateWithdraw, uint64ToBytes(1003))
	CandidateWithdraw = append(CandidateWithdraw, []byte("CandidateWithdraw1"))
	CandidateWithdraw = append(CandidateWithdraw, nodeId)
	bufWith := new(bytes.Buffer)
	err = rlp.Encode(bufWith, CandidateWithdraw)
	if err != nil {
		fmt.Println(err)
		t.Errorf("CandidateWithdraw encode rlp data fail")
	} else {
		fmt.Println("CandidateWithdraw data rlp: ", hexutil.Encode(bufWith.Bytes()))
	}

	// CandidateWithdrawInfos(nodeId discover.NodeID)
	var CandidateWithdrawInfos [][]byte
	CandidateWithdrawInfos = make([][]byte, 0)
	CandidateWithdrawInfos = append(CandidateWithdrawInfos, uint64ToBytes(0xf1))
	CandidateWithdrawInfos = append(CandidateWithdrawInfos, []byte("CandidateWithdrawInfos"))
	CandidateWithdrawInfos = append(CandidateWithdrawInfos, nodeId)
	bufWithInfos := new(bytes.Buffer)
	err = rlp.Encode(bufWithInfos, CandidateWithdrawInfos)
	if err != nil {
		fmt.Println(err)
		t.Errorf("CandidateWithdrawInfos encode rlp data fail")
	} else {
		fmt.Println("CandidateWithdrawInfos data rlp: ", hexutil.Encode(bufWithInfos.Bytes()))
	}

	// CandidateDetails(nodeId discover.NodeID)
	var CandidateDetails [][]byte
	CandidateDetails = make([][]byte, 0)
	CandidateDetails = append(CandidateDetails, uint64ToBytes(0xf1))
	CandidateDetails = append(CandidateDetails, []byte("CandidateDetails"))
	CandidateDetails = append(CandidateDetails, nodeId)
	bufDetails := new(bytes.Buffer)
	err = rlp.Encode(bufDetails, CandidateDetails)
	if err != nil {
		fmt.Println(err)
		t.Errorf("CandidateDetails encode rlp data fail")
	} else {
		fmt.Println("CandidateDetails data rlp: ", hexutil.Encode(bufDetails.Bytes()))
	}

	// CandidateList()
	var CandidateList [][]byte
	CandidateList = make([][]byte, 0)
	CandidateList = append(CandidateList, uint64ToBytes(0xf1))
	CandidateList = append(CandidateList, []byte("CandidateList"))
	bufCList := new(bytes.Buffer)
	err = rlp.Encode(bufCList, CandidateList)
	if err != nil {
		fmt.Println(err)
		t.Errorf("CandidateList encode rlp data fail")
	} else {
		fmt.Println("CandidateList data rlp: ", hexutil.Encode(bufCList.Bytes()))
	}
}

func TestCandidatePoolDecode(t *testing.T) {

	//HexString -> []byte
	rlpcode, _ := hex.DecodeString("f9024005853339059800829c409410000000000000000000000000000000000000019331303030303030303030303030303030303030b901c7f901c48800000000000000029043616e6469646174654465706f736974b88230786133363364313234333634366236656162663164343835316636343662353233663537303764303533636161623935303232663136383236303561636130353337656530633563313462346466613736646362636532363462376536386435396465373961343262376364613035396539643335383333366139616238643830aa3078663231366436653463313730393761363065653262386535633838393431636439663037323633628800000000000001f487302e302e302e30853330333033b8e27b226e6f64654e616d65223a22e88a82e782b9e5908de7a7b0222c226e6f64654469736372697074696f6e223a22e88a82e782b9e7ae80e4bb8b222c226e6f64654465706172746d656e74223a22e69cbae69e84e5908de7a7b0222c226f6666696369616c57656273697465223a227777772e706c61746f6e2e6e6574776f726b222c226e6f6465506f727472616974223a2255524c222c2274696d65223a313534333931333639353638352c226f776e6572223a22307866323136643665346331373039376136306565326238653563383839343163643966303732363362227d1ba0e789e2d95ed796dec19e7a40b760a9849a1ca09110e1f95e46ed0c18487cfdf3a021bfd18bdb4c32a70f2836c6b78944c88b92d09f5e5201f125444792538617e7")
	var source [][]byte
	if err := rlp.Decode(bytes.NewReader(rlpcode), &source); err != nil {
		fmt.Println(err)
		t.Errorf("TestRlpDecode decode rlp data fail")
	}

	for i, v := range source {
		fmt.Println("i: ", i, " v: ", hex.EncodeToString(v))
	}
}

func TestAppendSlice(t *testing.T) {
	a := []int{0, 1, 2, 3, 4}
	i := 2
	a = append(a[:i], a[i+1:]...)
	fmt.Println(a)
}

func uint64ToBytes(val uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, val)
	return buf[:]
}
