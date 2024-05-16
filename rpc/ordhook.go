package rpc

import (
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime/debug"
	"satmine/satmine"
	"satmine/store"
	"strings"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// Define structs to match the JSON structure
type OrdHookEvent struct {
	Apply     []OrdHookBlock `json:"apply"`
	Chainhook OrdHookChain   `json:"chainhook"`
	Rollback  []interface{}  `json:"rollback"` // Assuming rollback has a similar structure
}

type OrdHookBlock struct {
	BlockIdentifier       OrdHookBlockIdentifier `json:"block_identifier"`
	Metadata              map[string]interface{} `json:"metadata"`
	ParentBlockIdentifier OrdHookBlockIdentifier `json:"parent_block_identifier"`
	Timestamp             int64                  `json:"timestamp"`
	Transactions          []OrdHookTransaction   `json:"transactions"`
}

type OrdHookBlockIdentifier struct {
	Hash  string `json:"hash"`
	Index int    `json:"index"`
}

type OrdHookTransaction struct {
	Metadata              OrdHookTransactionMetadata `json:"metadata"`
	Operations            []interface{}              `json:"operations"`
	TransactionIdentifier OrdHookBlockIdentifier     `json:"transaction_identifier"`
}

type OrdHookTransactionMetadata struct {
	OrdinalOperations []OrdHookOrdinalOperation `json:"ordinal_operations"`
	Proof             interface{}               `json:"proof"`
}

type OrdHookOrdinalOperation struct {
	InscriptionTransferred *OrdHookInscriptionTransferred `json:"inscription_transferred,omitempty"`
	InscriptionRevealed    *OrdHookInscriptionRevealed    `json:"inscription_revealed,omitempty"`
}

type OrdHookInscriptionTransferred struct {
	Destination             OrdHookDestination `json:"destination"`
	InscriptionID           string             `json:"inscription_id"`
	PostTransferOutputValue int                `json:"post_transfer_output_value"`
	SatpointPostTransfer    string             `json:"satpoint_post_transfer"`
	SatpointPreTransfer     string             `json:"satpoint_pre_transfer"`
	TxIndex                 int                `json:"tx_index"`
}

type OrdHookInscriptionRevealed struct {
	ContentBytes            string            `json:"content_bytes"`
	ContentLength           int               `json:"content_length"`
	ContentType             string            `json:"content_type"`
	CurseType               *string           `json:"curse_type"` // Assuming curse_type can be null, hence using a pointer
	InscriberAddress        string            `json:"inscriber_address"`
	InscriptionFee          int               `json:"inscription_fee"`
	InscriptionID           string            `json:"inscription_id"`
	InscriptionInputIndex   int               `json:"inscription_input_index"`
	InscriptionNumber       InscriptionNumber `json:"inscription_number"`
	InscriptionOutputValue  int               `json:"inscription_output_value"`
	OrdinalBlockHeight      int               `json:"ordinal_block_height"`
	OrdinalNumber           int64             `json:"ordinal_number"`
	OrdinalOffset           int64             `json:"ordinal_offset"`
	SatpointPostInscription string            `json:"satpoint_post_inscription"`
	TransfersPreInscription int               `json:"transfers_pre_inscription"`
	TxIndex                 int               `json:"tx_index"`
}

type OrdHookDestination struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type OrdHookChain struct {
	IsStreamingBlocks bool             `json:"is_streaming_blocks"`
	Predicate         OrdHookPredicate `json:"predicate"`
	UUID              string           `json:"uuid"`
}

type OrdHookPredicate struct {
	Operation string `json:"operation"`
	Scope     string `json:"scope"`
}

// New struct to handle the two possible formats of inscription_number
type InscriptionNumber struct {
	Classic int `json:"classic"`
	Jubilee int `json:"jubilee"`
}

// Custom UnmarshalJSON method to handle the different inscription_number formats
func (in *InscriptionNumber) UnmarshalJSON(data []byte) error {
	// First, try to parse the data as an integer
	var intNumber int
	if err := jsoniter.Unmarshal(data, &intNumber); err == nil {
		in.Classic = intNumber
		in.Jubilee = -1 // If it's a single number, set Jubilee to -1
		return nil
	}

	// If it's not an integer, try to parse it as an InscriptionNumber struct
	var structNumber struct {
		Classic int `json:"classic"`
		Jubilee int `json:"jubilee"`
	}
	if err := jsoniter.Unmarshal(data, &structNumber); err != nil {
		return err
	}

	in.Classic = structNumber.Classic
	in.Jubilee = structNumber.Jubilee

	return nil
}

// @Summary Process OrdHook events
// @Schemes
// @Description Parses and processes events related to OrdHook.
// @Tags OrdHook
// @Accept json
// @Produce json
// @Param event body OrdHookEvent true "Event Payload"
// @Success 200 {object} map[string]interface{} "A message confirming successful processing"
// @Failure 400 {object} map[string]interface{} "Error message in case of failure to process the event"
// @Router /mrc20/hookevents [post]
func ordHookEvents(c *gin.Context) {
	//fmt.Println("ordHookEvents()1")

	defer func() {
		if r := recover(); r != nil {
			// Handle the panic, convert it to error if needed
			fmt.Println("ordHookEvents err= ", r)
			fmt.Printf("Stack Trace:\n%s\n", debug.Stack())
			os.Exit(1)
		}
	}()

	// Check if the request comes from localhost (127.0.0.1)
	if !isRequestFromLocalhost(c.Request.RemoteAddr) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied. This endpoint is only accessible from localhost."})
		return
	}

	hookBlock := satmine.HookBlock{}
	hookBlock.Inscriptions = make([]satmine.HookInscription, 0)
	hookBlock.Transfers = make([]satmine.HookTransfer, 0)
	// c.Request.Body
	//fmt.Printf("ordHookEvents Body: %+v\n", c.Request.Body)

	// Use io to read the request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to read request body"})
		return
	}

	// Convert the body to a string and print it
	//bodyString := string(body)
	//fmt.Printf("ordHookEvents Body: %s\n", bodyString)

	//fmt.Println("ordHookEvents()2")

	// Unmarshal the JSON data into the OrdHookEvent struct
	var event OrdHookEvent
	err = jsoniter.Unmarshal(body, &event)
	if err != nil {
		//fmt.Println("jsoniter.Unmarshal(body, &event) err = ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to unmarshal JSON"})
		return
	}
	//fmt.Println("ordHookEvents()3")

	hookBlock.BlockHeight = fmt.Sprintf("%d", event.Apply[0].BlockIdentifier.Index)
	hookBlock.BlockHash = event.Apply[0].BlockIdentifier.Hash
	hookBlock.Timestamp = event.Apply[0].Timestamp

	//fmt.Println("ordHookEvents()3")

	// Loop through the "apply" array in the event
	for _, block := range event.Apply {
		if len(block.Transactions) > 0 {
			for i, transaction := range block.Transactions {
				//fmt.Println("transaction009 =", transaction)
				//fmt.Println("transaction010 =", transaction.Metadata.OrdinalOperations)
				if transaction.Metadata.OrdinalOperations != nil {
					//fmt.Println("transaction020 =", transaction.Metadata.OrdinalOperations)
					for j, op := range transaction.Metadata.OrdinalOperations {
						//fmt.Println("transaction030 =", transaction.Metadata.OrdinalOperations)
						if op.InscriptionTransferred != nil {
							// fmt.Println("op.InscriptionTransferred =", op.InscriptionTransferred)
							// //Log inscription transferred
							fmt.Printf("inscription transfer: %d %d %d %s -> %s %s\n",
								block.BlockIdentifier.Index, i, j,
								op.InscriptionTransferred.InscriptionID,
								op.InscriptionTransferred.Destination.Type,
								op.InscriptionTransferred.Destination.Value)

							tr := satmine.HookTransfer{}
							tr.ID = op.InscriptionTransferred.InscriptionID
							tr.Type = op.InscriptionTransferred.Destination.Type
							tr.ToAddress = op.InscriptionTransferred.Destination.Value
							tr.PostTransferOutputValue = op.InscriptionTransferred.PostTransferOutputValue
							tr.SatpointPostTransfer = op.InscriptionTransferred.SatpointPostTransfer
							tr.SatpointPreTransfer = op.InscriptionTransferred.SatpointPreTransfer
							tr.TxIndex = op.InscriptionTransferred.TxIndex

							hookBlock.Transfers = append(hookBlock.Transfers, tr)

						} else if op.InscriptionRevealed != nil {
							// Log inscription revealed
							// fmt.Printf("establish: %d %d %d %s -mint-> %s\n",
							// 	block.BlockIdentifier.Index, i, j,
							// 	op.InscriptionRevealed.InscriptionID,
							// 	op.InscriptionRevealed.InscriberAddress)

							ins := satmine.HookInscription{}
							ins.ID = op.InscriptionRevealed.InscriptionID
							ins.Number = op.InscriptionRevealed.InscriptionNumber.Classic
							ins.Address = op.InscriptionRevealed.InscriberAddress
							ins.Offset = fmt.Sprintf("%d", op.InscriptionRevealed.OrdinalOffset)
							ins.Sat = int64(op.InscriptionRevealed.OrdinalNumber)
							ins.OrdinalHeight = op.InscriptionRevealed.OrdinalBlockHeight
							ins.BlockHeight = event.Apply[0].BlockIdentifier.Index

							// contentBytes, err := hex.DecodeString(op.InscriptionRevealed.ContentBytes)
							// if err != nil {
							// 	// Handle error (e.g., log it, return a response, etc.)
							// 	fmt.Printf("%s", op.InscriptionRevealed.ContentBytes)
							// 	fmt.Printf("Error decoding hex string: %s\n", err)
							// 	c.JSON(http.StatusBadRequest, gin.H{"error": "op.InscriptionRevealed.ContentBytes to []byte"})
							// 	return
							// }
							// ins.ContentByte = &contentBytes

							// Check if ContentBytes starts with "0x" and remove it
							hexString := op.InscriptionRevealed.ContentBytes
							if len(hexString) >= 2 && hexString[:2] == "0x" {
								hexString = hexString[2:]
							}
							// Convert hexString from hex string to []byte
							contentBytes, err := hex.DecodeString(hexString)
							if err != nil {
								// Handle error (e.g., log it, return a response, etc.)
								c.JSON(http.StatusBadRequest, gin.H{"error": "op.InscriptionRevealed.ContentBytes to []byte"})
								return
							}
							ins.ContentByte = &contentBytes

							ins.ContentType = op.InscriptionRevealed.ContentType
							ins.ContentLength = op.InscriptionRevealed.ContentLength
							ins.CurseType = op.InscriptionRevealed.CurseType
							ins.InscriptionFee = op.InscriptionRevealed.InscriptionFee
							ins.InscriptionInputIndex = op.InscriptionRevealed.InscriptionInputIndex
							ins.InscriptionOutputValue = op.InscriptionRevealed.InscriptionOutputValue
							ins.SatpointPostInscription = op.InscriptionRevealed.SatpointPostInscription
							ins.TransfersPreInscription = op.InscriptionRevealed.TransfersPreInscription
							ins.TxIndex = op.InscriptionRevealed.TxIndex

							hookBlock.Inscriptions = append(hookBlock.Inscriptions, ins)

						} else {
							fmt.Printf("find not inscription_transferred inscription_revealed : %d %d %d\n",
								block.BlockIdentifier.Index, i, j)
						}
					}
				} else {
					fmt.Printf("find not ordinal_operations ts: %d %d\n",
						block.BlockIdentifier.Index, i)
				}
			}
		} else {
			fmt.Printf("Discovery of non-traded blocks: %d\n", block.BlockIdentifier.Index)
		}
	}

	//fmt.Println("ordHookEvents()4")

	//c.JSON(http.StatusOK, gin.H{"message": "Event processed successfully"})

	//fmt.Printf("hookBlock: %+v\n", hookBlock.BlockHeight)

	store := store.Instance()
	err = store.OrdIdx.WriteBlock(&hookBlock)
	if err != nil {

		fmt.Printf("hookBlock.BlockHeight is already stored  %s\n", hookBlock.BlockHeight)

		// c.JSON(http.StatusOK, gin.H{"message": "Event processed successfully"})
		//c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		// os.Exit(1)

		//return
	}

	// Print the parsed data
	//fmt.Printf("Parsed Event Data: %+v\n", event)
	c.JSON(http.StatusOK, gin.H{"message": "Event processed successfully"})
}

// isRequestFromLocalhost checks if the given address is from localhost (127.0.0.1).
// It extracts the IP part from addresses like "127.0.0.1:12345" and compares it to "127.0.0.1".
func isRequestFromLocalhost(addr string) bool {
	ip := strings.Split(addr, ":")[0] // Extract the IP part from the address
	return ip == "127.0.0.1"
}
