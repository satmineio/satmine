package rpc

import (
	"encoding/json"
	"net/http"
	"satmine/store"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PostRecord godoc
// @Summary Write a record to the database
// @Schemes
// @Description Writes a new record with an incrementing index to the database
// @Tags record
// @Accept json
// @Produce json
// @Param address query string true "Address identifier"
// @Param rectype query string true "Record type identifier"
// @Param message body string true "Message content to be stored"
// @Success 200 {object} map[string]interface{} "Success message with code and message text"
// @Failure 400 {object} map[string]interface{} "Error message with code and message text"
// @Router /mrc20/postrecord [post]
func PostRecord(c *gin.Context) {
	address := c.Query("address")
	rectype := c.Query("rectype")
	message, _ := c.GetRawData() // Get message body as raw data

	store := store.Instance() // Get the global instance of the store

	err := store.RecIdx.WriteRecorder(address, rectype, string(message))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	//store.AllSocketPush(string(message)) // Push the message to all connected WebSocket clients
	// Construct the JSON message
	jsonMessage, err := json.Marshal(gin.H{
		"address": address,
		"rectype": rectype,
		"message": string(message),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create JSON message"})
		return
	}
	store.AllSocketPush(string(jsonMessage)) // Push the JSON message to all connected WebSocket clients

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Record written successfully"})
}

// GetRecords godoc
// @Summary Retrieve records from the database
// @Schemes
// @Description Retrieves a list of records based on address, record type, page number, page size, and order
// @Tags record
// @Accept json
// @Produce json
// @Param address query string true "Address identifier"
// @Param rectype query string true "Record type identifier"
// @Param pageNum query int true "Page number for pagination"
// @Param pageSize query int true "Number of records per page"
// @Param ascend query bool false "Order of record retrieval (true for ascending, false for descending)"
// @Success 200 {object} []string "List of records"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/getrecords [get]
func GetRecords(c *gin.Context) {
	address := c.Query("address")
	rectype := c.Query("rectype")
	pageNum, _ := strconv.Atoi(c.Query("pageNum"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))
	ascend, _ := strconv.ParseBool(c.Query("ascend"))

	store := store.Instance()

	records, err := store.RecIdx.ReadeRecorder(address, rectype, pageNum, pageSize, ascend)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}
