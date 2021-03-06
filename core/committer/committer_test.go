/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

		 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package committer

import (
	"testing"

	"github.com/hyperledger/fabric/common/configtx/test"
	"github.com/hyperledger/fabric/common/ledger/testutil"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/hyperledger/fabric/core/ledger/ledgermgmt"
	"github.com/hyperledger/fabric/core/mocks/validator"
	"github.com/hyperledger/fabric/protos/common"
)

func TestKVLedgerBlockStorage(t *testing.T) {
	viper.Set("peer.fileSystemPath", "/tmp/fabric/committertest")
	ledgermgmt.InitializeTestEnv()
	defer ledgermgmt.CleanupTestEnv()
	gb, _ := test.MakeGenesisBlock("TestLedger")
	gbHash := gb.Header.Hash()
	ledger, err := ledgermgmt.CreateLedger(gb)
	assert.NoError(t, err, "Error while creating ledger: %s", err)
	defer ledger.Close()

	committer := NewLedgerCommitter(ledger, &validator.MockValidator{})
	height, err := committer.LedgerHeight()
	assert.Equal(t, uint64(1), height)
	assert.NoError(t, err)

	bcInfo, _ := ledger.GetBlockchainInfo()
	testutil.AssertEquals(t, bcInfo, &common.BlockchainInfo{
		Height: 1, CurrentBlockHash: gbHash, PreviousBlockHash: nil})

	simulator, _ := ledger.NewTxSimulator()
	simulator.SetState("ns1", "key1", []byte("value1"))
	simulator.SetState("ns1", "key2", []byte("value2"))
	simulator.SetState("ns1", "key3", []byte("value3"))
	simulator.Done()

	simRes, _ := simulator.GetTxSimulationResults()
	block1 := testutil.ConstructBlock(t, 1, gbHash, [][]byte{simRes}, true)

	err = committer.Commit(block1)
	assert.NoError(t, err)

	height, err = committer.LedgerHeight()
	assert.Equal(t, uint64(2), height)
	assert.NoError(t, err)

	blocks := committer.GetBlocks([]uint64{0})
	assert.Equal(t, 1, len(blocks))
	assert.NoError(t, err)

	bcInfo, _ = ledger.GetBlockchainInfo()
	block1Hash := block1.Header.Hash()
	testutil.AssertEquals(t, bcInfo, &common.BlockchainInfo{
		Height: 2, CurrentBlockHash: block1Hash, PreviousBlockHash: gbHash})
}
