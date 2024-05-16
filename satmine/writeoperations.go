package satmine

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger/v4"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

// Write the list of newly inscribed inscriptions into the KV database.
func (b *BTOrdIdx) addInscriptionList(txn *badger.Txn, block *HookBlock) (err error) {
	// Iterate over the inscriptions in the block
	for _, inscription := range block.Inscriptions {

		// inscr::inscription_id -> HookInscription{}
		inscriptionKey := fmt.Sprintf("inscr::%s", inscription.ID)
		// Check if the inscription key already exists in the database
		_, err := txn.Get([]byte(inscriptionKey))
		if err == nil {
			// If the key already exists, log an error and return
			logger.Error("Inscription already exists in database", zap.String("key", inscriptionKey))
			return fmt.Errorf("inscription %s already exists in database", inscriptionKey)
		} else if err != badger.ErrKeyNotFound {
			// If there's an error other than key not found, log it and return
			logger.Error("Error checking for inscription existence: ", zap.Error(err))
			return err
		}
		// Serialize the inscription to JSON using jsoniter
		inscriptionJSON, err := jsoniter.Marshal(inscription)
		if err != nil {
			logger.Error("Failed to marshal inscription: ", zap.Error(err))
			return err
		}
		// Write the serialized inscription to the database
		if err := txn.Set([]byte(inscriptionKey), inscriptionJSON); err != nil {
			logger.Error("Failed to write inscription to database: ", zap.Error(err))
			return err
		}

		// inscr::number::[inscription.Number] -> [inscription.ID]
		inscriptionNumberKey := fmt.Sprintf("inscr::number::%d", inscription.Number)
		if err := txn.Set([]byte(inscriptionNumberKey), []byte(inscription.ID)); err != nil {
			logger.Error("Failed to write inscription number to database: ", zap.Error(err))
			return err
		}

		// Check if ContentByte is not nil
		if inscription.ContentByte != nil {
			// Use ValidateProtocolData to check the data type (mrc-721 or mrc-20)
			isValid, protocolType, err := ValidateProtocolData(*inscription.ContentByte)
			//fmt.Println("---ValidateProtocolData protocolType=", isValid, protocolType, err)

			if err != nil {
				//logger.Error("Failed to validate protocol data: ", zap.Error(err))
				continue // Skip to next inscription on error
			}

			if isValid {
				// Log the identified protocol type
				//logger.Info("Valid protocol data identified: ", zap.String("protocolType", protocolType))

				// Depending on the protocol type, call the respective parsing function
				switch protocolType {
				case "mrc-721":
					mrc721Data, err := ParseMRC721Protocol(*inscription.ContentByte)
					if err != nil {
						logger.Info("Failed to parse MRC-721 data: ", zap.Error(err))
					} else {
						logger.Info("Parsed MRC-721 Data: ", zap.Reflect("mrc721Data", mrc721Data))
						err := b.writeMrc721(txn, block, &inscription, mrc721Data)
						if err != nil {
							logger.Info("Failed to write MRC-721 data: ", zap.Error(err))
							return err
						}
					}
				case "mrc-721html":
					mrc721Data, err := ParseMRC721HtmlProtocol(txn, *inscription.ContentByte)
					if err != nil {
						logger.Info("Failed to parse 721html data: ", zap.Error(err))
					} else {
						//logger.Info("Parsed MRC-721html Data: ", zap.Reflect("mrc721Data", mrc721Data))
						err := b.writeMrc721(txn, block, &inscription, mrc721Data)
						if err != nil {
							logger.Info("Failed to write MRC-721html data: ", zap.Error(err))
							return err
						}
					}
				case "mrc-721svg":
					mrc721Data, err := ParseMRC721SvgProtocol(txn, *inscription.ContentByte)
					if err != nil {
						logger.Info("Failed to parse 721svg data: ", zap.Error(err))
					} else {
						//logger.Info("Parsed MRC-721svg Data: ", zap.Reflect("mrc721Data", mrc721Data))
						err := b.writeMrc721(txn, block, &inscription, mrc721Data)
						if err != nil {
							logger.Info("Failed to write MRC-721svg data: ", zap.Error(err))
							return err
						}
					}

				case "mrc-20":
					mrc20Data, err := ParseMRC20Protocol(*inscription.ContentByte)
					//fmt.Println("ParseMRC20Protocol =", block.BlockHeight, mrc20Data)
					if err != nil {
						logger.Info("Failed to parse MRC-20 data: ", zap.Error(err))
					} else {
						logger.Info("Parsed MRC-20 Data: ", zap.Reflect("mrc20Data", mrc20Data))
						err := b.writeMrc20(txn, block, &inscription, mrc20Data)
						if err != nil {
							logger.Info("Failed to write MRC-20 data: ", zap.Error(err))
							return err
						}
					}
				default:
					logger.Info("Unknown protocol type")
				}
			} else {
				logger.Info("Invalid protocol data")
			}
		}
	}

	return nil
}

// Writing MRC-721 inscriptions, can only be used within Badger's Update operation.
func (b *BTOrdIdx) writeMrc721(txn *badger.Txn, block *HookBlock, inscr *HookInscription, mrc721Data *MRC721Protocol) (err error) {
	// Construct the key using the inscription ID

	// mrc721::geninsc::[inscription_name]  -> inscription_id
	geninsc_key := fmt.Sprintf("mrc721::geninsc::%s", mrc721Data.Miner.GetUpperName())
	geninsc20_key := fmt.Sprintf("mrc20::geninsc::%s", mrc721Data.Token.GetLowerTick())
	// Search for the key in the transaction
	item, err := txn.Get([]byte(geninsc_key))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Check if the MRC20 genesis inscription key exists
			mrc20GenInscKey := fmt.Sprintf("mrc20::geninsc::%s", mrc721Data.Token.GetLowerTick())
			_, mrc20Err := txn.Get([]byte(mrc20GenInscKey))

			// If the MRC20 genesis inscription key exists, log the information and return nil
			if mrc20Err == nil {
				logger.Info("MRC20 genesis inscription key already exists", zap.String("key", mrc20GenInscKey))
				return nil
			}

			// Key not found, create a new Mrc721GenesisData instance
			geninscMax, err := strconv.Atoi(mrc721Data.Miner.Max)
			if err != nil {
				logger.Error("Failed to strconv.Atoi(mrc721Data.Miner.Max) ", zap.Error(err))
				return err
			}

			mrc721GenesisInscription := Mrc721GenesisData{
				ID:                   inscr.ID,
				Number:               inscr.Number,
				Name:                 mrc721Data.Miner.GetUpperName(),
				PrevName:             mrc721Data.Miner.Name,
				BlockHeight:          block.BlockHeight,
				GenesisAddress:       inscr.Address,
				InscriptionsCount:    1, // Initialize count as 1 for new inscription
				InscriptionsMax:      geninscMax,
				PrizePoolTokens:      "0",
				TotalMinedTokens:     "0",
				TotalPrizePoolTokens: "0",
				Tick:                 mrc721Data.Token.GetLowerTick(),
				PrevTick:             mrc721Data.Token.Tick,
				GenesisBlockHeight:   block.BlockHeight,
				GenesisTimestamp:     block.Timestamp,
				TotalPrizeRound:      0,
				TotalBurn:            "0",
			}

			// Serialize Mrc721GenesisData to JSON
			mrc721GenesisInscriptionJSON, err := jsoniter.Marshal(mrc721GenesisInscription)
			if err != nil {
				logger.Error("Failed to marshal MRC-721 genesis inscription: ", zap.Error(err))
				return err
			}

			// Set the new value in the database
			err = txn.Set([]byte(geninsc_key), mrc721GenesisInscriptionJSON)
			if err != nil {
				logger.Error("Error setting new MRC-721 genesis inscription: ", zap.Error(err))
				return err
			}
			// Set the new value in the database
			err = txn.Set([]byte(geninsc20_key), []byte(mrc721Data.Miner.GetUpperName()))
			if err != nil {
				logger.Error("Error setting new MRC-20 genesis inscription: ", zap.Error(err))
				return err
			}

			err = b.addNewMrc721(txn, block, inscr, mrc721Data, 0)
			if err != nil {
				logger.Error("Error addNewMrc721 MRC-721 genesis inscription: ", zap.Error(err))
				return nil
			}

		} else {
			// An error occurred while searching for the key
			logger.Error("Error searching for MRC-721 inscription: ", zap.Error(err))
			return err
		}
	} else {
		// Key found, update the existing Mrc721GenesisData
		var genesisData Mrc721GenesisData
		err := item.Value(func(val []byte) error {
			return jsoniter.Unmarshal(val, &genesisData)
		})
		if err != nil {
			logger.Error("Failed to unmarshal existing MRC-721 genesis inscription: ", zap.Error(err))
			return err
		}

		// Read the existing HookInscription using the key 'inscr::genesisData.ID'
		existingInscrKey := fmt.Sprintf("inscr::%s", genesisData.ID)
		var existingHookInscription HookInscription
		item, err = txn.Get([]byte(existingInscrKey))
		if err != nil {
			if err != badger.ErrKeyNotFound {
				// If there's an error other than key not found, log it and return
				logger.Error("Error reading existing HookInscription: ", zap.Error(err))
				return err
			}
			// If the key is not found, proceed without incrementing the count
			logger.Info("Existing HookInscription not found, proceeding without incrementing count")
			return err
		} else {
			// Unmarshal the existing HookInscription
			err = item.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &existingHookInscription)
			})
			if err != nil {
				logger.Error("Failed to unmarshal existing HookInscription: ", zap.Error(err))
				return err
			}

			// Compare the content byte using IsEqual721Data function
			//if !IsEqual721Data(existingHookInscription.ContentByte, inscr.ContentByte) {
			if !IsEqual721Data(&existingHookInscription, inscr) {
				// If the data is not equal, log a message and proceed
				logger.Info("Existing and current HookInscription data are not identical, proceeding")
				logger.Info("existingHookInscription.ID=" + existingHookInscription.ID + " inscr.ID=" + inscr.ID)
			} else {

				// Retrieve HookInscription using the key 'inscr::genesisData.ID'
				inscrKey := fmt.Sprintf("inscr::%s", genesisData.ID)
				hookInscrItem, err := txn.Get([]byte(inscrKey))
				if err != nil {
					logger.Error("Error retrieving HookInscription: ", zap.Error(err))
					return err
				}

				// Decode the HookInscription
				var hookInscription HookInscription
				err = hookInscrItem.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
				if err != nil {
					logger.Error("Failed to unmarshal HookInscription: ", zap.Error(err))
					return err
				}

				// Parse the MRC721 protocol data from ContentByte
				firstMrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
				if err != nil {
					logger.Error("Failed to parse MRC721 protocol: ", zap.Error(err))
					return err
				}
				// Compare firstMrc721.Miner.Max with genesisData.InscriptionsCount
				maxInscriptions, err := strconv.Atoi(firstMrc721.Miner.Max)
				if err != nil {
					logger.Error("Failed to convert firstMrc721.Miner.Max to integer: ", zap.Error(err))
					return err
				}
				// Check if the maximum number of inscriptions specified by the miner is less than the current inscriptions count
				if genesisData.InscriptionsCount >= maxInscriptions {
					logger.Error("Max inscriptions limit exceeded", zap.String("firstMrc721.Miner.Name", firstMrc721.Miner.Name))
					return nil
				}

				// this add code..

				err = b.addNewMrc721(txn, block, inscr, mrc721Data, genesisData.InscriptionsCount)
				if err != nil {
					logger.Error("Error addNewMrc721 MRC-721 genesis inscription: ", zap.Error(err))
					return nil
				}

				genesisData.EndID = inscr.ID
				genesisData.EndBlockHeight = block.BlockHeight
				genesisData.EndTimestamp = block.Timestamp

				// Increment the inscriptions count if data is identical
				genesisData.InscriptionsCount++

				// Serialize the updated Mrc721GenesisData
				updatedInscriptionJSON, err := jsoniter.Marshal(genesisData)
				if err != nil {
					logger.Error("Failed to marshal updated MRC-721 genesis inscription: ", zap.Error(err))
					return err
				}

				// Update the value in the database
				err = txn.Set([]byte(geninsc_key), updatedInscriptionJSON)
				if err != nil {
					logger.Error("Error updating MRC-721 genesis inscription: ", zap.Error(err))
					return err
				}

			}
		}

	}

	return nil
}

// After all validations are successful, write the newly added inscription into the KV (Key-Value) database.
func (b *BTOrdIdx) addNewMrc721(txn *badger.Txn, block *HookBlock, inscr *HookInscription, mrc721Data *MRC721Protocol, mrc721Count int) (err error) {

	// Define the key for the address inscription count
	addrNumKey := fmt.Sprintf("mrc721::addr_num::%s::%s", mrc721Data.Miner.GetUpperName(), inscr.Address)
	// Retrieve the inscription count for the address
	item, err := txn.Get([]byte(addrNumKey))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			// Key not found, create a new entry with count 1
			err = txn.Set([]byte(addrNumKey), []byte("1"))
			if err != nil {
				logger.Error("Error setting new inscription count: ", zap.Error(err))
				return err
			}
		} else {
			// An error occurred while searching for the key
			logger.Error("Error searching for address inscription count: ", zap.Error(err))
			return err
		}
	} else {
		// Key found, increment the inscription count if below the limit
		var count int
		err := item.Value(func(val []byte) error {
			count, err = strconv.Atoi(string(val))
			return err
		})
		if err != nil {
			logger.Error("Failed to parse inscription count: ", zap.Error(err))
			return err
		}

		// Compare the current count with the maximum allowed inscriptions
		maxInscriptions, err := strconv.Atoi(mrc721Data.Miner.Lim)

		if err != nil {
			logger.Error("Failed to convert Max to integer: ", zap.Error(err))
			return err
		}

		if count < maxInscriptions {
			// Increment the count and update the database
			count++
			err = txn.Set([]byte(addrNumKey), []byte(strconv.Itoa(count)))
			if err != nil {
				logger.Error("Error updating inscription count: ", zap.Error(err))
				return err
			}
		} else {
			// Limit reached, do not increment
			return errors.New("inscription limit reached")
		}
	}

	// Construct keys for the new inscriptions
	// mrc721::name_inscr::[mrc721_name]::[inscription_id] -> nil
	// This key maps the MRC-721 name to the inscription ID, used for storing a particular name with a lot of inscriptions underneath it.
	keyNameAddr := fmt.Sprintf("mrc721::name_inscr::%s::%s", mrc721Data.Miner.GetUpperName(), inscr.ID)
	if err := txn.Set([]byte(keyNameAddr), nil); err != nil {
		logger.Error("Failed to write name to inscription mapping: ", zap.Error(err))
		return err
	}

	// mrc721::addr_inscr::[user_addr]::[inscription_id] -> nil
	// This key maps the user address to the inscription ID, used for storing all inscriptions owned by a user.
	keyAddrInscr := fmt.Sprintf("mrc721::addr_inscr::%s::%s", inscr.Address, inscr.ID)
	if err := txn.Set([]byte(keyAddrInscr), nil); err != nil {
		logger.Error("Failed to write address to inscription mapping: ", zap.Error(err))
		return err
	}

	// mrc721::inscr_addr::[inscription_id]::[user_addr] -> nil
	// This key maps the inscription ID to the user address.
	keyInscrAddr := fmt.Sprintf("mrc721::inscr_addr::%s::%s", inscr.ID, inscr.Address)
	if err := txn.Set([]byte(keyInscrAddr), nil); err != nil {
		logger.Error("Failed to write inscription to address mapping: ", zap.Error(err))
		return err
	}

	// New key for mapping MRC-721 series count to inscription ID
	// Format: mrc721::count_inscr::[mrc721_name]::[mrc721_count] -> inscription_id
	keyCountInscr := fmt.Sprintf("mrc721::count_inscr::%s::%d", mrc721Data.Miner.GetUpperName(), mrc721Count)
	if err := txn.Set([]byte(keyCountInscr), []byte(inscr.ID)); err != nil {
		logger.Error("Failed to write series count to inscription mapping: ", zap.Error(err))
		return err
	}

	// Format: mrc721::count_inscr::[mrc721_name]::[inscription_id] -> mrc721_count
	keyInscrCount := fmt.Sprintf("mrc721::inscr_count::%s::%s", mrc721Data.Miner.GetUpperName(), inscr.ID)
	mrc721CountStr := strconv.Itoa(mrc721Count)
	if err := txn.Set([]byte(keyInscrCount), []byte(mrc721CountStr)); err != nil {
		logger.Error("Failed to write  mrc721Count", zap.Error(err))
		return err
	}

	return nil
}

// addTransferList processes and updates key-value pairs for each transfer in a block.
func (b *BTOrdIdx) addTransferList(txn *badger.Txn, block *HookBlock) (err error) {
	//fmt.Println("addTransferList0=", block.Transfers)

	for _, transferItem := range block.Transfers {
		// Log the transfer from and to addresses
		//fmt.Println("addTransferList1", transferItem.ID, transferItem.ToAddress)

		toAddress := ""
		if transferItem.Type == "transferred" {
			toAddress = transferItem.ToAddress
		} else if transferItem.Type == "burnt" {
			// Default processing to black hole address after combustion
			toAddress = "1BitcoinEaterAddressDontSendf59kuE"
		} else {
			// Consider removing the error when formalizing your use here
			logger.Info(fmt.Sprintf("unknown transfer type %s %s", transferItem.Type, transferItem.ToAddress))
			continue
			//return errors.New("unknown transfer type")
		}

		//fmt.Println("addTransferList2", transferItem.ID, toAddress)

		// --- transfer 721 ---
		{
			// Operation 1: Update mrc721::inscr_addr::[inscription_id]::[user_addr]
			prefix := []byte(fmt.Sprintf("mrc721::inscr_addr::%s::", transferItem.ID))
			opts := badger.DefaultIteratorOptions
			opts.Prefix = prefix
			it := txn.NewIterator(opts)
			defer it.Close()

			var oldAddr, oldKey string
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				key := item.Key()
				oldKey = string(key)
				oldAddr = strings.TrimPrefix(oldKey, string(prefix))
				break // assuming only one key matches the prefix
			}

			if oldAddr == "" {
				//old address not found"
				//logger.Info(fmt.Sprintf("Key not found for transfer ID: %s from address: %s", transferItem.ID, toAddress))
			} else {
				// Delete old key-value pair
				err = txn.Delete([]byte(oldKey))
				if err != nil {
					return fmt.Errorf("error deleting old key: %w", err)
				}
				//fmt.Println("addTransferList transferItem.ID=", transferItem.ID)

				// Add new key-value pair
				newKey := fmt.Sprintf("mrc721::inscr_addr::%s::%s", transferItem.ID, toAddress)
				err = txn.Set([]byte(newKey), nil)
				if err != nil {
					return fmt.Errorf("error setting new key: %w", err)
				}

				// Operation 2: Update mrc721::addr_inscr::[user_addr]::[inscription_id]
				oldKeyValue := []byte(fmt.Sprintf("mrc721::addr_inscr::%s::%s", oldAddr, transferItem.ID))
				_, err = txn.Get(oldKeyValue)
				if err == badger.ErrKeyNotFound {
					return errors.New("key-value pair does not exist")
				}
				if err != nil {
					return fmt.Errorf("error checking old key-value pair: %w", err)
				}

				// Delete old key-value pair
				err = txn.Delete(oldKeyValue)
				if err != nil {
					return fmt.Errorf("error deleting old key-value pair: %w", err)
				}

				// Add new key-value pair
				newKeyValue := fmt.Sprintf("mrc721::addr_inscr::%s::%s", toAddress, transferItem.ID)
				err = txn.Set([]byte(newKeyValue), nil)
				if err != nil {
					return fmt.Errorf("error setting new key-value pair: %w", err)
				}

				// Retrieve the HookInscription associated with the current transfer item.
				inscrKey := fmt.Sprintf("inscr::%s", transferItem.ID)
				item, err := txn.Get([]byte(inscrKey))
				if err != nil {
					return fmt.Errorf("error retrieving HookInscription: %w", err)
				}

				var hookInscription HookInscription
				err = item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &hookInscription)
				})
				if err != nil {
					return fmt.Errorf("error unmarshalling HookInscription: %w", err)
				}

				// Update the address in the HookInscription with the new toAddress.
				hookInscription.Address = toAddress

				// Marshal the updated HookInscription and write it back to the database.
				updatedInscrBytes, err := jsoniter.Marshal(hookInscription)
				if err != nil {
					return fmt.Errorf("error marshalling updated HookInscription: %w", err)
				}
				err = txn.Set([]byte(inscrKey), updatedInscrBytes)
				if err != nil {
					return fmt.Errorf("error writing updated HookInscription back to database: %w", err)
				}

			}

		}

		// --- transfer 20 ---
		{

			// Operation 1: Update mrc20::inscr_addr::[inscription_id]::[user_addr]
			prefix := []byte(fmt.Sprintf("mrc20::inscr_addr::%s::", transferItem.ID))
			opts := badger.DefaultIteratorOptions
			opts.Prefix = prefix
			it := txn.NewIterator(opts)
			defer it.Close()

			var oldAddr, oldKey string
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				key := item.Key()
				oldKey = string(key)
				oldAddr = strings.TrimPrefix(oldKey, string(prefix))
				break // assuming only one key matches the prefix
			}

			if oldAddr == "" {
				// Old address not found, log error but do not return it
				//logger.Error(fmt.Sprintf("Old address not found for transfer ID: %s", transferItem.ID))
			} else {
				fmt.Println("addTransferList mrc20 transferItem.ID=", transferItem.ID)

				// Retrieve HookInscription for the MRC-20 transfer
				inscrKey := fmt.Sprintf("inscr::%s", transferItem.ID)
				item, err := txn.Get([]byte(inscrKey))
				if err != nil {
					return err
				}

				var mrc20Inscription HookInscription
				err = item.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &mrc20Inscription)
				})
				if err != nil {
					return err
				}

				// Parse MRC-20 protocol data
				mrc20Data, err := ParseMRC20Protocol(*mrc20Inscription.ContentByte)
				if err != nil {
					return err
				}

				// Handle balance update for the receiving address
				balanceKey := fmt.Sprintf("mrc20::balance::%s::%s", toAddress, mrc20Data.Tick)
				var currentBalance *big.Int
				item, err = txn.Get([]byte(balanceKey))
				if err != nil {
					if err == badger.ErrKeyNotFound {
						currentBalance = big.NewInt(0) // Initialize balance if not exist
					} else {
						return err
					}
				} else {
					err = item.Value(func(val []byte) error {
						currentBalance = new(big.Int)
						return currentBalance.UnmarshalText(val)
					})
					if err != nil {
						return err
					}
				}
				transferAmount := new(big.Int)
				transferAmount, ok := transferAmount.SetString(mrc20Data.Amt, 10)
				if !ok {
					return fmt.Errorf("invalid amount format: %s", mrc20Data.Amt)
				}
				newBalance := new(big.Int).Add(currentBalance, transferAmount)
				//fmt.Println("addTransferList newBalance=", newBalance)
				//fmt.Println("addTransferList balanceKey=", balanceKey)
				balanceBytes := newBalance.Text(10)
				err = txn.Set([]byte(balanceKey), []byte(balanceBytes))
				if err != nil {
					return err
				}

				// Delete the following keys
				keysToDelete := []string{
					fmt.Sprintf("mrc20::name_inscr::%s::%s", mrc20Data.Tick, transferItem.ID),
					fmt.Sprintf("mrc20::addr_inscr::%s::%s", oldAddr, transferItem.ID),
					oldKey,
				}
				for _, key := range keysToDelete {
					err := txn.Delete([]byte(key))
					if err != nil {
						return fmt.Errorf("error deleting key %s: %v", key, err)
					}
				}
			}
		}
	}
	return nil
}

// writeMrc20 handles the MRC-20 token inscription writing.
func (b *BTOrdIdx) writeMrc20(txn *badger.Txn, block *HookBlock, inscr *HookInscription, mrc20Data *MRC20Protocol) (err error) {
	//fmt.Println("writeMrc20 mrc20Data=", mrc20Data)

	// Check if the operation is 'transfer'
	if mrc20Data.Op == "transfer" {
		// Check if the token exists
		mrc721nameKey := "mrc20::geninsc::" + mrc20Data.Tick
		_, err := txn.Get([]byte(mrc721nameKey))
		if err != nil {
			fmt.Println("Token does not exist")
			return nil // Token does not exist, no error, stop execution
		}
		// var mrc721name string
		// err = item.Value(func(val []byte) error {
		// 	mrc721name = string(val)
		// 	return nil
		// })
		// if err != nil {
		// 	return err
		// }

		// Retrieve the balance for the address and convert it to a big.Int
		balanceKey := "mrc20::balance::" + inscr.Address + "::" + mrc20Data.Tick

		//fmt.Println("writeMrc20 balanceKey=", balanceKey+"|")
		item, err := txn.Get([]byte(balanceKey))
		if err != nil {
			fmt.Println("Error retrieving balance")
			return nil
		}
		//var balanceBigInt *big.Int
		balanceBigInt := new(big.Int)
		err = item.Value(func(val []byte) error {
			//fmt.Println("writeMrc20 string(val)=", string(val))
			//balanceBigInt, _ = new(big.Int).SetString(string(val), 10)
			balanceBigInt.SetString(string(val), 10)

			return nil
		})
		if err != nil {
			return err
		}

		// Convert mrc20Data.Amt to big.Int and compare with the balance
		amountBigInt := new(big.Int)
		amountBigInt.SetString(mrc20Data.Amt, 10) // Assuming mrc20Data.Amt is a base 10 string
		if amountBigInt.Cmp(balanceBigInt) > 0 {
			//fmt.Println("Insufficient balance for the transaction")
			return nil // Balance is less than amount, no error, stop execution
		}

		//fmt.Println("writeMrc20 inscr=", inscr.ID)
		//fmt.Println("writeMrc20 amountBigInt=", amountBigInt.String())
		//fmt.Println("writeMrc20 balanceBigInt=", balanceBigInt.String())
		// Update the balance and write back to the database
		newBalanceBigInt := new(big.Int).Sub(balanceBigInt, amountBigInt)
		//err = txn.Set([]byte(balanceKey), newBalanceBigInt.Bytes())
		err = txn.Set([]byte(balanceKey), []byte(newBalanceBigInt.String()))
		if err != nil {
			return err
		}

		//fmt.Println("writeMrc20 newBalanceBigInt=", newBalanceBigInt.String())

		// Write the additional key-value pairs as required
		err = txn.Set([]byte("mrc20::name_inscr::"+mrc20Data.Tick+"::"+inscr.ID), nil)
		if err != nil {
			return err
		}
		err = txn.Set([]byte("mrc20::addr_inscr::"+inscr.Address+"::"+inscr.ID), nil)
		if err != nil {
			return err
		}
		err = txn.Set([]byte("mrc20::inscr_addr::"+inscr.ID+"::"+inscr.Address), nil)
		if err != nil {
			return err
		}

		// Serialize the HookInscription instance to JSON.
		// This is done to convert the complex structure into a format suitable for storage.
		inscrBytes, err := jsoniter.Marshal(inscr)
		if err != nil {
			// If there is an error during serialization, log the error and return.
			zap.L().Error("Failed to serialize HookInscription", zap.Error(err))
			return err
		}

		// Store the serialized HookInscription in the database with the key formed by prefixing 'inscr::' to the Inscription ID.
		// This allows for easy retrieval of HookInscription by its ID.
		err = txn.Set([]byte("inscr::"+inscr.ID), inscrBytes)
		if err != nil {
			// If there is an error while setting the value in the database, log the error and return.
			zap.L().Error("Failed to store serialized HookInscription", zap.Error(err))
			return err
		}

	} else if mrc20Data.Op == "burn" {
		//fmt.Println("burn", mrc20Data)
		if mrc20Data.Insc == nil {
			logger.Info("mrc20Data.Insc is nil, inscr.ID=" + inscr.ID)
			return nil
		}

		// Retrieve the balance for the address and convert it to a big.Int
		balanceKey := "mrc20::balance::" + inscr.Address + "::" + mrc20Data.Tick
		item, err := txn.Get([]byte(balanceKey))
		if err != nil {
			fmt.Println("Error retrieving balance")
			return err
		}
		var balanceBigInt *big.Int
		err = item.Value(func(val []byte) error {
			balanceBigInt = new(big.Int).SetBytes(val)
			return nil
		})
		if err != nil {
			return err
		}

		// Convert mrc20Data.Amt to big.Int and compare with the balance
		amountBigInt := new(big.Int)
		amountBigInt.SetString(mrc20Data.Amt, 10) // Assuming mrc20Data.Amt is a base 10 string
		if amountBigInt.Cmp(balanceBigInt) > 0 {
			fmt.Println("Insufficient balance for the burn")
			return nil // Balance is less than amount, no error, stop execution
		}

		// Update the balance and write back to the database
		newBalanceBigInt := new(big.Int).Sub(balanceBigInt, amountBigInt)
		err = txn.Set([]byte(balanceKey), newBalanceBigInt.Bytes())
		if err != nil {
			return err
		}

		// Convert mrc20Data.Insc to a string for the key
		burnKey := fmt.Sprintf("mrc721::burn::%s", *mrc20Data.Insc)
		item, err = txn.Get([]byte(burnKey))
		var totalBurnt *big.Int

		// Check if the burn record exists
		if err == badger.ErrKeyNotFound {
			// If not found, initialize totalBurnt to zero
			totalBurnt = big.NewInt(0)
		} else if err != nil {
			// On other errors, return the error
			return err
		} else {
			// Read the existing burn value
			err = item.Value(func(val []byte) error {
				totalBurnt = new(big.Int).SetBytes(val)
				return nil
			})
			if err != nil {
				return err
			}
		}

		// Add amountBigInt to totalBurnt and write back
		totalBurnt.Add(totalBurnt, amountBigInt)
		err = txn.Set([]byte(burnKey), totalBurnt.Bytes())
		if err != nil {
			return err
		}

		// Retrieve the name associated with the MRC20 token using its ticker.
		to721Item, to721Eerr := txn.Get([]byte("mrc20::geninsc::" + mrc20Data.Tick))
		if to721Eerr != nil {
			// Handle error if the key does not exist or any other error occurs
			fmt.Println("Failed to retrieve MRC721 name:", to721Eerr)
			return to721Eerr
		}

		var mrc721Name string
		// Extract the value from the item and assign it to mrc721Name
		err = to721Item.Value(func(val []byte) error {
			mrc721Name = string(val)
			return nil
		})
		if err != nil {
			// Handle error if unable to extract the value
			fmt.Println("Failed to extract MRC721 name:", err)
			return err
		}

		// Print the name associated with the MRC20 token.
		//fmt.Println("TotalBurn test code MRC721 Name:", mrc721Name)

		geninsc_key := fmt.Sprintf("mrc721::geninsc::%s", mrc721Name)
		genItem, genErr := txn.Get([]byte(geninsc_key))
		if genErr != nil {
			// Key found, update the existing Mrc721GenesisData
			var genesisData Mrc721GenesisData
			err := genItem.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &genesisData)
			})
			if err != nil {
				logger.Error("Failed to unmarshal existing MRC-721 genesis inscription: ", zap.Error(err))
				return err
			}

			// this add code
			// Convert the TotalBurn value from string to *big.Int for manipulation
			totalBurnBigInt := new(big.Int)
			_, ok := totalBurnBigInt.SetString(genesisData.TotalBurn, 10)
			if !ok {
				// Handle error if the string conversion to big.Int fails
				logger.Error("Failed to convert TotalBurn to big.Int", zap.String("TotalBurn", genesisData.TotalBurn))
				return errors.New("failed to convert TotalBurn to big.Int")
			}

			// Add the amount from amountBigInt to the totalBurnBigInt
			totalBurnBigInt = totalBurnBigInt.Add(totalBurnBigInt, amountBigInt)

			// Convert the updated total burn value back to a string
			genesisData.TotalBurn = totalBurnBigInt.String()

			// Serialize the updated Mrc721GenesisData
			updatedInscriptionJSON, err := jsoniter.Marshal(genesisData)
			if err != nil {
				logger.Error("Failed to marshal updated MRC-721 genesis inscription: ", zap.Error(err))
				return err
			}

			// Update the value in the database
			err = txn.Set([]byte(geninsc_key), updatedInscriptionJSON)
			if err != nil {
				logger.Error("Error updating MRC-721 genesis inscription: ", zap.Error(err))
				return err
			}

		}

	}

	return nil
}

// Mining using MRC721 inscriptions.
func (b *BTOrdIdx) mineWithMrc721Inscription(txn *badger.Txn, block *HookBlock) (err error) {
	// Define the prefix for MRC-721 genesis inscriptions
	prefix := []byte("mrc721::geninsc::")

	// Use the Badger iterator to iterate over all keys with the specified prefix
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	it := txn.NewIterator(opts)
	defer it.Close()

	// Iterate over keys
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := item.Key()

		// Extract the mrc721 name from the key
		mrc721Name := string(key[len(prefix):])

		// Retrieve the value (Mrc721GenesisData)
		err := item.Value(func(val []byte) error {
			var genesisData Mrc721GenesisData
			if err := jsoniter.Unmarshal(val, &genesisData); err != nil {
				logger.Error("Failed to unmarshal MRC-721 genesis genesisData: ", zap.Error(err))
				return err
			}

			// Retrieve HookInscription using the key 'inscr::genesisData.ID'
			inscrKey := fmt.Sprintf("inscr::%s", genesisData.ID)
			hookInscrItem, err := txn.Get([]byte(inscrKey))
			if err != nil {
				logger.Error("Error retrieving HookInscription: ", zap.Error(err))
				return err
			}

			// Decode the HookInscription
			var hookInscription HookInscription
			err = hookInscrItem.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &hookInscription)
			})
			if err != nil {
				logger.Error("Failed to unmarshal HookInscription: ", zap.Error(err))
				return err
			}

			// Parse the MRC721 protocol data from ContentByte
			firstMrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
			if err != nil {
				logger.Error("Failed to parse MRC721 protocol: ", zap.Error(err))
				return err
			}

			// Print the mrc721 name and the genesisData details
			//logger.Info(fmt.Sprintf("MRC-721 Name: %s, Inscription: %+v", mrc721Name, genesisData))
			// Print the genesis MRC-721 data
			//logger.Info(fmt.Sprintf("Genesis MRC-721 Data: %+v", firstMrc721))

			if genesisData.TotalMinedTokens == firstMrc721.Token.Total {
				logger.Info(fmt.Sprintf("MRC-721 Name: %s,  miner is end. ", mrc721Name))
				return nil
			}

			// ------------

			// Create an instance of Mrc721MinerMap
			minerMap := Mrc721MinerMap{Data: make(map[string]*Mrc721MinerData)}
			// Perform prefix search for mrc721::name_inscr::mrc721Name
			nameInscrPrefix := []byte(fmt.Sprintf("mrc721::name_inscr::%s::", mrc721Name))
			it2 := txn.NewIterator(opts)
			for it2.Seek(nameInscrPrefix); it2.ValidForPrefix(nameInscrPrefix); it2.Next() {
				item := it2.Item()
				inscriptionID := string(item.Key())[len(nameInscrPrefix):]

				// Retrieve HookInscription using inscriptionID
				inscrKey := fmt.Sprintf("inscr::%s", inscriptionID)
				hookInscrItem, err := txn.Get([]byte(inscrKey))
				if err != nil {
					logger.Error("Error retrieving HookInscription: ", zap.Error(err))
					continue
				}

				// Decode the HookInscription
				var minerInscription HookInscription
				err = hookInscrItem.Value(func(val []byte) error {
					return jsoniter.Unmarshal(val, &minerInscription)
				})
				if err != nil {
					logger.Error("Failed to unmarshal HookInscription: ", zap.Error(err))
					continue
				}

				// Retrieve the burn number from mrc721::burn::[inscription_id]
				burnKey := fmt.Sprintf("mrc721::burn::%s", inscriptionID)
				burnItem, err := txn.Get([]byte(burnKey))
				var burnNum string = "0" // Default value if the key is not found or is empty
				if err == nil {
					err = burnItem.Value(func(val []byte) error {
						if len(val) > 0 {
							burnNum = string(val)
						}
						return nil
					})
					if err != nil {
						return err
					}
				}

				// Populate minerMap
				minerMap.Data[inscriptionID] = &Mrc721MinerData{
					InscriptionsID:     inscriptionID,
					InscriptionsNumber: minerInscription.Number,
					Address:            minerInscription.Address,
					BurnNum:            burnNum,
					Tick:               genesisData.Tick,
					MinedAmount:        "0", // To be calculated and filled later
					Power:              *big.NewInt(1000),
				}
			}
			it2.Close()
			// // Print minerMap in JSON format
			// minerMapJSON, err := jsoniter.Marshal(minerMap)
			// if err != nil {
			// 	logger.Error("Failed to marshal minerMap: ", zap.Error(err))
			// 	return err
			// }
			// logger.Info(fmt.Sprintf("Mrc721MinerMap: %s", string(minerMapJSON)))

			// mining
			calcResult, err := CalculateMiningRewards(block.BlockHeight, &genesisData, firstMrc721, &minerMap)
			if err != nil {
				logger.Error("Failed to calculate mining rewards: ", zap.Error(err))
				return err
			}
			//fmt.Printf("calcResult %v", calcResult)

			if !calcResult.IsMiningEnd {

				// Iterate over the minerMap to update user balances
				for _, minerData := range minerMap.Data {

					// Retrieve the existing mined amount for the miner
					minerKey := fmt.Sprintf("mrc721::inscr_miner::%s", minerData.InscriptionsID)
					var existingMinedAmountBigInt *big.Int
					item, err := txn.Get([]byte(minerKey))
					if err == nil {
						err = item.Value(func(val []byte) error {
							existingMinedAmountBigInt = new(big.Int)
							existingMinedAmountBigInt.SetBytes(val)
							return nil
						})
						if err != nil {
							logger.Error("Error retrieving existing mined amount: ", zap.Error(err))
							return err
						}
					} else {
						// If no existing mined amount, initialize to zero
						existingMinedAmountBigInt = big.NewInt(0)
					}

					// Convert minerData.MinedAmount to big.Int and add it to the existing mined amount
					minedAmountBigInt := new(big.Int)
					_, ok := minedAmountBigInt.SetString(minerData.MinedAmount, 10)
					if !ok {
						logger.Error("Failed to parse mined amount to big.Int")
						return fmt.Errorf("failed to parse mined amount to big.Int")
					}
					updatedMinedAmountBigInt := new(big.Int).Add(existingMinedAmountBigInt, minedAmountBigInt)

					// Write the updated mined amount to the database
					err = txn.Set([]byte(minerKey), updatedMinedAmountBigInt.Bytes())
					if err != nil {
						logger.Error("Failed to update mined amount: ", zap.Error(err))
						return err
					}

					// minerData.Power is already a big.Int (assumed), so we use it directly
					powerBigInt := minerData.Power

					powerKey := fmt.Sprintf("mrc721::inscr_power::%s", minerData.InscriptionsID)

					// Write the power data to the database
					err = txn.Set([]byte(powerKey), powerBigInt.Bytes())
					if err != nil {
						logger.Error("Failed to set power data: ", zap.Error(err))
						return err
					}

					// -----
					balanceKey := fmt.Sprintf("mrc20::balance::%s::%s", minerData.Address, minerData.Tick)

					// Retrieve the current balance
					var currentBalance *big.Int
					item, err = txn.Get([]byte(balanceKey))
					if err == nil {
						err = item.Value(func(val []byte) error {
							currentBalance = new(big.Int)
							currentBalance.SetString(string(val), 10)
							return nil
						})
						if err != nil {
							logger.Error("Error retrieving current balance: ", zap.Error(err))
							return err
						}
					} else {
						// If no current balance, initialize to zero
						currentBalance = big.NewInt(0)
					}

					// Parse the mined amount
					minedAmount := new(big.Int)
					_, ok = minedAmount.SetString(minerData.MinedAmount, 10)
					if !ok {
						logger.Error("Failed to parse mined amount")
						return fmt.Errorf("failed to parse mined amount")
					}

					// Add the mined amount to the current balance
					newBalance := new(big.Int).Add(currentBalance, minedAmount)

					// Serialize the new balance
					newBalanceBytes := []byte(newBalance.String())

					// Write the updated balance to the database
					err = txn.Set([]byte(balanceKey), newBalanceBytes)
					if err != nil {
						logger.Error("Failed to update balance: ", zap.Error(err))
						return err
					}
				}

				// Convert genesisData values to big.Int using new(big.Int).SetString()
				prizePoolTokensBigInt, _ := new(big.Int).SetString(genesisData.PrizePoolTokens, 10)
				totalMinedTokensBigInt, _ := new(big.Int).SetString(genesisData.TotalMinedTokens, 10)
				totalPrizePoolTokensBigInt, _ := new(big.Int).SetString(genesisData.TotalPrizePoolTokens, 10)

				// Convert calcResult values to big.Int
				currentPrizePoolAllNumBigInt, _ := new(big.Int).SetString(calcResult.CurrentPrizePoolAllNum, 10)
				currentMiningAllNumBigInt, _ := new(big.Int).SetString(calcResult.CurrentMiningAllNum, 10)

				// Perform the addition operations
				prizePoolTokensBigInt.Add(prizePoolTokensBigInt, currentPrizePoolAllNumBigInt)
				totalMinedTokensBigInt.Add(totalMinedTokensBigInt, currentMiningAllNumBigInt)
				totalPrizePoolTokensBigInt.Add(totalPrizePoolTokensBigInt, currentPrizePoolAllNumBigInt)

				// Convert the big.Int results back to strings
				genesisData.PrizePoolTokens = prizePoolTokensBigInt.String()
				genesisData.TotalMinedTokens = totalMinedTokensBigInt.String()
				genesisData.TotalPrizePoolTokens = totalPrizePoolTokensBigInt.String()

				// Serialize the updated genesisData to JSON
				updatedGenesisDataJSON, err := jsoniter.Marshal(genesisData)
				if err != nil {
					logger.Error("Failed to marshal updated genesisData: ", zap.Error(err))
					return err
				}

				// Write the updated genesisData to the database
				genesisDataKey := fmt.Sprintf("mrc721::geninsc::%s", mrc721Name)
				err = txn.Set([]byte(genesisDataKey), updatedGenesisDataJSON)
				if err != nil {
					logger.Error("Failed to update genesisData in the database: ", zap.Error(err))
					return err
				}

			}

			return nil
		})

		if err != nil {
			logger.Error("Error processing item: ", zap.Error(err))
			return err
		}
	}

	return nil
}

// lotteryWithBlockHash performs a lottery draw using the block hash as a seed.
// This function retrieves genesisData and firstMrc721 data and prints them out.
func (b *BTOrdIdx) lotteryWithBlockHash(txn *badger.Txn, block *HookBlock) (err error) {
	if block.BlockHash == NO_INSCRIPTION_BLOCK_HASH {
		return nil
	}

	// Define the prefix for MRC-721 genesis inscriptions
	prefix := []byte("mrc721::geninsc::")
	// Use the Badger iterator to iterate over all keys with the specified prefix
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = true
	it := txn.NewIterator(opts)
	defer it.Close()

	// Iterate over keys
	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		item := it.Item()
		key := item.Key()

		// Extract the mrc721 name from the key
		mrc721Name := string(key[len(prefix):])

		// Retrieve the value (Mrc721GenesisData)
		err := item.Value(func(val []byte) error {
			var genesisData Mrc721GenesisData
			if err := jsoniter.Unmarshal(val, &genesisData); err != nil {
				logger.Error("Failed to unmarshal MRC-721 genesis genesisData: ", zap.Error(err))
				return err
			}

			if genesisData.PrizePoolTokens == "0" {
				//logger.Info(fmt.Sprintf("MRC-721 Name: %s,  lottery tokens is 0. ", mrc721Name))
				return nil
			}

			// Retrieve HookInscription using the key 'inscr::genesisData.ID'
			inscrKey := fmt.Sprintf("inscr::%s", genesisData.ID)
			hookInscrItem, err := txn.Get([]byte(inscrKey))
			if err != nil {
				logger.Error("Error retrieving HookInscription: ", zap.Error(err))
				return err
			}

			// Decode the HookInscription
			var hookInscription HookInscription
			err = hookInscrItem.Value(func(val []byte) error {
				return jsoniter.Unmarshal(val, &hookInscription)
			})
			if err != nil {
				logger.Error("Failed to unmarshal HookInscription: ", zap.Error(err))
				return err
			}

			// Parse the MRC721 protocol data from ContentByte
			firstMrc721, err := ParseMRC721Protocol(*hookInscription.ContentByte)
			if err != nil {
				logger.Error("Failed to parse MRC721 protocol: ", zap.Error(err))
				return err
			}

			if firstMrc721.Ltry == nil {
				logger.Info(fmt.Sprintf("MRC-721 Name: %s,  lottery is nil. ", mrc721Name))
				return nil
			}

			genesisHeightBigInt, _ := new(big.Int).SetString(genesisData.BlockHeight, 10)
			currentHeightBigInt, _ := new(big.Int).SetString(block.BlockHeight, 10)
			intvlBigInt, _ := new(big.Int).SetString(firstMrc721.Ltry.Intvl, 10)

			blockCount := new(big.Int).Sub(currentHeightBigInt, genesisHeightBigInt)

			if (blockCount.Cmp(big.NewInt(0)) != 0) && (new(big.Int).Rem(blockCount, intvlBigInt).Cmp(big.NewInt(0)) == 0) {
				//block.BlockHash = ""

				winpBigInt := stringToPercentageBigInt(firstMrc721.Ltry.Winp)
				distBigInt := stringToPercentageBigInt(firstMrc721.Ltry.Dist)
				// Print the retrieved genesisData and firstMrc721

				randomWinp, err := convertHashToBigInt(block.BlockHash, 1000)
				if err != nil {
					logger.Error("Failed to convert hash to big.Int: ", zap.Error(err))
					return err
				}

				// Perform lottery logic if randomWinp is less than or equal to winpBigInt
				if randomWinp.Cmp(winpBigInt) <= 0 {
					// Convert InscriptionsCount to int64 and decrement by 1 for lottery
					inscriptionsCount := int64(genesisData.InscriptionsCount)
					luckNumLimit := inscriptionsCount // - 1

					// Call convertHashToBigInt with the new limit
					luckNum, err := convertHashToBigInt(block.BlockHash, luckNumLimit)
					if err != nil {

						logger.Error("Failed to convert hash to big.Int for luckNum: block.BlockHash="+block.BlockHash, zap.Error(err))
						return err
					}

					// Print the luckNum
					logger.Info("Lottery luck number: ", zap.String("LuckNum", luckNum.String()))

					// Calculate the actual amount of prize money to be distributed
					prizePoolTokensBigInt, _ := new(big.Int).SetString(genesisData.PrizePoolTokens, 10)
					actualPrizeAmount := new(big.Int).Mul(prizePoolTokensBigInt, distBigInt)
					actualPrizeAmount.Div(actualPrizeAmount, big.NewInt(1000))

					// this code add ...
					// Read the inscription ID using the 'mrc721::count_inscr::[mrc721Name]::[luckNum]' key
					luckInscriptionIDKey := fmt.Sprintf("mrc721::count_inscr::%s::%s", mrc721Name, luckNum.String())
					item, err := txn.Get([]byte(luckInscriptionIDKey))
					if err != nil {
						logger.Error("Failed to get luck inscription ID: ", zap.Error(err))
						return err
					}

					var luckInscriptionID string
					err = item.Value(func(val []byte) error {
						luckInscriptionID = string(val)
						return nil
					})
					if err != nil {
						logger.Error("Failed to read luck inscription ID value: ", zap.Error(err))
						return err
					}

					// Find the user address associated with the lucky inscription ID
					prefix := fmt.Sprintf("mrc721::inscr_addr::%s::", luckInscriptionID)
					var luckAddress string
					it := txn.NewIterator(badger.DefaultIteratorOptions)
					for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
						item := it.Item()
						key := item.Key()
						luckAddress = string(key[len(prefix):])
						break
					}
					it.Close()

					if luckAddress == "" {
						logger.Error("No luck address found")
						return errors.New("no luck address found")
					}

					var balanceStr string
					balanceStr = "0"
					// Retrieve the balance of the lucky address for the specific token
					balanceKey := fmt.Sprintf("mrc20::balance::%s::%s", luckAddress, firstMrc721.Token.Tick)
					balanceItem, err := txn.Get([]byte(balanceKey))
					if err != nil {
						logger.Info("Failed to get balance: ", zap.Error(err))
						//return err
					} else {
						err = balanceItem.Value(func(val []byte) error {
							balanceStr = string(val)
							return nil
						})
						if err != nil {
							logger.Error("Failed to read balance value: ", zap.Error(err))
							return err
						}
					}

					// Convert the balance to big.Int
					balanceBigInt, ok := new(big.Int).SetString(balanceStr, 10)
					if !ok {
						logger.Error("Invalid balance format")
						return errors.New("invalid balance format")
					}

					// Add the prize amount to the balance
					updatedBalance := new(big.Int).Add(balanceBigInt, actualPrizeAmount)

					// Write the updated balance back to the KV store
					if err := txn.Set([]byte(balanceKey), []byte(updatedBalance.String())); err != nil {
						logger.Error("Failed to write updated balance back to KV store: ", zap.Error(err))
						return err
					}

					// Subtract the actual prize amount from the prize pool
					updatedPrizePool := new(big.Int).Sub(prizePoolTokensBigInt, actualPrizeAmount)

					oldPrizePoolTokens := genesisData.PrizePoolTokens
					// Update the genesisData with the new prize pool amount
					genesisData.PrizePoolTokens = updatedPrizePool.String()
					genesisData.TotalPrizeRound += 1

					// Serialize the updated genesisData to JSON using jsoniter
					updatedGenesisDataJSON, err := jsoniter.Marshal(genesisData)
					if err != nil {
						logger.Error("Failed to marshal updated genesisData: ", zap.Error(err))
						return err
					}

					// Write the updated genesisData back to the KV store
					if err := txn.Set([]byte(fmt.Sprintf("mrc721::geninsc::%s", mrc721Name)), updatedGenesisDataJSON); err != nil {
						logger.Error("Failed to write updated genesisData back to KV store: ", zap.Error(err))
						return err
					}

					// Create a new LotteryData structure to record the details of the lottery win
					lotteryData := LotteryData{
						BlockHeight:    block.BlockHeight,
						BlockHash:      block.BlockHash,
						BlockTimestamp: block.Timestamp,
						Address:        luckAddress,
						InscriptionID:  luckInscriptionID,
						Number:         0,
						Mrc721name:     mrc721Name,
						WinAmount:      actualPrizeAmount.String(),
						JackpotAccum:   oldPrizePoolTokens,          // Accumulated jackpot before the win
						Round:          genesisData.TotalPrizeRound, // Current round of the lottery
						Winp:           firstMrc721.Ltry.Winp,
						Dist:           firstMrc721.Ltry.Dist,
					}

					// Serialize the LotteryData to JSON
					lotteryDataJSON, err := jsoniter.Marshal(lotteryData)
					if err != nil {
						logger.Error("Failed to marshal lotteryData: ", zap.Error(err))
						return err
					}

					// Generate the key for the new lottery win entry
					lotteryKey := fmt.Sprintf("lottery::mrc721::%s::%d", mrc721Name, genesisData.TotalPrizeRound)

					// Write the LotteryData to the KV store
					if err := txn.Set([]byte(lotteryKey), lotteryDataJSON); err != nil {
						logger.Error("Failed to write lotteryData to KV store: ", zap.Error(err))
						return err
					}

				}

				// logger.Info("lotteryWithBlockHash Genesis Data: ", zap.Reflect("GenesisData", genesisData))
				// logger.Info("lotteryWithBlockHash First MRC-721 Data: ", zap.String("mrc721Name", mrc721Name), zap.Reflect("FirstMRC721", firstMrc721))

			}

			return nil
		})

		if err != nil {
			logger.Error("Error processing item: ", zap.Error(err))
			return err
		}
	}

	return nil
}
