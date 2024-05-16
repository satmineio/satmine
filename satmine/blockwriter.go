// filePath: satmine/blockwriter.go

package satmine

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v4"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// BTOrdIdx is a specialized class designed for indexing Bitcoin Ordinals. This class facilitates the efficient
// retrieval and management of Ordinal information within the Bitcoin blockchain. It efficiently organizes and provides
// quick access to Ordinal data, aiding in operations like tracking, searching, and analysis of specific Ordinals.
type BTOrdIdx struct {
	db     *badger.DB
	rwLock sync.RWMutex
}

// NewBTOrdIdx initializes a new instance of BTOrdIdx with a given Manager.
func NewBTOrdIdx(db *badger.DB) *BTOrdIdx {
	return &BTOrdIdx{
		db: db,
	}
}

// // WriteBlock writes a new block of data to the database.
// func (b *BTOrdIdx) StartWriteBlock(block *HookBlock) (err error) {

// }

// WriteBlock writes a new block of data to the database.
func (b *BTOrdIdx) WriteBlock(newBlock *HookBlock) (err error) {
	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// Handle the panic, convert it to error if needed
	// 		err = fmt.Errorf("write newBlock failed: %v", r)
	// 		logger.Error("WriteBlock failed: ", zap.Error(err))

	// 	}
	// }()
	b.rwLock.Lock()         // Acquire the write lock
	defer b.rwLock.Unlock() // Release the lock when the function returns

	// bjson := block.ToJson()
	// logger.Info("bjson: %+v\n", zap.Reflect("bjson", bjson))

	//

	//logger.Info(fmt.Sprintf("--Block:%s  %s--", newBlock.BlockHeight, newBlock.BlockHash))

	//logger.Info("block: %+v\n", zap.Reflect("newBlock", newBlock))

	//copying all data except Transfers into a newBlock.

	//fmt.Println("newBlock =", newBlock)

	filterBlock, err := b.filterBlockData(newBlock)
	if err != nil {
		logger.Error("filterBlockData failed: ", zap.Error(err))
		return err
	}

	//filterBlock := newBlock

	logger.Info(fmt.Sprintf("--Block:%s %d len:%d=%d--", filterBlock.BlockHeight, len(newBlock.Inscriptions), len(newBlock.Transfers), len(filterBlock.Transfers)))

	// Start a new transaction
	err = b.db.Update(func(txn *badger.Txn) error {

		// // //Verify Block Continuity
		if err := b.checkBlockContinuity(txn, filterBlock); err != nil {
			//return err

			logger.Info("Block height duplication: ", zap.String("BlockHeight", filterBlock.BlockHeight))
			return nil
		}

		err, blocks := b.fillMissingBlocks(txn, filterBlock)
		if err != nil {
			return err
		}
		for _, block := range blocks {
			logger.Info(fmt.Sprintf("make block %s", block.BlockHeight))
			// Write the list of newly inscribed inscriptions into the KV database.
			if err := b.addInscriptionList(txn, block); err != nil {
				return err
			}

			// Add the inscription transfer list to the key-value store data.
			if err := b.addTransferList(txn, block); err != nil {
				return err
			}

			//Mining using MRC721 inscriptions.
			if err := b.mineWithMrc721Inscription(txn, block); err != nil {
				return err
			}

			// lotteryWithBlockHash conducts a lottery draw based on a given block hash.
			if err := b.lotteryWithBlockHash(txn, block); err != nil {
				return err
			}

			// Serialize the block to JSON using jsoniter
			blockJSON, err := jsoniter.Marshal(block)
			if err != nil {
				logger.Error("Failed to marshal block: ", zap.Error(err))
				return err
			}
			// Write the serialized block to the database
			if err := txn.Set([]byte("latestblock"), []byte(block.BlockHeight)); err != nil {
				return err
			}
			if err := txn.Set([]byte("block::"+block.BlockHeight), blockJSON); err != nil {
				return err
			}
			if err := txn.Set([]byte("bkhash::"+block.BlockHash), []byte(block.BlockHeight)); err != nil {
				return err
			}

		}
		// Additional transaction operations can be added here

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("WriteBlock: ", zap.Error(err))
		return err
	}

	//logger.Info("Block successful num:", zap.String("BlockHeight", block.BlockHeight))

	return
}
