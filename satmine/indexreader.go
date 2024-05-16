// filePath: satmine/ider.go

package satmine

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// GetLastBlock retrieves the most recent block number from the database.
func (b *BTOrdIdx) GetLastBlock() (*big.Int, error) {
	var blockNumberStr string

	b.rwLock.RLock()         // Acquire the read lock
	defer b.rwLock.RUnlock() // Release the lock when the function returns

	// Retrieve the latest block number from the database
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("latestblock"))
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		// Deserialize the block number from the retrieved item
		err = item.Value(func(val []byte) error {
			blockNumberStr = string(val)
			return nil
		})
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("Failed to get the latest block number: ", zap.Error(err))
		return nil, err
	}

	// Convert the block number string to *big.Int
	blockNumber, ok := new(big.Int).SetString(blockNumberStr, 10)
	if !ok {
		logger.Error("Failed to convert block number to *big.Int", zap.String("blockNumberStr", blockNumberStr))
		return nil, errors.New("invalid block number")
	}

	logger.Info("Latest block number retrieved successfully")
	return blockNumber, nil
}

// GetBlockByHeight retrieves a block from the database by its height.
// It also checks if the blockHeight is a valid integer.
func (b *BTOrdIdx) GetBlockByHeight(blockHeight string) (*HookBlock, error) {
	var block HookBlock

	// Validate that blockHeight is an integer
	if _, err := strconv.Atoi(blockHeight); err != nil {
		logger.Error("Invalid block height: ", zap.String("blockHeight", blockHeight), zap.Error(err))
		return nil, fmt.Errorf("invalid block height: %s", blockHeight)
	}

	b.rwLock.RLock()         // Acquire the read lock
	defer b.rwLock.RUnlock() // Release the lock when the function returns

	// Retrieve the block from the database using the block height
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("block::" + blockHeight))
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		// Deserialize the block from the retrieved item
		err = item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &block)
		})
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("Failed to get the block by height: ", zap.Error(err))
		return nil, err
	}

	logger.Info("Block retrieved successfully by height")
	return &block, nil
}

// GetBlockByHash retrieves a block from the database by its hash.
func (b *BTOrdIdx) GetBlockByHash(blockHash string) (*HookBlock, error) {
	var blockHeight string

	b.rwLock.RLock()         // Acquire the read lock
	defer b.rwLock.RUnlock() // Release the lock when the function returns

	// Retrieve the block height from the database using the block hash
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("bkhash::" + blockHash))
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		// Retrieve the block height
		err = item.Value(func(val []byte) error {
			blockHeight = string(val)
			return nil
		})
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("Failed to get the block height by hash: ", zap.Error(err))
		return nil, err
	}

	// Retrieve the block using the obtained block height
	return b.GetBlockByHeight(blockHeight)
}

// GetBlocks retrieves blocks from the database based on the specified blockHeight and offsetHeight.
func (b *BTOrdIdx) GetBlocks(blockHeight string, offsetHeight string) ([]SimpleHookBlock, error) {
	var blocks []SimpleHookBlock

	startHeight, err := strconv.Atoi(blockHeight)
	if err != nil {
		return nil, fmt.Errorf("invalid block height: %v", err)
	}

	offset, err := strconv.Atoi(offsetHeight)
	if err != nil {
		return nil, fmt.Errorf("invalid offset height: %v", err)
	}

	b.rwLock.RLock()
	defer b.rwLock.RUnlock()

	// Adjust startHeight based on the offset direction
	if offset < 0 {
		startHeight += offset // If offset is negative, move startHeight backward
		offset = -offset      // Make offset positive for iteration
	}

	// Iterate over the blocks based on the offset direction
	for i := 0; i < offset; i++ {
		currentHeight := startHeight
		if offset >= 0 {
			currentHeight += i // Move forward for positive offset
		} else {
			currentHeight -= i // Move backward for negative offset
		}

		// Ensure currentHeight is not negative
		if currentHeight < 0 {
			continue
		}

		var block SimpleHookBlock
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(fmt.Sprintf("block::%d", currentHeight)))
			if err != nil {
				return err
			}

			err = item.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &block)
			})
			return err
		})

		if err != nil {
			// Continue to the next block if an error occurred
			continue
		}

		blocks = append(blocks, block)
	}

	// Check if no blocks were retrieved
	if len(blocks) == 0 {
		return nil, fmt.Errorf("no blocks found within the specified range")
	}

	return blocks, nil
}

// GetAddressInfo retrieves inscription IDs for both MRC721 and MRC20 tokens associated with a given address.
func (b *BTOrdIdx) GetAddressInfo(address string) ([]string, []string, error) {
	var mrc721s, mrc20s []string

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Release lock when the function returns

	// Retrieve MRC721 inscriptions
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc721::addr_inscr::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscription_id from key
			inscriptionID := string(key[len(prefix):])
			mrc721s = append(mrc721s, inscriptionID)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Retrieve MRC20 inscriptions
	err = b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc20::addr_inscr::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscription_id from key
			inscriptionID := string(key[len(prefix):])
			mrc20s = append(mrc20s, inscriptionID)
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return mrc721s, mrc20s, nil
}

// GetAddressBalance retrieves the balance for a specific address and token (tick)
// after trimming any leading and trailing spaces from address and tick.
func (b *BTOrdIdx) GetAddressBalance(address, tick string) (string, error) {
	var balance string

	// Trim leading and trailing spaces from address and tick
	address = strings.TrimSpace(address)
	tick = strings.TrimSpace(tick)

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Release lock when the function returns

	// Construct the key to retrieve the balance
	key := "mrc20::balance::" + address + "::" + tick

	// Retrieve the balance from the database
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		// Retrieve the balance value
		err = item.Value(func(val []byte) error {
			balance = string(val)
			return nil
		})
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("Failed to get the balance: ", zap.String("address", address), zap.String("tick", tick), zap.Error(err))
		return "", err
	}

	logger.Info("Balance retrieved successfully", zap.String("address", address), zap.String("tick", tick))
	return balance, nil
}

// GetAddressBalances retrieves all token balances associated with a given address.
// It returns a JSON string containing an array of balance information.
func (b *BTOrdIdx) GetAddressBalances(address string) (string, error) {
	// BalanceInfo represents the balance of a specific token for an address.
	type BalanceInfo struct {
		Tick    string `json:"tick"`
		Balance string `json:"balance"`
	}

	var balances []BalanceInfo

	// Trim leading and trailing spaces from the address
	address = strings.TrimSpace(address)

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Release lock when the function returns

	// Prefix for the keys to search in the database
	prefix := "mrc20::balance::" + address + "::"

	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = []byte(prefix)
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
			item := it.Item()
			key := item.Key()
			tick := strings.TrimPrefix(string(key), prefix)

			var balance string
			err := item.Value(func(val []byte) error {
				balance = string(val)
				return nil
			})
			if err != nil {
				return err // Error while retrieving the value
			}

			balances = append(balances, BalanceInfo{Tick: tick, Balance: balance})
		}
		return nil
	})

	if err != nil {
		logger.Error("Failed to get one balances: ", zap.String("address", address), zap.Error(err))
		return "", err
	}

	// Marshal the balances slice to JSON
	balancesJSON, err := jsoniter.MarshalToString(balances)
	if err != nil {
		logger.Error("Failed to marshal balances to JSON: ", zap.Error(err))
		return "", err
	}

	return balancesJSON, nil
}

// GetInscription retrieves the inscription information for a given inscription ID.
func (b *BTOrdIdx) GetInscription(id string) (HookInscription, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	var inscription HookInscription

	// Trim leading and trailing spaces from the inscription ID
	id = strings.TrimSpace(id)

	// Construct the key to retrieve the inscription
	key := "inscr::" + id

	// Retrieve the inscription from the database
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		// Deserialize the inscription from the retrieved item
		err = item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &inscription)
		})
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("Failed to get the inscription: ", zap.String("id", id), zap.Error(err))
		return HookInscription{}, err
	}

	return inscription, nil
}

func (b *BTOrdIdx) GetMiningProfitChart(firstMrc721 *MRC721Protocol, step, max int) (string, error) {

	type MiningProfitChartResult struct {
		EndHeight     int    `json:"end_height"`
		EndReason     string `json:"end_reason"`
		ProfitDetails []struct {
			BlockHeight        int    `json:"block_height"`
			MinedFunds         string `json:"mined_funds"`
			PrizePoolFunds     string `json:"prize_pool_funds"`
			TotalReleasedFunds string `json:"total_released_funds"`
		} `json:"profit_details"`
	}

	var result MiningProfitChartResult

	genesisData := Mrc721GenesisData{}
	minerMap := Mrc721MinerMap{Data: make(map[string]*Mrc721MinerData)}

	genesisData.Tick = firstMrc721.Token.Tick
	genesisData.TotalPrizePoolTokens = "0"
	genesisData.BlockHeight = "0"
	genesisData.TotalMinedTokens = "0"
	genesisData.InscriptionsCount = 1

	minerMap.Data["address111111111"] = &Mrc721MinerData{
		InscriptionsID:     "inc111111111",
		InscriptionsNumber: 1,
		Address:            "address111111111",
		BurnNum:            "0",
		Tick:               "demo",
		MinedAmount:        "0",
		Power:              *big.NewInt(1000),
	}

	for i := 0; i < max; i++ {
		blockHeight := strconv.Itoa(i)
		calcResult, err := CalculateMiningRewards(blockHeight, &genesisData, firstMrc721, &minerMap)
		if err != nil {
			return "{}", err
		}

		prizePoolTokensBigInt, _ := new(big.Int).SetString(genesisData.TotalPrizePoolTokens, 10)
		minedTokensBigInt, _ := new(big.Int).SetString(genesisData.TotalMinedTokens, 10)

		currentPrizePoolTokensBigInt, _ := new(big.Int).SetString(calcResult.CurrentPrizePoolAllNum, 10)
		currentMiningTokensBigInt, _ := new(big.Int).SetString(calcResult.CurrentMiningAllNum, 10)

		prizePoolTokensBigInt.Add(prizePoolTokensBigInt, currentPrizePoolTokensBigInt)
		minedTokensBigInt.Add(minedTokensBigInt, currentMiningTokensBigInt)

		var allToken big.Int
		allToken.Add(prizePoolTokensBigInt, minedTokensBigInt)

		// 将累加后的值转换为字符串格式
		genesisData.TotalPrizePoolTokens = prizePoolTokensBigInt.String()
		genesisData.TotalMinedTokens = minedTokensBigInt.String()

		//fmt.Println("GetMiningProfitChart TotalMinedTokens=", genesisData.TotalMinedTokens, calcResult.IsMiningEnd)
		//fmt.Println("GetMiningProfitChart minerMap=", minerMap.Data["address111111111"].MinedAmount)

		if i%step == 0 || calcResult.IsMiningEnd {
			// Adding the profit details to the result slice
			var profitDetail struct {
				BlockHeight        int    `json:"block_height"`
				MinedFunds         string `json:"mined_funds"`
				PrizePoolFunds     string `json:"prize_pool_funds"`
				TotalReleasedFunds string `json:"total_released_funds"`
			}

			profitDetail.BlockHeight = i
			// profitDetail.MinedFunds = genesisData.TotalMinedTokens
			// profitDetail.PrizePoolFunds = genesisData.TotalPrizePoolTokens
			// allToken := new(big.Int).Add(prizePoolTokensBigInt, minedTokensBigInt)
			// profitDetail.TotalReleasedFunds = allToken.String()
			formatFunds := func(funds string) string {

				fundBigInt, ok := new(big.Int).SetString(funds, 10)
				if !ok {
					return "0.000"
				}

				// 首先除以100000000取整
				fundBigInt.Div(fundBigInt, big.NewInt(100000000))

				// 再次除以100000000并保留三位小数
				fundFloat := new(big.Float).SetInt(fundBigInt)
				fundFloat.Quo(fundFloat, big.NewFloat(100000000))

				// 格式化为字符串并保留三位小数
				return fmt.Sprintf("%.3f", fundFloat)
			}

			// 使用formatFunds函数处理资金数值
			profitDetail.MinedFunds = formatFunds(genesisData.TotalMinedTokens)
			profitDetail.PrizePoolFunds = formatFunds(genesisData.TotalPrizePoolTokens)
			allToken := new(big.Int).Add(prizePoolTokensBigInt, minedTokensBigInt)
			profitDetail.TotalReleasedFunds = formatFunds(allToken.String())

			result.ProfitDetails = append(result.ProfitDetails, profitDetail)
		}

		if calcResult.IsMiningEnd {
			result.EndHeight = i
			result.EndReason = calcResult.EndReason
			break
		}

	}

	//fmt.Println("result.ProfitDetails", result.ProfitDetails)

	// Encoding the result to JSON using jsoniter
	jsonBytes, err := jsoniter.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// InscriptionPlus 包含了原始的 HookInscription 和额外的几个字段
type InscriptionPlus struct {
	Inscription HookInscription `json:"Inscription"`
	Burn        string          `json:"burn"`
	Count       string          `json:"count"`
	Power       string          `json:"power"`
}

// GetInscriptionPlus retrieves the inscription information for a given inscription ID, along with additional fields.
func (b *BTOrdIdx) GetInscriptionPlus(id string) (InscriptionPlus, error) {

	var inscription HookInscription

	// Trim leading and trailing spaces from the inscription ID
	id = strings.TrimSpace(id)

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Release lock when the function returns

	// Construct the key to retrieve the inscription
	key := "inscr::" + id

	// Retrieve the inscription from the database
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		// Deserialize the inscription from the retrieved item
		err = item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &inscription)
		})
		if err != nil {
			return err // Returning an error here will abort the transaction
		}

		return nil // Returning nil commits the transaction
	})

	if err != nil {
		logger.Error("Failed to get the inscription: ", zap.String("id", id), zap.Error(err))
		return InscriptionPlus{}, err
	}

	// Retrieve additional fields
	var inscriptionPlus InscriptionPlus
	inscriptionPlus.Inscription = inscription
	inscriptionPlus.Power = "1000" // Set power value to "1000"

	// Retrieve the burn value from the database
	burnKey := "mrc721::burn::" + inscription.ID
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(burnKey))
		if err != nil {
			inscriptionPlus.Burn = "0"
			return nil
		}
		return item.Value(func(val []byte) error {
			inscriptionPlus.Burn = string(val)
			return nil
		})
	})
	if err != nil {
		logger.Error("Failed to get the burn value: ", zap.Error(err))
		// Handle the error accordingly
	}

	// Parse the MRC721 protocol data
	itemMrc721, err := ParseMRC721Protocol(*inscription.ContentByte)
	if err != nil {
		logger.Error("Failed to parse MRC721 protocol: ", zap.Error(err))
		// Handle the error accordingly
	}

	// Retrieve the count value from the database
	countKey := "mrc721::inscr_count::" + itemMrc721.Miner.GetUpperName() + "::" + inscription.ID
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(countKey))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			inscriptionPlus.Count = string(val)
			return nil
		})
	})
	if err != nil {
		logger.Error("Failed to get the count value: ", zap.Error(err))
		// Handle the error accordingly
	}

	if itemMrc721.Burn != nil {
		boostBigInt := stringToPercentageBigInt(itemMrc721.Burn.Boost)
		// Convert the burn and unit values to big.Int and perform the power calculation
		burnBigInt, _ := new(big.Int).SetString(inscriptionPlus.Burn, 10)
		unitBigInt, _ := new(big.Int).SetString(itemMrc721.Burn.Unit, 10)
		powerBigInt := new(big.Int)
		powerBigInt.Div(burnBigInt, unitBigInt)
		powerBigInt.Mul(powerBigInt, boostBigInt)

		// Add 1000 to the calculated power
		powerBigInt.Add(powerBigInt, big.NewInt(1000))
		// Convert the resulting power big.Int back to a string
		inscriptionPlus.Power = powerBigInt.String()
	}

	return inscriptionPlus, nil
}

type Mrc721GenesisDataWeb struct {
	ID                   string `json:"id"`                      // Unique identifier for the inscription
	Number               int    `json:"number"`                  // Number for the inscription
	Name                 string `json:"name"`                    // Name for the inscription
	PrevName             string `json:"previous_name"`           // Previous name for the inscription
	BlockHeight          string `json:"block_height"`            // Height of the block in the blockchain where the inscription was recorded
	GenesisAddress       string `json:"genesis_address"`         // Address associated with the genesis transaction of the inscription
	InscriptionsCount    int    `json:"inscriptions_count"`      // Total count of inscriptions
	InscriptionsMax      int    `json:"inscriptions_max"`        // Total count of inscriptions
	PrizePoolTokens      string `json:"prize_pool_tokens"`       // The amount of tokens that have been added to the prize pool
	TotalMinedTokens     string `json:"mined_tokens"`            // The amount of tokens that have been mined
	TotalPrizePoolTokens string `json:"total_prize_pool_tokens"` // The cumulative total of tokens in the prize pool
	Tick                 string `json:"tick"`                    // The token ticker brc20name
	PrevTick             string `json:"previous_tick"`           // Previous The token ticker brc20name
	Holders              int    `json:"holders"`                 // Total Holders
	Mrc20Holders         int    `json:"mrc20_holders"`           // mrc20_holders
	Mrc20Json            string `json:"mrc20_json"`              // mrc20_json
	Mrc721ImgID          string `json:"mrc721_img_id"`           // mrc721_img_id
}

// // GetAllMrc721 retrieves all MRC-721 genesis inscriptions, sorts them by their Number field, and returns the sorted list.
// func (b *BTOrdIdx) GetAllMrc721() ([]Mrc721GenesisData, error) {
// 	fmt.Println("GetAllMrc721---------")

// 	var genesisDataList []Mrc721GenesisData

// 	// Define the prefix for MRC-721 genesis inscriptions
// 	prefix := []byte("mrc721::geninsc::")

// 	// Start a read-only transaction
// 	err := b.db.View(func(txn *badger.Txn) error {
// 		// Use the Badger iterator to iterate over all keys with the specified prefix
// 		opts := badger.DefaultIteratorOptions
// 		opts.PrefetchValues = true
// 		it := txn.NewIterator(opts)
// 		defer it.Close()

// 		// Iterate over keys
// 		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
// 			item := it.Item()
// 			key := item.Key()

// 			// Extract the MRC-721 name from the key
// 			mrc721Name := string(key[len(prefix):])

// 			fmt.Println("mrc721Name", mrc721Name)

// 			// Retrieve the value (Mrc721GenesisData)
// 			err := item.Value(func(val []byte) error {
// 				var genesisData Mrc721GenesisData
// 				if err := jsoniter.Unmarshal(val, &genesisData); err != nil {
// 					logger.Error("Failed to unmarshal MRC-721 genesis data: ", zap.String("mrc721Name", mrc721Name), zap.Error(err))
// 					return err
// 				}

// 				genesisDataList = append(genesisDataList, genesisData)
// 				return nil
// 			})
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		logger.Error("Failed to retrieve all MRC-721 genesis inscriptions: ", zap.Error(err))
// 		return nil, err
// 	}

// 	// Sort the genesis data list by Number
// 	sort.Slice(genesisDataList, func(i, j int) bool {
// 		return genesisDataList[i].Number < genesisDataList[j].Number
// 	})

// 	return genesisDataList, nil
// }

// GetAllMrc721 retrieves all MRC-721 genesis inscriptions, sorts them by their Number field,
// and returns the sorted list in the form of []Mrc721GenesisDataWeb.
func (b *BTOrdIdx) GetAllMrc721() ([]Mrc721GenesisDataWeb, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// calculateHolders calculates the number of holders for a given MRC-721 name.
	calculateHolders := func(txn *badger.Txn, mrc721Name string) int {

		// Prefix for finding all inscription IDs for a given MRC-721 name.
		inscriptionsPrefix := []byte(fmt.Sprintf("mrc721::name_inscr::%s::", mrc721Name))
		// A map to track unique user addresses.
		uniqueAddresses := make(map[string]struct{})

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		// First, find all inscription IDs for the MRC-721 name.
		for it.Seek(inscriptionsPrefix); it.ValidForPrefix(inscriptionsPrefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extracting inscription ID from the key.
			inscriptionID := string(key[len(inscriptionsPrefix):])

			// Prefix for finding all user addresses for a given inscription ID.
			addressesPrefix := []byte(fmt.Sprintf("mrc721::inscr_addr::%s::", inscriptionID))

			addrIt := txn.NewIterator(badger.DefaultIteratorOptions)
			// Find all user addresses for the inscription ID.
			for addrIt.Seek(addressesPrefix); addrIt.ValidForPrefix(addressesPrefix); addrIt.Next() {
				addrItem := addrIt.Item()
				addrKey := addrItem.Key()
				// Extracting user address from the key.
				userAddr := string(addrKey[len(addressesPrefix):])
				// Adding user address to the map to ensure uniqueness.
				uniqueAddresses[userAddr] = struct{}{}
			}
			addrIt.Close() // Ensure the iterator is closed after use.
		}

		// The number of unique holders is the size of the map.
		return len(uniqueAddresses)
	}

	calculateMrc20Holders := func(txn *badger.Txn, mrc20Name string) int {
		// Prefix for finding all inscriptions associated with the MRC-20 token.
		inscriptionsPrefix := []byte(fmt.Sprintf("mrc20::name_inscr::%s::", mrc20Name))
		uniqueAddresses := make(map[string]struct{}) // Map to record unique addresses.

		// Iterate over all inscriptions for the given MRC-20 token.
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(inscriptionsPrefix); it.ValidForPrefix(inscriptionsPrefix); it.Next() {
			item := it.Item()
			inscriptionID := strings.TrimPrefix(string(item.Key()), string(inscriptionsPrefix))

			// Prefix to find all addresses associated with the inscription.
			addressesPrefix := []byte(fmt.Sprintf("mrc20::inscr_addr::%s::", inscriptionID))

			// Iterate over all addresses for the given inscription.
			addrIt := txn.NewIterator(badger.DefaultIteratorOptions)
			for addrIt.Seek(addressesPrefix); addrIt.ValidForPrefix(addressesPrefix); addrIt.Next() {
				addrItem := addrIt.Item()
				address := strings.TrimPrefix(string(addrItem.Key()), string(addressesPrefix))

				// Record the unique address.
				uniqueAddresses[address] = struct{}{}
			}
			addrIt.Close() // Ensure the iterator is closed after use.
		}

		// Return the count of unique addresses.
		return len(uniqueAddresses)
	}

	var genesisDataWebList []Mrc721GenesisDataWeb

	// Define the prefix for MRC-721 genesis inscriptions
	prefix := []byte("mrc721::geninsc::")

	err := b.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			mrc721Name := string(key[len(prefix):])

			err := item.Value(func(val []byte) error {
				var genesisData Mrc721GenesisData
				if err := jsoniter.Unmarshal(val, &genesisData); err != nil {
					logger.Error("Failed to unmarshal MRC-721 genesis data", zap.String("mrc721Name", mrc721Name), zap.Error(err))
					return err
				}

				Holders := calculateHolders(txn, mrc721Name)
				Mrc20Holders := calculateMrc20Holders(txn, genesisData.Tick)
				Mrc721ImgID, imgErr := b.unlockedFindMrc721ImgID(txn, mrc721Name)
				if imgErr != nil {
					Mrc721ImgID = ""
				}

				// Generate a MRC-20 JSON configuration string using the current MRC-721 genesis data
				Mrc20JsonTemplate := `{	"p": "mrc-20",	"op": "deploy",	"tick": "%s",	"max": "%s","dec": "8" }`
				// Replace placeholders with actual values from the current MRC-721 genesis data
				Mrc20Json := fmt.Sprintf(Mrc20JsonTemplate, genesisData.Tick, genesisData.TotalMinedTokens)
				Mrc20JsonBase64 := base64.StdEncoding.EncodeToString([]byte(Mrc20Json))
				// Print the generated MRC-20 JSON configuration string
				//fmt.Println("MRC-20 JSON:", Mrc20Json)

				// Convert Mrc721GenesisData to Mrc721GenesisDataWeb
				genesisDataWeb := Mrc721GenesisDataWeb{
					ID:                   genesisData.ID,
					Number:               genesisData.Number,
					Name:                 genesisData.Name,
					PrevName:             genesisData.PrevName,
					BlockHeight:          genesisData.BlockHeight,
					GenesisAddress:       genesisData.GenesisAddress,
					InscriptionsCount:    genesisData.InscriptionsCount,
					InscriptionsMax:      genesisData.InscriptionsMax,
					PrizePoolTokens:      genesisData.PrizePoolTokens,
					TotalMinedTokens:     genesisData.TotalMinedTokens,
					TotalPrizePoolTokens: genesisData.TotalPrizePoolTokens,
					Tick:                 genesisData.Tick,
					PrevTick:             genesisData.PrevTick,
					Holders:              Holders,                // Calculate Holders
					Mrc20Holders:         Mrc20Holders + Holders, // Calculate Mrc20Holders
					Mrc20Json:            Mrc20JsonBase64,
					Mrc721ImgID:          Mrc721ImgID,
				}

				// // fix code
				// if genesisDataWeb.Name == "BTC" {
				// 	genesisDataWeb.InscriptionsCount = 862
				// 	genesisDataWeb.Holders = 541
				// }

				genesisDataWebList = append(genesisDataWebList, genesisDataWeb)
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to retrieve all MRC-721 genesis inscriptions", zap.Error(err))
		return nil, err
	}

	sort.Slice(genesisDataWebList, func(i, j int) bool {
		return genesisDataWebList[i].Number < genesisDataWebList[j].Number
	})

	return genesisDataWebList, nil
}

// GetOneMrc721 retrieves a single MRC-721 genesis inscription based on the provided mrc721Name.
// It returns the corresponding Mrc721GenesisDataWeb structure or an error if the retrieval fails.
func (b *BTOrdIdx) GetOneMrc721(mrc721Name string) (Mrc721GenesisDataWeb, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	var genesisDataWeb Mrc721GenesisDataWeb

	// calculateHolders calculates the number of holders for a given MRC-721 name.
	calculateHolders := func(txn *badger.Txn, mrc721Name string) int {
		//fmt.Println("---calculateHolders 1 mrc721Name=", mrc721Name)
		holdersPrefix := []byte(fmt.Sprintf("mrc721::addr_num::%s::", mrc721Name))
		holdersCount := 0

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		for it.Seek(holdersPrefix); it.ValidForPrefix(holdersPrefix); it.Next() {
			//fmt.Println("---calculateHolders 2")
			holdersCount++
		}

		return holdersCount
	}

	prefix := []byte(fmt.Sprintf("mrc721::geninsc::%s", mrc721Name))
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(prefix)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return fmt.Errorf("MRC-721 genesis inscription with name '%s' not found", mrc721Name)
			}
			return err
		}

		return item.Value(func(val []byte) error {
			var genesisData Mrc721GenesisData
			if err := jsoniter.Unmarshal(val, &genesisData); err != nil {
				logger.Error("Failed to unmarshal MRC-721 genesis data", zap.String("mrc721Name", mrc721Name), zap.Error(err))
				return err
			}

			genesisDataWeb = Mrc721GenesisDataWeb{
				ID:                   genesisData.ID,
				Number:               genesisData.Number,
				Name:                 genesisData.Name,
				PrevName:             genesisData.PrevName,
				BlockHeight:          genesisData.BlockHeight,
				GenesisAddress:       genesisData.GenesisAddress,
				InscriptionsCount:    genesisData.InscriptionsCount,
				InscriptionsMax:      genesisData.InscriptionsMax,
				PrizePoolTokens:      genesisData.PrizePoolTokens,
				TotalMinedTokens:     genesisData.TotalMinedTokens,
				TotalPrizePoolTokens: genesisData.TotalPrizePoolTokens,
				Tick:                 genesisData.Tick,
				PrevTick:             genesisData.PrevTick,
				Holders:              calculateHolders(txn, mrc721Name),
			}
			return nil
		})
	})

	if err != nil {
		logger.Error("Failed to retrieve MRC-721 genesis inscription", zap.String("mrc721Name", mrc721Name), zap.Error(err))
		return Mrc721GenesisDataWeb{}, err
	}

	return genesisDataWeb, nil
}

// GetAddressMrc721List retrieves a paginated list of WebInscription for a given address and MRC721 name.
// It fetches a list of inscription IDs using the address prefix, then fetches the corresponding HookInscription details.
// The results are paginated based on provided pageIndex and pageSize, and sorted by HookInscription.Number in ascending order.
func (b *BTOrdIdx) GetAddressMrc721List(address string, mrc721name string, pageIndex int, pageSize int) ([]WebInscription, int, error) {
	var inscriptions []HookInscription
	var allCount int
	allCount = 0

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// Retrieve MRC721 inscriptions using the address prefix
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc721::addr_inscr::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscriptionID from the key
			inscriptionID := string(key[len(prefix):])
			//fmt.Println("GetAddressMrc721List inscriptionID=", inscriptionID)

			// Fetch HookInscription details using inscriptionID
			var hookInscription HookInscription
			err := b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("inscr::" + inscriptionID))
				if err != nil {
					return err
				}
				return item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
			})
			if err != nil {
				return err
			}

			if mrc721name == "" {
				inscriptions = append(inscriptions, hookInscription)
			} else {
				name := ""
				name, _, err = ConvertToNameID(*hookInscription.ContentByte)
				if err != nil {
					mrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
					if err != nil {
					} else {
						name = mrc721.Miner.GetUpperName()
					}
				}

				if mrc721name == name {
					inscriptions = append(inscriptions, hookInscription)
				}

			}

			//fmt.Println("GetAddressMrc721List hookInscription=", hookInscription)
		}
		return nil
	})
	if err != nil {
		return nil, allCount, err
	}

	// Sort the inscriptions by Number in ascending order
	sort.Slice(inscriptions, func(i, j int) bool {
		return inscriptions[i].Number < inscriptions[j].Number
	})

	allCount = len(inscriptions)

	// Apply pagination
	// Compute the start and end indices for the slice of inscriptions
	start := pageIndex * pageSize
	if start > len(inscriptions) {
		start = len(inscriptions) // Ensure start is within the slice bounds
	}
	end := start + pageSize
	if end > len(inscriptions) {
		end = len(inscriptions) // Ensure end is within the slice bounds
	}
	paginatedInscriptions := inscriptions[start:end]

	// Convert HookInscription to WebInscription
	var webInscriptions []WebInscription
	for _, insc := range paginatedInscriptions {
		// Fetch MinedAmount and Power for each inscription
		var minedAmount, power *big.Int
		minedAmount, _ = new(big.Int).SetString("0", 10)
		power, _ = new(big.Int).SetString("0", 10)

		// Fetch MinedAmount
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("mrc721::inscr_miner::" + insc.ID))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				minedAmount = new(big.Int)
				minedAmount.SetBytes(val)
				return nil
			})
		})
		if err != nil {
			logger.Info("Failed to get the mined amount: ", zap.String("id", insc.ID), zap.Error(err))
		}

		// Fetch Power
		err = b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("mrc721::inscr_power::" + insc.ID))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				power = new(big.Int)
				power.SetBytes(val)
				return nil
			})
		})
		if err != nil {
			logger.Info("Failed to get the power: ", zap.String("id", insc.ID), zap.Error(err))
		}

		name := ""
		name, _, err = ConvertToNameID(*insc.ContentByte)
		if err != nil {
			mrc721, err := ParseMRC721Protocol(*insc.ContentByte)
			if err != nil {
			} else {
				name = mrc721.Miner.GetUpperName()
			}
		}

		// Fetch MRC721 genesis data for the given mrc721name
		var mrc721GenesisData Mrc721GenesisData
		err = b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("mrc721::geninsc::" + name))
			if err != nil {
				return err
			}
			return item.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &mrc721GenesisData)
			})
		})
		if err != nil {
			// handle error, e.g., log or return
			return nil, allCount, err
		}

		mrc20name := mrc721GenesisData.Tick

		webInscriptions = append(webInscriptions, WebInscription{
			Inscription: insc,
			Mrc721name:  name,
			Mrc20name:   mrc20name,
			MinedAmount: minedAmount.String(), // Convert BigInt to string
			Power:       power.String(),       // Convert BigInt to string
		})

	}

	return webInscriptions, allCount, nil
}

func (b *BTOrdIdx) GetAddressMrc721Bar(address string) ([]string, error) {
	var inscriptions []HookInscription

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// Retrieve MRC721 inscriptions using the address prefix
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc721::addr_inscr::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscriptionID from the key
			inscriptionID := string(key[len(prefix):])
			//fmt.Println("GetAddressMrc721List inscriptionID=", inscriptionID)

			// Fetch HookInscription details using inscriptionID
			var hookInscription HookInscription
			err := b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("inscr::" + inscriptionID))
				if err != nil {
					return err
				}
				return item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
			})
			if err != nil {
				return err
			}

			inscriptions = append(inscriptions, hookInscription)

			//fmt.Println("GetAddressMrc721List hookInscription=", hookInscription)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Sort the inscriptions by Number in ascending order
	sort.Slice(inscriptions, func(i, j int) bool {
		return inscriptions[i].Number < inscriptions[j].Number
	})

	nameSet := make(map[string]struct{}) // Create a set to store unique names
	var names []string                   // Slice to store the final list of unique names

	for _, insc := range inscriptions {
		var name string
		name, _, err = ConvertToNameID(*insc.ContentByte)
		if err != nil {
			mrc721, err := ParseMRC721Protocol(*insc.ContentByte)
			if err != nil {
				// Handle error if needed
			} else {
				name = mrc721.Miner.GetUpperName()
			}
		}

		if _, exists := nameSet[name]; !exists && name != "" {
			nameSet[name] = struct{}{}  // Add name to the set if it's not already present
			names = append(names, name) // Append name to the slice
		}
	}

	return names, nil
}

func (b *BTOrdIdx) GetMrc721Collections(mrc721name string) ([]WebCollections, error) {
	var collections []WebCollections

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// Search for all inscription IDs associated with the given MRC721 name
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc721::name_inscr::" + mrc721name + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()

			// Extract inscriptionID from the key
			inscriptionID := string(key[len(prefix):])

			// Fetch HookInscription details using inscriptionID
			var hookInscription HookInscription
			err := b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("inscr::" + inscriptionID))
				if err != nil {
					return err
				}
				return item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
			})
			if err != nil {
				return err
			}

			// Create a WebCollection item for each inscription
			collection := WebCollections{
				Id: inscriptionID,
				Meta: WebCollectionItem{
					Name: mrc721name + " #" + strconv.Itoa(hookInscription.Number),
				},
			}

			collections = append(collections, collection)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return collections, nil
}

// GetValidateMRC721OrMRC20Name checks the existence of a key in the database based on the provided name and kind.
// If kind is "mrc721", it checks for the key "mrc721::geninsc::[name]".
// If kind is "mrc20", it checks for the key "mrc20::geninsc::[name]".
// Returns true if the key exists, false otherwise.
func (b *BTOrdIdx) GetValidateMRC721OrMRC20Name(name string, kind string) (bool, error) {
	// Construct the key based on the kind and name
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	var key string
	if kind == "mrc721" {
		key = "mrc721::geninsc::" + name
	} else if kind == "mrc20" {
		key = "mrc20::geninsc::" + name
	} else {
		return false, errors.New("invalid kind: must be 'mrc721' or 'mrc20'")
	}

	// Check if the key exists in the database
	exists := false
	err := b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err == badger.ErrKeyNotFound {
			return nil // Key does not exist
		}
		if err != nil {
			return err // An error occurred
		}

		exists = true // Key exists
		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

// GetGenesisMRC721Protocol retrieves the MRC721 protocol data for a given MRC721 name.
// It first fetches the Mrc721GenesisData using the key mrc721::geninsc::[mrc721name],
// then retrieves the HookInscription data using the key inscr::[Mrc721GenesisData.ID],
// and finally parses the MRC721 protocol data using ParseMRC721Protocol.
func (b *BTOrdIdx) GetGenesisMRC721Protocol(mrc721name string) (*MRC721Protocol, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	var genesisData Mrc721GenesisData
	var inscription HookInscription

	// Retrieve Mrc721GenesisData from the database
	genesisKey := "mrc721::geninsc::" + mrc721name
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(genesisKey))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &genesisData)
		})
	})
	if err != nil {
		return nil, err
	}

	// Retrieve HookInscription data using the genesis data ID
	inscriptionKey := "inscr::" + genesisData.ID
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(inscriptionKey))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &inscription)
		})
	})
	if err != nil {
		return nil, err
	}

	// Parse the MRC721 protocol from the inscription content
	return ParseMRC721Protocol(*inscription.ContentByte)
}

// GetAddressMrc20Bar retrieves the balance for all MRC20 tokens for a specific address.
// It constructs a slice of WebMrc20Bar, each representing a token's name, available balance, and transferable balance.
func (b *BTOrdIdx) GetAddressMrc20Bar(address string, mrc20Name string) ([]WebMrc20Bar, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	var mrc20Bars []WebMrc20Bar
	mrc20Bars = make([]WebMrc20Bar, 0)

	// Retrieve all MRC20 tick information using the address prefix
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc20::balance::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()

			// Extract tick (token name) from the key
			tick := string(key[len(prefix):])

			// Retrieve the balance for the tick
			var balance string
			err := item.Value(func(val []byte) error {
				balance = string(val)
				return nil
			})
			if err != nil {
				return err
			}

			// Initialize total transfer amount as big.Int
			totalTransferAmount := big.NewInt(0)

			// Use the prefix to find all inscription IDs for the address
			inscrPrefix := "mrc20::addr_inscr::" + address + "::"
			itInscr := txn.NewIterator(badger.DefaultIteratorOptions)
			defer itInscr.Close()
			for itInscr.Seek([]byte(inscrPrefix)); itInscr.ValidForPrefix([]byte(inscrPrefix)); itInscr.Next() {
				itemInscr := itInscr.Item()
				inscriptionID := string(itemInscr.Key()[len(inscrPrefix):])

				// Fetch HookInscription details using inscriptionID
				var hookInscription HookInscription
				itemInscr, err := txn.Get([]byte("inscr::" + inscriptionID))
				if err != nil {
					return err
				}
				err = itemInscr.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
				if err != nil {
					return err
				}

				// Parse MRC20Protocol from HookInscription.ContentByte
				mrc20data, err := ParseMRC20Protocol(*hookInscription.ContentByte)
				if err != nil {
					return err
				}

				if tick == mrc20data.Tick {
					// Convert mrc20data.Amt to big.Int and accumulate
					amtBigInt := new(big.Int)
					amtBigInt, ok := amtBigInt.SetString(mrc20data.Amt, 10)
					if !ok {
						return errors.New("Failed to convert Amt to big.Int")
					}
					totalTransferAmount.Add(totalTransferAmount, amtBigInt)
				}
			}

			// Convert total transfer amount to string
			totalTransferStr := totalTransferAmount.String()

			// Convert balance and total transfer amount to big.Int for subtraction
			balanceBigInt := new(big.Int)
			balanceBigInt, ok := balanceBigInt.SetString(balance, 10)
			if !ok {
				return errors.New("Failed to convert Balance to big.Int")
			}

			// Calculate the available amount (Balance + Transferable)
			availableBigInt := new(big.Int).Add(balanceBigInt, totalTransferAmount)
			availableStr := availableBigInt.String()

			// Construct WebMrc20Bar and append it to the slice
			mrc20Bar := WebMrc20Bar{
				Mrc20name:    tick,
				Balance:      availableStr,
				Avaliable:    balance,          // Set the available amount
				Transferable: totalTransferStr, // Set the accumulated transfer amount
			}

			fmt.Println("****mrc20Name,mrc20data.Tick=", mrc20Name, tick)
			if mrc20Name != "" && tick != mrc20Name {
				continue
			}

			mrc20Bars = append(mrc20Bars, mrc20Bar)
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to get MRC20 bars: ", zap.String("address", address), zap.Error(err))
		return nil, err
	}

	logger.Info("MRC20 bars retrieved successfully", zap.String("address", address))
	return mrc20Bars, nil
}

// GetAddressMrc20List retrieves a paginated list of WebMrc20Inscription for a given address.
// It fetches a list of inscription IDs using the address prefix, then fetches the corresponding HookInscription details.
// The results are filtered by MRC20 token name (mrc20name) and paginated based on provided pageIndex and pageSize.
// Only inscriptions matching the specified mrc20name are included in the result. If mrc20name is empty, no filtering is applied.
func (b *BTOrdIdx) GetAddressMrc20List(address string, mrc20name string, pageIndex int, pageSize int) ([]WebMrc20Inscription, int, error) {
	var inscriptions []HookInscription
	var mrc20Inscriptions []WebMrc20Inscription
	allCount := 0

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// Retrieve inscriptions using the address prefix
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc20::addr_inscr::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscriptionID from the key
			inscriptionID := string(key[len(prefix):])

			// Fetch HookInscription details using inscriptionID
			var hookInscription HookInscription
			err := b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("inscr::" + inscriptionID))
				if err != nil {
					return err
				}
				return item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
			})
			if err != nil {
				return err
			}

			// Parse MRC20Protocol from HookInscription.ContentByte
			mrc20data, err := ParseMRC20Protocol(*hookInscription.ContentByte)
			if err != nil {
				return err
			}

			// Filter by mrc20name if it's not empty
			if mrc20name == "" || mrc20name == mrc20data.Tick {
				inscriptions = append(inscriptions, hookInscription)

				// Construct WebMrc20Inscription and add to the list
				webMrc20Inscription := WebMrc20Inscription{
					Inscription: hookInscription,
					Mrc20name:   mrc20data.Tick,
					Amount:      mrc20data.Amt,
				}
				mrc20Inscriptions = append(mrc20Inscriptions, webMrc20Inscription)
			}
		}
		return nil
	})
	if err != nil {
		return nil, allCount, err
	}

	allCount = len(mrc20Inscriptions)

	// Apply pagination
	start := pageIndex * pageSize
	if start > len(mrc20Inscriptions) {
		start = len(mrc20Inscriptions) // Ensure start is within the slice bounds
	}
	end := start + pageSize
	if end > len(mrc20Inscriptions) {
		end = len(mrc20Inscriptions) // Ensure end is within the slice bounds
	}
	paginatedInscriptions := mrc20Inscriptions[start:end]

	return paginatedInscriptions, allCount, nil
}

func (b *BTOrdIdx) GetAddressMrc721BarPlus(address string) ([]WebMrc721Bar, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// Handle the panic, convert it to error if needed
	// 		err := fmt.Errorf("write newBlock failed: %v", r)
	// 		logger.Error("panic panic panic: ", zap.Error(err))

	// 	}
	// }()

	var webMrc721Bars []WebMrc721Bar
	mrc721Stats := make(map[string]*WebMrc721Bar)
	//First721ID := ""

	// Retrieve inscriptions using the address prefix
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc721::addr_inscr::" + address + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscriptionID from the key
			inscriptionID := string(key[len(prefix):])

			// Fetch HookInscription details using inscriptionID
			var hookInscription HookInscription
			err := b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("inscr::" + inscriptionID))
				if err != nil {
					return fmt.Errorf(" GetAddressMrc721BarPlus error1 : %w", err)
				}
				return item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
			})
			if err != nil {
				return fmt.Errorf("GetAddressMrc721BarPlus error2 : %w", err)
			}

			mrc721name := ""
			mrc721name, _, err = ConvertToNameID(*hookInscription.ContentByte)
			if err != nil {
				mrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
				if err != nil {
				} else {
					mrc721name = mrc721.Miner.GetUpperName()
				}
			}

			// Initialize or update statistics for the MRC721 name
			if _, exists := mrc721Stats[mrc721name]; !exists {
				mrc721Stats[mrc721name] = &WebMrc721Bar{
					Mrc721name:  mrc721name,
					Amount:      "0",
					TotalPower:  "0",
					TotalReward: "0",
					First721ID:  inscriptionID,
					BlockHeight: hookInscription.BlockHeight,
				}
			}
			mrc721Bar := mrc721Stats[mrc721name]

			// Increment amount
			amountBigInt, _ := new(big.Int).SetString(mrc721Bar.Amount, 10)
			amountBigInt = amountBigInt.Add(amountBigInt, big.NewInt(1))
			mrc721Bar.Amount = amountBigInt.String()

			// Accumulate power
			powerBigInt, _ := new(big.Int).SetString(mrc721Bar.TotalPower, 10)
			powerItem, err := txn.Get([]byte("mrc721::inscr_power::" + inscriptionID))
			if err == nil {
				err = powerItem.Value(func(val []byte) error {
					powerBigInt = powerBigInt.Add(powerBigInt, new(big.Int).SetBytes(val))
					return nil
				})
				if err != nil {
					return fmt.Errorf(" GetAddressMrc721BarPlus error4 : %w", err)
				}
			} else {
				logger.Info(" GetAddressMrc721BarPlus mrc721::inscr_power::inscriptionID error", zap.Error(err))
			}

			// if err != nil {
			// 	//return fmt.Errorf(" GetAddressMrc721BarPlus error3 : %w", err)
			// }

			mrc721Bar.TotalPower = powerBigInt.String()

			//  Accumulate reward (similar to power accumulation)
			rewardBigInt, _ := new(big.Int).SetString(mrc721Bar.TotalReward, 10)
			rewardItem, err := txn.Get([]byte("mrc721::inscr_miner::" + inscriptionID))
			if err == nil {
				err = rewardItem.Value(func(val []byte) error {
					rewardBigInt = rewardBigInt.Add(rewardBigInt, new(big.Int).SetBytes(val))
					return nil
				})
				if err != nil {
					return fmt.Errorf(" GetAddressMrc721BarPlus error5 : %w", err)
				}
				//fmt.Println("rewardBigInt=", rewardBigInt.String())
				mrc721Bar.TotalReward = rewardBigInt.String()
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf(" GetAddressMrc721BarPlus error6 : %w", err)
	}

	// // Convert map to slice
	// for _, bar := range mrc721Stats {
	// 	webMrc721Bars = append(webMrc721Bars, *bar)
	// }

	for _, mrc721Bar := range mrc721Stats {
		// Fetch MRC721GenesisData for the Mrc721name
		var mrc721GenesisData Mrc721GenesisData
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("mrc721::geninsc::" + mrc721Bar.Mrc721name))
			if err != nil {
				return err // handle error (e.g., not found)
			}
			return item.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &mrc721GenesisData)
			})
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get Mrc721GenesisData for %s: %w", mrc721Bar.Mrc721name, err)
		}

		// Fill in the Mrc20name from the Mrc721GenesisData.Tick
		mrc721Bar.Mrc20name = mrc721GenesisData.Tick
		webMrc721Bars = append(webMrc721Bars, *mrc721Bar)
	}

	sort.Slice(webMrc721Bars, func(i, j int) bool {
		return webMrc721Bars[i].BlockHeight > webMrc721Bars[j].BlockHeight
	})

	return webMrc721Bars, nil
}

// GetAddressMrc721Holders retrieves a paginated list of WebMrc721Holder for a given MRC721 name.
// It fetches a list of unique addresses associated with each inscription ID, groups them by address,
// and computes the total count of inscriptions for each address. The results are paginated based on
// provided pageIndex and pageSize, and sorted by the number of inscriptions per address in descending order.
func (b *BTOrdIdx) GetAddressMrc721Holders(mrc721name string, pageIndex int, pageSize int) ([]WebMrc721Holder, int, error) {
	addressCountMap := make(map[string]int)
	var totalInscriptions int

	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// Step 1: Retrieve all inscription IDs associated with the given MRC721 name
	err := b.db.View(func(txn *badger.Txn) error {
		prefix := []byte("mrc721::name_inscr::" + mrc721name + "::")
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			// Extract inscriptionID from the key
			inscriptionID := string(key[len(prefix):])

			// Step 2: Fetch the unique address associated with each inscriptionID
			addressPrefix := []byte("mrc721::inscr_addr::" + inscriptionID + "::")
			err := b.db.View(func(txn *badger.Txn) error {
				opts := badger.DefaultIteratorOptions
				opts.Prefix = addressPrefix
				it := txn.NewIterator(opts)
				defer it.Close()

				for it.Seek(addressPrefix); it.ValidForPrefix(addressPrefix); it.Next() {
					addressItem := it.Item()
					addressKey := addressItem.Key()
					address := string(addressKey[len(addressPrefix):])
					addressCountMap[address]++
					totalInscriptions++
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	// Convert the addressCountMap to a slice of WebMrc721Holder
	var holders []WebMrc721Holder
	for address, count := range addressCountMap {
		holders = append(holders, WebMrc721Holder{
			Address:    address,
			Amount:     strconv.Itoa(count),
			Percentage: "", // Will be calculated later
			Rank:       "", // Will be calculated later
		})
	}

	// // Sort the holders by the number of inscriptions in descending order
	// sort.Slice(holders, func(i, j int) bool {
	// 	return holders[i].Amount > holders[j].Amount
	// })
	sort.Slice(holders, func(i, j int) bool {
		iAmount, err := strconv.Atoi(holders[i].Amount)
		if err != nil {
			log.Fatalf("Failed to convert Amount to int: %s", err)
		}

		jAmount, err := strconv.Atoi(holders[j].Amount)
		if err != nil {
			log.Fatalf("Failed to convert Amount to int: %s", err)
		}

		return iAmount > jAmount
	})

	// Assign rank and calculate percentage
	for i, holder := range holders {
		holders[i].Rank = strconv.Itoa(i + 1)
		amount, _ := strconv.Atoi(holder.Amount)
		percentage := float64(amount) / float64(totalInscriptions) * 100
		holders[i].Percentage = fmt.Sprintf("%.2f%%", percentage)
	}

	// Apply pagination
	start := pageIndex * pageSize
	if start > len(holders) {
		start = len(holders)
	}
	end := start + pageSize
	if end > len(holders) {
		end = len(holders)
	}
	paginatedHolders := holders[start:end]

	return paginatedHolders, len(holders), nil
}

// ScanMissingBlocks scans the range of blocks from 'begin' to 'end' (inclusive)
// and returns a list of block numbers that are missing in the database or have no inscription.
// It ensures that 'end' is greater than 'begin' and iteratively checks the existence
// of each block by its key 'block::i'. If a key does not exist, or if the block's hash equals NO_INSCRIPTION_BLOCK_HASH,
// it means the block is considered "missing", and its number is added to the result list.
func (b *BTOrdIdx) ScanMissingBlocks(begin, end int) ([]int, error) {
	// Check if 'end' is greater than 'begin'.
	if end <= begin {
		return nil, errors.New("end must be greater than begin")
	}

	// Initialize an empty list to store the numbers of missing or no-inscription blocks.
	var missingOrNoInscriptionBlocks []int

	// Iterate through each block number from 'begin' to 'end'.
	for i := begin; i <= end; i++ {
		blockExists := false // Flag to check if block exists

		// Attempt to retrieve the block from the database.
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(fmt.Sprintf("block::%d", i)))
			if err == badger.ErrKeyNotFound {
				// The block does not exist.
				blockExists = false
			} else if err != nil {
				// An unexpected error occurred while retrieving the block.
				return err
			} else {
				// The block exists, check if it has no inscription.
				var hookBlock HookBlock
				err := item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookBlock)
				})
				if err != nil {
					return err // Return error if unmarshalling fails.
				}

				// Check if the block's hash equals NO_INSCRIPTION_BLOCK_HASH.
				if hookBlock.BlockHash == NO_INSCRIPTION_BLOCK_HASH {
					// Block exists but has no inscription.
					blockExists = false
				} else {
					// Block exists and has an inscription.
					blockExists = true
				}
			}
			return nil
		})

		if err != nil {
			// Return the encountered error.
			return nil, err
		}

		// If the block does not exist or has no inscription, add its number to the list.
		if !blockExists {
			missingOrNoInscriptionBlocks = append(missingOrNoInscriptionBlocks, i)
		}
	}

	// Return the list of missing or no-inscription blocks after checking all blocks in the range.
	return missingOrNoInscriptionBlocks, nil
}

// GetGenesisData retrieves genesis data for a given MRC-721 name and parses it into Mrc721GenesisData structure.
// It fetches the data from the database using the key constructed with mrc721name and returns the parsed result.
func (b *BTOrdIdx) GetGenesisData(mrc721name string) (Mrc721GenesisData, error) {
	var genesisData Mrc721GenesisData

	// Construct the key for fetching genesis data of the MRC-721 name
	key := "mrc721::geninsc::" + mrc721name

	// Retrieve the genesis data from the database
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			// Return error if the key does not exist or any other issue with fetching data
			return fmt.Errorf("error retrieving genesis data for MRC-721 name '%s': %w", mrc721name, err)
		}

		// Extract and parse the data into the Mrc721GenesisData structure
		err = item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &genesisData)
		})
		if err != nil {
			// Return error if parsing the data fails
			return fmt.Errorf("error parsing genesis data for MRC-721 name '%s': %w", mrc721name, err)
		}

		return nil
	})

	// Return the parsed genesis data or error
	if err != nil {
		return Mrc721GenesisData{}, err
	}
	return genesisData, nil
}

// GetBurnInfo retrieves burn information for a given inscription ID.
// It fetches necessary details from several data sources to construct the WebBurnInfo object.
func (b *BTOrdIdx) GetBurnInfo(inscriptionID string) (WebBurnInfo, error) {
	b.rwLock.RLock()         // Acquire read lock
	defer b.rwLock.RUnlock() // Ensure lock is released after the function execution

	// Initialize the return object
	var webBurnInfo WebBurnInfo

	// Retrieve HookInscription details using inscriptionID
	var hookInscription HookInscription
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("inscr::" + inscriptionID))
		if err != nil {
			return fmt.Errorf("GetBurnInfo error1: %w", err)
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &hookInscription)
		})
	})
	if err != nil {
		return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error2: %w", err)
	}

	// Extract MRC721 name from the HookInscription
	mrc721name := ""
	mrc721name, _, err = ConvertToNameID(*hookInscription.ContentByte)
	if err != nil {
		mrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
		if err != nil {
			return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error3: %w", err)
		} else {
			mrc721name = mrc721.Miner.GetUpperName()
		}
	}

	// Retrieve Mrc721GenesisData
	var genesisData Mrc721GenesisData
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("mrc721::geninsc::" + mrc721name))
		if err != nil {
			return fmt.Errorf("GetBurnInfo error4: %w", err)
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &genesisData)
		})
	})
	if err != nil {
		return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error5: %w", err)
	}

	var genInscription HookInscription
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("inscr::" + genesisData.ID))
		if err != nil {
			return fmt.Errorf("GetBurnInfo error5.2: %w", err)
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &genInscription)
		})
	})
	if err != nil {
		return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error5.5: %w", err)
	}

	// Extract genesis inscription data
	var genInscMrc721 *MRC721Protocol
	genInscMrc721, err = ParseMRC721Protocol(*genInscription.ContentByte)
	if err != nil {
		return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error6: %w", err)
	}

	webBurnInfo.Mrc721name = mrc721name
	webBurnInfo.Mrc20name = genInscMrc721.Token.Tick
	webBurnInfo.Unit = "0"
	webBurnInfo.Boost = "0"
	webBurnInfo.BurnMax = "0"
	if genInscMrc721.Burn != nil {
		webBurnInfo.Unit = genInscMrc721.Burn.Unit
		webBurnInfo.Boost = genInscMrc721.Burn.Boost

		// Calculate BurnMax
		unitBigInt, _ := new(big.Int).SetString(genInscMrc721.Burn.Unit, 10)
		boostBigInt := stringToPercentageBigInt(genInscMrc721.Burn.Boost)
		webBurnInfo.BurnMax = new(big.Int).Div(big.NewInt(11000), boostBigInt).Mul(unitBigInt, big.NewInt(10)).String()

		// webBurnInfo.BurnMax = unitBigInt.Mul(unitBigInt, big.NewInt(10)).String()

	}

	// Retrieve balance
	err = b.db.View(func(txn *badger.Txn) error {
		balanceItem, err := txn.Get([]byte("mrc20::balance::" + hookInscription.Address + "::" + genInscMrc721.Token.Tick))
		if err != nil {
			return fmt.Errorf("GetBurnInfo error7: %w", err)
		}
		return balanceItem.Value(func(val []byte) error {
			webBurnInfo.Balance = string(val)
			return nil
		})
	})
	if err != nil {
		webBurnInfo.Balance = "0"
		//return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error8: %w", err)
	}

	// Retrieve power
	err = b.db.View(func(txn *badger.Txn) error {
		powerItem, err := txn.Get([]byte("mrc721::inscr_power::" + inscriptionID))
		if err != nil {
			return fmt.Errorf("GetBurnInfo error9: %w", err)
		}
		return powerItem.Value(func(val []byte) error {
			powerBigInt := new(big.Int).SetBytes(val)
			webBurnInfo.Power = powerBigInt.String()
			return nil
		})
	})
	if err != nil {
		webBurnInfo.Power = "0"
		//return WebBurnInfo{}, fmt.Errorf("GetBurnInfo error10: %w", err)
	}

	// Retrieve burn amount
	err = b.db.View(func(txn *badger.Txn) error {
		burnAmountItem, err := txn.Get([]byte("mrc721::burn::" + inscriptionID))
		if err != nil {
			return fmt.Errorf("GetBurnInfo error11: %w", err)
		}
		return burnAmountItem.Value(func(val []byte) error {
			burnAmountBigInt := new(big.Int).SetBytes(val)
			webBurnInfo.BurnAmount = burnAmountBigInt.String()
			return nil
		})
	})
	if err != nil {
		webBurnInfo.BurnAmount = "0"
	}

	return webBurnInfo, nil
}

// GetMrcAllInscription retrieves all relevant data related to an inscription and organizes it into the WebMrcAllInscription structure.
func (b *BTOrdIdx) GetMrcAllInscription(inscriptionID string) (WebMrcAllInscription, error) {
	var result WebMrcAllInscription

	// Acquire read lock
	b.rwLock.RLock()
	defer b.rwLock.RUnlock()

	// Retrieve HookInscription details using inscriptionID
	var hookInscription HookInscription
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("inscr::" + inscriptionID))
		if err != nil {
			return fmt.Errorf("GetMrcAllInscription error1: %w", err)
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &hookInscription)
		})
	})
	if err != nil {
		result.MrcType = "unknown"
		return result, nil
	}
	result.Inscription = hookInscription

	// Check for MRC721 type using the prefix method
	var genesisData Mrc721GenesisData
	found, err := b.checkAndRetrieveMRC721(inscriptionID, hookInscription, &genesisData)
	if err != nil {
		return WebMrcAllInscription{}, err
	}
	if found {
		result.MrcType = "mrc721"
		result.GenesisData = genesisData

		// Retrieve genesis inscription and parse MRC721Protocol
		var genesisInscription HookInscription
		err = b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte("inscr::" + genesisData.ID))
			if err != nil {
				return fmt.Errorf("GetMrcAllInscription error3: %w", err)
			}
			return item.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &genesisInscription)
			})
		})
		if err != nil {
			return WebMrcAllInscription{}, fmt.Errorf("GetMrcAllInscription error4: %w", err)
		}

		mrc721p, err := ParseMRC721Protocol(*genesisInscription.ContentByte)
		if err != nil {
			return WebMrcAllInscription{}, fmt.Errorf("GetMrcAllInscription error5: %w", err)
		}
		result.Mrc721P = mrc721p

	} else {
		// If MRC721 data not found, check for MRC20 type using the prefix method
		var mrc20p *MRC20Protocol
		mrc20p, err = b.checkAndRetrieveMRC20(inscriptionID, &hookInscription)
		if err != nil {
			return WebMrcAllInscription{}, err
		}
		if mrc20p != nil {
			result.MrcType = "mrc20"
			result.Mrc20P = mrc20p
			mrc721name := ""
			// Retrieve Mrc721GenesisData for the MRC20 token
			err = b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("mrc20::geninsc::" + mrc20p.Tick))
				if err != nil {
					return fmt.Errorf("GetMrcAllInscription error6: %w", err)
				}
				return item.Value(func(val []byte) error {
					mrc721name = string(val)
					return nil
				})

			})
			if err != nil {
				return WebMrcAllInscription{}, fmt.Errorf("GetMrcAllInscription error7: %w", err)
			}

			// Retrieve Mrc721GenesisData for the MRC20 token
			err = b.db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte("mrc721::geninsc::" + mrc721name))
				if err != nil {
					return fmt.Errorf("GetMrcAllInscription error6: %w", err)
				}
				return item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &genesisData)
				})
			})
			if err != nil {
				return WebMrcAllInscription{}, fmt.Errorf("GetMrcAllInscription error7: %w", err)
			}
			result.GenesisData = genesisData
		} else {
			// If neither MRC721 nor MRC20 data found, set MrcType to nomrc
			result.MrcType = "inscription"
		}
	}

	return result, nil
}

// checkAndRetrieveMRC721 checks if the given inscriptionID belongs to an MRC-721 inscription.
// If it does, the function retrieves the associated Mrc721GenesisData from the database.
func (b *BTOrdIdx) checkAndRetrieveMRC721(inscriptionID string, hookInscription HookInscription, genesisData *Mrc721GenesisData) (bool, error) {
	found := false
	//fmt.Println("checkAndRetrieveMRC721 AA inscriptionID =", inscriptionID)

	// Use BadgerDB's View transaction to perform read operations.
	err := b.db.View(func(txn *badger.Txn) error {
		// Use prefix search to find the key. The prefix is "mrc721::inscr_addr::" concatenated with the inscriptionID.
		prefix := []byte("mrc721::inscr_addr::" + inscriptionID)
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		opts.PrefetchSize = 10 // Adjust this value based on expected number of records with the same prefix.
		it := txn.NewIterator(opts)
		defer it.Close()

		// Iterate over keys with the given prefix.
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			//item := it.Item()
			//k := item.Key()
			//fmt.Printf("Key=%s, ValueSize=%d\n", k, item.ValueSize())
			found = true
			// If necessary, item's value can be fetched here.
			// However, as per the original logic, we just need to know if the key exists.
			break // Once a matching key is found, exit the loop.
		}

		return nil // Return nil to indicate successful transaction.
	})

	if err != nil {
		return false, fmt.Errorf("checkAndRetrieveMRC721 error1: %w", err)
	}
	if !found {
		return false, nil
	}

	// Rest of the function remains unchanged, retrieving Mrc721GenesisData based on the found MRC721 name.

	// Extract MRC721 name from the HookInscription
	mrc721name := ""
	mrc721name, _, err = ConvertToNameID(*hookInscription.ContentByte)
	if err != nil {
		mrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
		if err != nil {
			return false, fmt.Errorf("checkAndRetrieveMRC721 error2: %w", err)
		} else {
			mrc721name = mrc721.Miner.GetUpperName()
		}
	}

	// Retrieve Mrc721GenesisData
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("mrc721::geninsc::" + mrc721name))
		if err != nil {
			return fmt.Errorf("checkAndRetrieveMRC721 error3: %w", err)
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, genesisData)
		})
	})
	if err != nil {
		return false, fmt.Errorf("checkAndRetrieveMRC721 error4: %w", err)
	}

	return true, nil
}

// checkAndRetrieveMRC20 checks and retrieves MRC20 data if the inscription is of MRC20 type.
func (b *BTOrdIdx) checkAndRetrieveMRC20(inscriptionID string, hookInscription *HookInscription) (*MRC20Protocol, error) {
	found := false
	var mrc20p *MRC20Protocol

	fmt.Println("checkAndRetrieveMRC20 =", inscriptionID)

	// Create a transaction to read data from the badger database.
	err := b.db.View(func(txn *badger.Txn) error {
		// Define the prefix to search for.
		prefix := []byte("mrc20::inscr_addr::" + inscriptionID)

		// Create an iterator over the transaction.
		// For the iterator, we set the prefix as the seek value.
		// We use the WithPrefix option to limit the iteration to keys with the given prefix.
		opts := badger.DefaultIteratorOptions
		opts.Prefix = prefix
		it := txn.NewIterator(opts)
		defer it.Close()

		// Iterate through the keys.
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			found = true
			break
			// item := it.Item()
			// k := item.Key()

			// // If a key is found, set found to true.
			// if string(k) == string(prefix) {
			// 	found = true
			// 	break
			// }
		}

		return nil // Ignore errors as missing key is part of the logic
	})

	fmt.Println("checkAndRetrieveMRC20 bbb=", inscriptionID)

	// Handle any error that occurred during the transaction.
	if err != nil {
		return nil, fmt.Errorf("checkAndRetrieveMRC20 error1: %w", err)
	}

	// If a matching key was found, parse the MRC20 protocol data.
	if found {
		mrc20p, err = ParseMRC20Protocol(*hookInscription.ContentByte)
		if err != nil {
			return nil, fmt.Errorf("checkAndRetrieveMRC20 error2: %w", err)
		}
	}

	// Return the MRC20 protocol data or nil if not found.
	return mrc20p, nil
}

// GetLotteryList retrieves a list of lottery data for a given MRC-721 name.
// It fetches the genesis inscription data to determine the total number of prize rounds
// and then retrieves the lottery data for each round.
func (b *BTOrdIdx) GetLotteryList(mrc721name string) ([]LotteryData, error) {
	b.rwLock.RLock()
	defer b.rwLock.RUnlock()

	var lotteryList []LotteryData

	// Retrieve Mrc721GenesisData
	var genesisData Mrc721GenesisData
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("mrc721::geninsc::" + mrc721name))
		if err != nil {
			return fmt.Errorf("GetLotteryList error1: %w", err)
		}
		return item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &genesisData)
		})
	})
	if err != nil {
		return nil, fmt.Errorf("GetLotteryList error2: %w", err)
	}

	// Retrieve lottery data for each round
	for i := genesisData.TotalPrizeRound - 1; i > 0; i-- {
		var lotteryData LotteryData
		err := b.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(fmt.Sprintf("lottery::mrc721::%s::%d", mrc721name, i)))
			if err != nil {
				return fmt.Errorf("GetLotteryList error3: %w", err)
			}
			return item.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &lotteryData)
			})
		})
		if err != nil {
			// Handle the error according to your needs.
			// For instance, you can ignore not found errors if some rounds might not have lottery data.
			// return nil, fmt.Errorf("GetLotteryList error4: %w", err)
			continue
		}
		lotteryList = append(lotteryList, lotteryData)
	}

	return lotteryList, nil
}

// UnlockedFindMrc721ImgID searches for the image ID of a given MRC-721 name by scanning through its inscriptions.
// It tries to find a valid image source URL from the inscription's content.
func (b *BTOrdIdx) unlockedFindMrc721ImgID(txn *badger.Txn, mrc721Name string) (string, error) {
	// Loop through inscription numbers from 0 to 100.
	for i := 0; i <= 100; i++ {
		if mrc721Name == "SATMINE" && i < 50 {
			continue
		}
		// Construct the key for the current inscription count.
		key := fmt.Sprintf("mrc721::count_inscr::%s::%d", mrc721Name, i)

		// Retrieve the inscription ID associated with the current key.
		item, err := txn.Get([]byte(key))
		if err != nil {
			// If the key does not exist, continue to the next iteration.
			if err == badger.ErrKeyNotFound {
				continue
			}
			// Return any other error immediately.
			return "", err
		}

		var inscriptionID string
		err = item.Value(func(val []byte) error {
			// Decode the value (inscription ID) as a string.
			inscriptionID = string(val)
			return nil
		})
		if err != nil {
			return "", err
		}

		// Construct the key to retrieve the HookInscription object.
		inscriptionKey := fmt.Sprintf("inscr::%s", inscriptionID)
		inscriptionItem, err := txn.Get([]byte(inscriptionKey))
		if err != nil {
			return "", err
		}

		var hookInscription HookInscription
		err = inscriptionItem.Value(func(val []byte) error {
			// Unmarshal the value into a HookInscription object.
			return jsoniter.Unmarshal(val, &hookInscription)
		})
		if err != nil {
			return "", err
		}

		// If ContentByte is not nil, try to parse the HTML to find an img src.
		if hookInscription.ContentByte != nil {
			imgSrc, err := HtmlToImgSrc(*hookInscription.ContentByte)
			if err == nil {
				// If an img src is found, return it.
				return imgSrc, nil
			} else {
				imgSvgSrc, svgErr := SvgToImgSrc(*hookInscription.ContentByte)
				if svgErr == nil {
					return imgSvgSrc, nil
				}
			}
			// If an error occurs (e.g., img src not found), continue to the next inscription.
		}
	}

	// If the loop completes without finding a valid img src, return an error.
	return "", fmt.Errorf("no valid img src found for MRC-721 name: %s", mrc721Name)
}
