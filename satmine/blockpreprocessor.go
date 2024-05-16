package satmine

import (
	"fmt"
	"strconv"

	"github.com/dgraph-io/badger/v4"
)

// filterBlockData filters the given HookBlock, copying all data except Transfers into a newBlock.
// It then iterates over Transfers and includes only those with a corresponding entry
// in the key-value database. The key is searched using two prefixes: mrc721::inscr_addr::[ID]
// and mrc20::inscr_addr::[ID]. If a matching key is found with either prefix, the transfer is included.
func (b *BTOrdIdx) filterBlockData(block *HookBlock) (newBlock *HookBlock, err error) {
	// Initialize newBlock as a pointer to a new HookBlock instance
	newBlock = &HookBlock{
		BlockHeight:  block.BlockHeight,
		BlockHash:    block.BlockHash,
		Timestamp:    block.Timestamp,
		Inscriptions: block.Inscriptions,
		// Transfers will be filtered and added later
	}

	// Begin a read-only transaction with the database
	err = b.db.View(func(txn *badger.Txn) error {
		// Iterate over each transfer in the block
		fmt.Println("filterBlockData block.Transfers=", len(block.Transfers))
		for _, transfer := range block.Transfers {
			// Construct the prefixes for the key with the given ID
			prefixMrc721 := []byte(fmt.Sprintf("mrc721::inscr_addr::%s", transfer.ID))
			prefixMrc20 := []byte(fmt.Sprintf("mrc20::inscr_addr::%s", transfer.ID))
			//fmt.Println("filterBlockData prefixMrc721=", string(prefixMrc721))
			//fmt.Println("filterBlockData prefixMrc20=", string(prefixMrc20))

			// Check both prefixes in the database
			found := checkPrefixInDB(txn, prefixMrc721) || checkPrefixInDB(txn, prefixMrc20)

			if found {
				// Key exists with one of the prefixes, include this transfer in newBlock
				newBlock.Transfers = append(newBlock.Transfers, transfer)
				//fmt.Println("filterBlockData used transfer.ID ", transfer.ID)
			}
		}
		return nil
	})

	// Return the pointer to the filtered block and any error encountered
	return newBlock, err
}

// checkPrefixInDB checks if there is any key in the database that matches the given prefix.
// Returns true if a matching key is found, false otherwise.
func checkPrefixInDB(txn *badger.Txn, prefix []byte) bool {
	opts := badger.DefaultIteratorOptions
	opts.Prefix = prefix
	it := txn.NewIterator(opts)
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		//item := it.Item()
		//k := item.Key()
		//fmt.Println("Key found with prefix", string(k))
		return true // Key exists with the prefix, return true
	}
	return false // No key found with the prefix, return false
}

// checkBlockContinuity verifies the continuity of block heights in the blockchain.
// It checks if the current block height is exactly one more than the last block height.
func (b *BTOrdIdx) checkBlockContinuity(txn *badger.Txn, block *HookBlock) (err error) {
	// Define the key for the latest block
	item, err := txn.Get([]byte("latestblock"))
	if err != nil && err != badger.ErrKeyNotFound {
		return err // Returning an error here will cause the transaction to be discarded
	}
	if item != nil {
		var lastBlockHeightStr string
		err := item.Value(func(val []byte) error {
			lastBlockHeightStr = string(val)
			return nil
		})
		if err != nil {
			return err // Returning an error here will cause the transaction to be discarded
		}

		lastBlockHeight, err := strconv.Atoi(lastBlockHeightStr)
		if err != nil {
			return err // Returning an error here will cause the transaction to be discarded
		}

		currentBlockHeight, err := strconv.Atoi(block.BlockHeight)
		if err != nil {
			return err // Returning an error here will cause the transaction to be discarded
		}

		//fmt.Println("currentBlockHeight <= lastBlockHeight :", currentBlockHeight <= lastBlockHeight)
		if currentBlockHeight <= lastBlockHeight {
			return fmt.Errorf("block height mismatch: expected %d, got %d", lastBlockHeight, currentBlockHeight)
		}
	} else {
		// Uncomment and modify the below code if you need to check the height of the first block
		// currentBlockHeight, err := strconv.Atoi(block.BlockHeight)
		// if err != nil {
		// 	return err // Returning an error here will cause the transaction to be discarded
		// }
		// if currentBlockHeight != 0 {
		// 	return fmt.Errorf("expected first block height to be 0, got %d", currentBlockHeight)
		// }
	}
	return nil
}

// fillMissingBlocks fills the gaps between the last block and the current block with empty blocks.
func (b *BTOrdIdx) fillMissingBlocks(txn *badger.Txn, block *HookBlock) (error, []*HookBlock) {
	var lastBlockNumberStr string

	// Retrieve the latest block number
	item, err := txn.Get([]byte("latestblock"))
	if err != nil && err != badger.ErrKeyNotFound {
		return err, nil // Returning an error here will abort the transaction
	}

	var blocks []*HookBlock

	if item != nil {
		// Get the block number as a string
		err := item.Value(func(val []byte) error {
			lastBlockNumberStr = string(val)
			return nil
		})
		if err != nil {
			return err, nil // Returning an error here will abort the transaction
		}

		// Convert block number strings to integers
		lastBlockNumber, err := strconv.Atoi(lastBlockNumberStr)
		if err != nil {
			return err, nil // Returning an error here will abort the transaction
		}
		currentBlockNumber, err := strconv.Atoi(block.BlockHeight)
		if err != nil {
			return err, nil // Returning an error here will abort the transaction
		}

		// Check for missing blocks and fill the gap
		for h := lastBlockNumber + 1; h < currentBlockNumber; h++ {
			emptyBlock := &HookBlock{
				BlockHeight: strconv.Itoa(h),
				BlockHash:   NO_INSCRIPTION_BLOCK_HASH, // Use a default empty block hash
				// Initialize other fields as needed
			}
			blocks = append(blocks, emptyBlock)
		}
	}

	// Add the current block
	blocks = append(blocks, block)

	return nil, blocks
}
