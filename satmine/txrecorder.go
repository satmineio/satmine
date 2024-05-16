// filePath: satmine/txrecorder.go

package satmine

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/badger/v4"
)

type BTRecIdx struct {
	db     *badger.DB
	rwLock sync.RWMutex
}

// NewBTRecIdx initializes a new instance of NewBTRecIdx with a given Manager.
func NewBTRecIdx(db *badger.DB) *BTRecIdx {
	return &BTRecIdx{
		db: db,
	}
}

// WriteRecorder writes a message to the database with an incrementing index.
// The key format is "rec::address::rectype::index" and the index increments for each new entry.
func (b *BTRecIdx) WriteRecorder(address, rectype, msg string) (err error) {
	b.rwLock.Lock()
	defer b.rwLock.Unlock()

	// Key for storing the current index
	indexKey := fmt.Sprintf("num::%s::%s", address, rectype)

	// Retrieve the current index from the database
	var index int
	err = b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(indexKey))
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if err == badger.ErrKeyNotFound {
			index = 0
		} else {
			err = item.Value(func(val []byte) error {
				index = int(val[0]) // Assuming the index is stored as a single byte (not safe for large indices)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Construct the key for the new message
	recordKey := fmt.Sprintf("rec::%s::%s::%d", address, rectype, index)

	//fmt.Println("recordKey =", recordKey)

	// Increment the index for future operations
	newIndex := index + 1
	newIndexBytes := []byte{byte(newIndex)} // Assuming the index is stored as a single byte

	// Start a write transaction
	err = b.db.Update(func(txn *badger.Txn) error {
		// Write the new message
		if err := txn.Set([]byte(recordKey), []byte(msg)); err != nil {
			return err
		}

		// Update the index key with the new index
		if err := txn.Set([]byte(indexKey), newIndexBytes); err != nil {
			return err
		}
		return nil
	})
	return err
}

// ReadeRecorderResponse defines the structure of the response for the ReadeRecorder method.
type ReadeRecorderResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TotalCount int      `json:"totalCount"`
		Address    string   `json:"address"`
		RecType    string   `json:"rectype"`
		PageNum    int      `json:"pageNum"`
		PageSize   int      `json:"pageSize"`
		Records    []string `json:"records"`
	} `json:"data"`
}

// ReadeRecorder reads records from the database based on pagination and order, and returns them in a structured format.
// This function takes an address and record type to identify the data scope, and it uses pagination along with sorting order.
func (b *BTRecIdx) ReadeRecorder(address, rectype string, pageNum, pageSize int, ascend bool) (ReadeRecorderResponse, error) {
	b.rwLock.RLock()
	defer b.rwLock.RUnlock()

	response := ReadeRecorderResponse{
		Code:    200,
		Message: "",
	}

	response.Data.Address = address
	response.Data.RecType = rectype
	response.Data.PageNum = pageNum
	response.Data.PageSize = pageSize

	// Key for retrieving the total number of records
	indexKey := fmt.Sprintf("num::%s::%s", address, rectype)

	// Read the total number of records from the database
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(indexKey))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				response.Data.TotalCount = 0 // No records exist if the index key is not found
			}
			return err
		}

		return item.Value(func(val []byte) error {
			response.Data.TotalCount = int(val[0]) // Assume that the count is stored as a single byte
			return nil
		})
	})
	if err != nil {
		response.Code = 500
		response.Message = err.Error()
		return response, err
	}

	// Calculate the start and end index for the page depending on the order
	var startIndex, endIndex int
	if ascend {
		startIndex = pageNum * pageSize
		if startIndex >= response.Data.TotalCount || response.Data.TotalCount == 0 {
			response.Message = "No records found or request out of range"
			return response, fmt.Errorf(response.Message)
		}
		endIndex = startIndex + pageSize
		if endIndex > response.Data.TotalCount {
			endIndex = response.Data.TotalCount
		}
	} else { // Descending order
		endIndex = response.Data.TotalCount - pageNum*pageSize
		if endIndex <= 0 {
			response.Message = "No records found or request out of range"
			return response, fmt.Errorf(response.Message)
		}
		startIndex = endIndex - pageSize
		if startIndex < 0 {
			startIndex = 0
		}
	}

	var indices []int
	if ascend {
		for i := startIndex; i < endIndex; i++ {
			indices = append(indices, i)
		}
	} else {
		for i := endIndex - 1; i >= startIndex; i-- {
			indices = append(indices, i)
		}
	}

	response.Data.Records = make([]string, 0, len(indices))

	// Retrieve the records within the specified range
	err = b.db.View(func(txn *badger.Txn) error {
		for _, i := range indices {
			recordKey := fmt.Sprintf("rec::%s::%s::%d", address, rectype, i)
			item, err := txn.Get([]byte(recordKey))
			if err != nil {
				return err
			}
			err = item.Value(func(val []byte) error {
				response.Data.Records = append(response.Data.Records, string(val))
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		response.Code = 500
		response.Message = err.Error()
		return response, err
	}

	return response, nil
}
