package rpc

import (
	"fmt"
	"io"
	"net/http"
	"satmine/satmine"
	"satmine/store"
	"strconv"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

// GetLatestBlock retrieves the most recent block number from the database and returns it in JSON format.
// @Summary Retrieve the latest block number
// @Schemes
// @Description Retrieves the most recent block number from the blockchain
// @Tags mrc20
// @Accept json
// @Produce json
// @Success 200 {object} string "The latest block number"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/latestblock [get]
func GetLatestBlock(c *gin.Context) {
	// Use the glo package's Instance() method to get the OrdIdx object
	store := store.Instance()

	// Call the GetLastBlock method to get the latest block number
	blockNumber, err := store.OrdIdx.GetLastBlock()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 501, "message": err.Error(), "data": -1})
		return
	}

	// Return the block number as a JSON string to the client
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": blockNumber, "message": "success"})
}

// GetBlockByHeight godoc
// @Summary Retrieve a block by its height
// @Schemes
// @Description Retrieves a block from the blockchain by its specified height
// @Tags mrc20
// @Accept json
// @Produce json
// @Param blockHeight query string true "Block Height"
// @Success 200 {object} satmine.HookBlock "Block information"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/blockbyheight [get]
func GetBlockByHeight(c *gin.Context) {
	blockHeight := c.Query("blockHeight")
	store := store.Instance()

	block, err := store.OrdIdx.GetBlockByHeight(blockHeight)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blockJSON, err := jsoniter.Marshal(block)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal block data"})
		return
	}

	c.Data(http.StatusOK, "application/json", blockJSON)
}

// GetBlockByHash godoc
// @Summary Retrieve a block by its hash
// @Schemes
// @Description Retrieves a block from the blockchain by its specified hash
// @Tags mrc20
// @Accept json
// @Produce json
// @Param blockHash query string true "Block Hash"
// @Success 200 {object} satmine.HookBlock "Block information"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/blockbyhash [get]
func GetBlockByHash(c *gin.Context) {
	blockHash := c.Query("blockHash")
	store := store.Instance()

	block, err := store.OrdIdx.GetBlockByHash(blockHash)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blockJSON, err := jsoniter.Marshal(block)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal block data"})
		return
	}

	c.Data(http.StatusOK, "application/json", blockJSON)
}

// GetBlocks godoc
// @Summary Retrieve blocks within a range
// @Schemes
// @Description Retrieves blocks from the blockchain based on a starting block height and an offset
// @Tags mrc20
// @Accept json
// @Produce json
// @Param blockHeight query string true "Starting Block Height"
// @Param offsetHeight query string true "Offset Height"
// @Success 200 {array} satmine.SimpleHookBlock "Array of blocks"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/blocks [get]
func GetBlocks(c *gin.Context) {
	blockHeight := c.Query("blockHeight")
	offsetHeight := c.Query("offsetHeight")
	store := store.Instance()

	blocks, err := store.OrdIdx.GetBlocks(blockHeight, offsetHeight)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	blocksJSON, err := jsoniter.Marshal(blocks)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal block data"})
		return
	}

	c.Data(http.StatusOK, "application/json", blocksJSON)
}

// GetAddressInfo godoc
// @Summary Retrieve inscription IDs for both MRC721 and MRC20 tokens
// @Schemes
// @Description Retrieves inscription IDs for both MRC721 and MRC20 tokens associated with a given address
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "User Address"
// @Success 200 {object} map[string][]string "Inscription IDs for MRC721 and MRC20 tokens"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressinfo [get]
func GetAddressInfo(c *gin.Context) {
	address := c.Query("address") // Get the user address from the query parameters
	store := store.Instance()

	// Call the GetAddressInfo method of OrdIdx to retrieve inscription IDs
	mrc721s, mrc20s, err := store.OrdIdx.GetAddressInfo(address)
	if err != nil {
		// Return an error message if the retrieval fails
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure that nil slices are represented as empty slices in the response
	if mrc721s == nil {
		mrc721s = []string{}
	}
	if mrc20s == nil {
		mrc20s = []string{}
	}

	// Create a response map containing the inscription IDs
	response := map[string][]string{
		"mrc721": mrc721s,
		"mrc20":  mrc20s,
	}

	// Use jsoniter to marshal the response into JSON format
	responseJSON, err := jsoniter.Marshal(response)
	if err != nil {
		// Return an error message if marshaling fails
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal response data"})
		return
	}

	// Send the JSON response to the client
	c.Data(http.StatusOK, "application/json", responseJSON)
}

// GetAddressBalance godoc
// @Summary Retrieve the balance for a specific address and token
// @Schemes
// @Description Retrieves the balance for a specific address and token (tick)
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address"
// @Param tick query string true "Token ticker"
// @Success 200 {string} string "Balance of the token for the specified address"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressbalance [get]
func GetAddressBalance(c *gin.Context) {
	address := c.Query("address") // Get the address from the query parameters
	tick := c.Query("tick")       // Get the token ticker from the query parameters

	store := store.Instance() // Get the store instance

	// Call the GetAddressBalance method of OrdIdx to retrieve the balance
	balance, err := store.OrdIdx.GetAddressBalance(address, tick)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the balance as a JSON response
	c.JSON(http.StatusOK, balance)
}

// GetAddressBalances godoc
// @Summary Retrieve all token balances for a specific address
// @Schemes
// @Description Retrieves all token balances associated with a given address
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address"
// @Success 200 {object} []map[string]string "Array of balances for each token"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressbalances [get]
func GetAddressBalances(c *gin.Context) {
	address := c.Query("address") // Get the address from the query parameters

	store := store.Instance() // Get the store instance

	// Call the GetAddressBalances method of OrdIdx to retrieve all balances
	balancesJSON, err := store.OrdIdx.GetAddressBalances(address)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return the balances as a JSON response
	c.Data(http.StatusOK, "application/json", []byte(balancesJSON))
}

// GetInscriptionResult defines the structure for the GetInscription API response.
type GetInscriptionResult struct {
	Code    int                      `json:"code"`    // HTTP status code
	Message string                   `json:"message"` // Descriptive message about the result
	Data    *satmine.HookInscription `json:"data"`    // Data payload containing the inscription details, nil if error occurs
}

// GetInscription godoc
// @Summary Retrieve inscription information by ID
// @Schemes
// @Description Retrieves the inscription information for a given inscription ID
// @Tags mrc20
// @Accept json
// @Produce json
// @Param id query string true "Inscription ID"
// @Success 200 {object} GetInscriptionResult "Inscription information wrapped in a result structure"
// @Failure 400 {object} GetInscriptionResult "Error message and code if retrieval fails"
// @Router /mrc20/inscription [get]
func GetInscription(c *gin.Context) {
	id := c.Query("id") // Get the inscription ID from the query parameters

	store := store.Instance() // Get the store instance

	// Call the GetInscription method of OrdIdx to retrieve the inscription information
	inscription, err := store.OrdIdx.GetInscription(id)
	if err != nil {
		// Return an error message if the retrieval fails, using the structured response format
		c.JSON(http.StatusBadRequest, GetInscriptionResult{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
			Data:    nil,
		})
		return
	}

	// Create and send success response using the structured response format
	result := GetInscriptionResult{
		Code:    http.StatusOK,
		Message: "Success",
		Data:    &inscription,
	}
	c.JSON(http.StatusOK, result)
}

// MiningProfitChart handles the request to parse and return MRC721Protocol data.
// @Summary Parse and Return MRC721Protocol Data
// @Schemes
// @Description Parses input JSON into MRC721Protocol format and returns it
// @Tags mrc20
// @Accept json
// @Produce json
// @Param data body string true "MRC721Protocol JSON data" example("{\"p\": \"mrc-721\", \"miner\": {\"name\": \"Demo 721\", \"max\": \"100\", \"lim\":\"5\"}, \"token\": {\"tick\": \"coin\", \"total\": \"2100000000000000\", \"beg\": \"50000000000\", \"halv\": \"10\", \"dcr\": \"0.555\"}, \"ltry\": {\"pool\": \"0.05\", \"intvl\": \"9\", \"winp\": \"0.10\", \"dist\": \"0.60\"}, \"burn\": {\"unit\": \"8000000000\", \"boost\": \"0.05\"}}")
// @Success 200 {object} satmine.MRC721Protocol "Parsed MRC721Protocol data"
// @Failure 400 {string} string "Invalid input"
// @Router /mrc20/miningprofitchart [post]
func MiningProfitChart(g *gin.Context) {
	// Read request body
	body, err := io.ReadAll(g.Request.Body)
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return
	}

	// Parse JSON to MRC721Protocol structure
	var protocol satmine.MRC721Protocol
	if err := jsoniter.Unmarshal(body, &protocol); err != nil {
		g.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing JSON"})
		return
	}

	store := store.Instance() // Get the store instance

	// Fixed step and max values
	const step = 1     // Assume some fixed step value
	const max = 280000 // Assume some fixed max value

	//fmt.Println("MiningProfitChart protocol=", protocol)

	// Call GetMiningProfitChart function
	profitChartJSON, err := store.OrdIdx.GetMiningProfitChart(&protocol, step, max)
	if err != nil {
		g.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("GetMiningProfitChart end....................................\n")

	// Directly return the JSON response from GetMiningProfitChart to the frontend
	g.Data(http.StatusOK, "application/json", []byte(profitChartJSON))

}

// GetInscription godoc
// @Summary Retrieve inscription information by ID
// @Schemes
// @Description Retrieves the inscription information for a given inscription ID
// @Tags mrc20
// @Accept json
// @Produce json
// @Param id query string true "Inscription ID"
// @Success 200 {object} satmine.HookInscription "Inscription information"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/inscriptionplus [get]
func GetInscriptionPlus(c *gin.Context) {
	id := c.Query("id") // Get the inscription ID from the query parameters

	store := store.Instance() // Get the store instance

	// Call the GetInscription method of OrdIdx to retrieve the inscription information
	inscription, err := store.OrdIdx.GetInscriptionPlus(id)
	if err != nil {
		// Return an error message if the retrieval fails
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Use jsoniter to marshal the inscription information into JSON format
	inscriptionJSON, err := jsoniter.Marshal(inscription)
	if err != nil {
		// Return an error message if marshaling fails
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal inscription data"})
		return
	}

	// Send the JSON response to the client
	c.Data(http.StatusOK, "application/json", inscriptionJSON)
}

// GetAllMrc721 godoc
// @Summary Retrieve all MRC-721 genesis inscriptions
// @Schemes
// @Description Retrieves all MRC-721 genesis inscriptions and sorts them by number
// @Tags mrc20
// @Accept json
// @Produce json
// @Success 200 {array} satmine.Mrc721GenesisDataWeb "List of MRC-721 genesis inscriptions"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/allmrc721 [get]
func GetAllMrc721(c *gin.Context) {
	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAllMrc721 method on the OrdIdx object of the store
	genesisDataList, err := store.OrdIdx.GetAllMrc721()
	if err != nil {
		// If an error occurs, send a 400 Bad Request response with the error message
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Serialize the list of MRC-721 genesis inscription data to JSON
	genesisDataJSON, err := jsoniter.Marshal(genesisDataList)
	if err != nil {
		// If serialization fails, send a 400 Bad Request response with the error message
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal genesis data"})
		return
	}

	// Send the serialized JSON data to the client with a 200 OK response
	c.Data(http.StatusOK, "application/json", genesisDataJSON)
}

// GetOneMrc721 godoc
// @Summary Retrieve a single MRC-721 genesis inscription
// @Schemes
// @Description Retrieves a single MRC-721 genesis inscription based on the provided name
// @Tags mrc20
// @Accept json
// @Produce json
// @Param mrc721Name query string true "MRC-721 Name"
// @Success 200 {object} satmine.Mrc721GenesisDataWeb "MRC-721 genesis inscription"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/onemrc721 [get]
func GetOneMrc721(c *gin.Context) {
	// Retrieve the MRC-721 name from the query string
	mrc721Name := c.Query("mrc721Name")
	if mrc721Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "MRC-721 name is required"})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetOneMrc721 method on the OrdIdx object of the store
	genesisData, err := store.OrdIdx.GetOneMrc721(mrc721Name)
	if err != nil {
		// If an error occurs, send a 400 Bad Request response with the error message
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Serialize the MRC-721 genesis inscription data to JSON
	genesisDataJSON, err := jsoniter.Marshal(genesisData)
	if err != nil {
		// If serialization fails, send a 400 Bad Request response with the error message
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to marshal genesis data"})
		return
	}

	// Send the serialized JSON data to the client with a 200 OK response
	c.Data(http.StatusOK, "application/json", genesisDataJSON)
}

type GetAddressMrc721ListResult struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    GetAddressMrc721ListData `json:"data"`
}

type GetAddressMrc721ListData struct {
	List     []satmine.WebInscription `json:"list"`
	AllCount int                      `json:"all_count"`
}

// GetAddressMrc721List godoc
// @Summary Retrieve a paginated list of MRC-721 inscriptions for a given address
// @Schemes
// @Description Retrieves a paginated list of MRC-721 inscriptions for a given address and MRC-721 name, sorted by inscription number
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address"
// @Param mrc721name query string false "MRC-721 Name"
// @Param pageIndex query int false "Page Index" default(0)
// @Param pageSize query int false "Page Size" default(100)
// @Success 200 {array} satmine.WebInscription "List of MRC-721 inscriptions"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressmrc721list [get]
func GetAddressMrc721List(c *gin.Context) {
	// Retrieve query parameters
	address := c.Query("address")
	mrc721Name := c.DefaultQuery("mrc721name", "")
	pageIndexStr := c.DefaultQuery("pageIndex", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "100")

	// Validate required parameters
	if address == "" {
		c.JSON(400, GetAddressMrc721ListResult{
			Code:    400,
			Message: "Address is required",
		})
		return
	}

	// Convert pageIndex and pageSize to integers
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil {
		c.JSON(400, GetAddressMrc721ListResult{
			Code:    400,
			Message: "Invalid page index",
		})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(400, GetAddressMrc721ListResult{
			Code:    400,
			Message: "Invalid page size",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAddressMrc721List method on the BTOrdIdx object of the store
	inscriptions, allCount, err := store.OrdIdx.GetAddressMrc721List(address, mrc721Name, pageIndex, pageSize)
	if err != nil {
		c.JSON(501, GetAddressMrc721ListResult{
			Code:    501,
			Message: err.Error(),
		})
		return
	}
	if inscriptions == nil {
		inscriptions = []satmine.WebInscription{}
	}

	// Create and send success response
	result := GetAddressMrc721ListResult{
		Code:    200,
		Message: "Success",
		Data: GetAddressMrc721ListData{
			List:     inscriptions,
			AllCount: allCount,
		},
	}
	c.JSON(http.StatusOK, result)
}

type GetAddressMrc721BarResult struct {
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

// GetAddressMrc721Bar godoc
// @Summary Retrieve MRC-721 inscriptions by address
// @Schemes
// @Description Retrieves a list of unique MRC-721 inscriptions associated with a given address
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address to query MRC-721 inscriptions"
// @Success 200 {object} GetAddressMrc721BarResult "List of unique MRC-721 inscription names"
// @Failure 501 {object} GetAddressMrc721BarResult "Error message if retrieval fails"
// @Router /mrc20/addressmrc721bar [get]
func GetAddressMrc721Bar(c *gin.Context) {
	// Retrieve the address from the query string
	address := c.Query("address")
	if address == "" {
		c.JSON(501, GetAddressMrc721BarResult{
			Code:    501,
			Message: "Address is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAddressMrc721Bar method on the BTOrdIdx object of the store
	names, err := store.OrdIdx.GetAddressMrc721Bar(address)
	if err != nil {
		// If an error occurs, send a 501 response with the error message
		c.JSON(501, GetAddressMrc721BarResult{
			Code:    501,
			Message: err.Error(),
		})
		return
	}

	// Create and send success response
	result := GetAddressMrc721BarResult{
		Code:    200,
		Message: "Success",
		Data:    names,
	}
	c.JSON(http.StatusOK, result)
}

type GetMrc721CollectionsResult struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    []satmine.WebCollections `json:"data"`
}

// GetMrc721CollectionsHandler godoc
// @Summary Retrieve MRC-721 collections by name
// @Schemes
// @Description Retrieves a list of collections associated with a given MRC-721 name
// @Tags mrc20
// @Accept json
// @Produce json
// @Param mrc721name query string true "MRC-721 name to query collections"
// @Success 200 {object} GetMrc721CollectionsResult "List of MRC-721 collections"
// @Failure 501 {object} GetMrc721CollectionsResult "Error message if retrieval fails"
// @Router /mrc20/mrc721collections [get]
func GetMrc721CollectionsHandler(c *gin.Context) {
	// Retrieve the MRC721 name from the query string
	mrc721name := c.Query("mrc721name")
	if mrc721name == "" {
		c.JSON(501, GetMrc721CollectionsResult{
			Code:    501,
			Message: "MRC721 name is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetMrc721Collections method on the BTOrdIdx object of the store
	collections, err := store.OrdIdx.GetMrc721Collections(mrc721name)
	if err != nil {
		// If an error occurs, send a 501 response with the error message
		c.JSON(501, GetMrc721CollectionsResult{
			Code:    501,
			Message: err.Error(),
		})
		return
	}

	// Create and send success response
	result := GetMrc721CollectionsResult{
		Code:    200,
		Message: "Success",
		Data:    collections,
	}
	c.JSON(http.StatusOK, result)
}

// ValidateMRCName checks whether a given MRC721 or MRC20 name exists in the database.
// @Summary Validate MRC721 or MRC20 name
// @Schemes
// @Description Checks if a given MRC721 or MRC20 name exists in the database
// @Tags mrc20
// @Accept json
// @Produce json
// @Param name query string true "Name to validate"
// @Param kind query string true "Kind of token ('mrc721' or 'mrc20')"
// @Success 200 {object} map[string]interface{} "Response with validation result and message"
// @Failure 400 {object} string "Error message if validation fails"
// @Router /mrc20/validatename [get]
func ValidateMRCName(c *gin.Context) {
	// Get name and kind parameters from the query
	name := c.Query("name")
	kind := c.Query("kind")

	// Use the glo package's Instance() method to get the BTOrdIdx object
	store := store.Instance()

	// Call the GetValidateMRC721OrMRC20Name method to check if the name exists
	exists, err := store.OrdIdx.GetValidateMRC721OrMRC20Name(name, kind)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 501, "message": err.Error(), "data": false})
		return
	}

	// Return the result as a JSON response to the client
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": exists, "message": "Validation completed"})
}

// GetGenesisMRC721ProtocolHandler handles requests for retrieving MRC721 genesis protocol data.
// @Summary Retrieve MRC721 genesis protocol
// @Schemes
// @Description Retrieves the genesis protocol data for a given MRC721 name
// @Tags mrc20
// @Accept json
// @Produce json
// @Param mrc721name query string true "MRC721 Name"
// @Success 200 {object} satmine.MRC721Protocol "The MRC721 genesis protocol data"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/genesisprotocol [get]
func GetGenesisMRC721ProtocolHandler(c *gin.Context) {
	// Retrieve MRC721 name from query parameters
	mrc721name := c.Query("mrc721name")
	if mrc721name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "MRC721 name is required"})
		return
	}

	// Use the global or context-specific BTOrdIdx instance
	store := store.Instance() // This should be modified according to how the instance is accessed

	// Call the GetGenesisMRC721Protocol method to get the genesis protocol data
	genesisProtocol, err := store.OrdIdx.GetGenesisMRC721Protocol(mrc721name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 501, "message": err.Error()})
		return
	}

	// Return the genesis protocol data as JSON to the client
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": genesisProtocol, "message": "success"})
}

// Define structs to match the JSON structure
type GetAddressMrc20BarResult struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    GetAddressMrc20BarData `json:"data"`
}

type GetAddressMrc20BarData struct {
	Bars []satmine.WebMrc20Bar `json:"bars"`
}

// GetAddressMrc20Bar godoc
// @Summary Retrieve a list of MRC-20 balances for a given address, optionally filtered by token name
// @Schemes
// @Description Retrieves a list of MRC-20 balances for a given address, showing each token's name and available balance. If a token name is provided, filter the results accordingly.
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address"
// @Param mrc20Name query string false "MRC20 Token Name"
// @Success 200 {object} GetAddressMrc20BarResult "List of MRC-20 balances"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressmrc20bar [get]
func GetAddressMrc20Bar(c *gin.Context) {
	// Retrieve query parameters
	address := c.Query("address")
	// mrc20Name 是非必须的，如果未提供，则默认为 ""
	mrc20Name := c.DefaultQuery("mrc20Name", "")

	//fmt.Println("GetAddressMrc20Bar address=", address, " mrc20Name=", mrc20Name)

	// Validate required parameters
	if address == "" {
		c.JSON(http.StatusBadRequest, GetAddressMrc20BarResult{
			Code:    http.StatusBadRequest,
			Message: "Address is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAddressMrc20Bar method on the BTOrdIdx object of the store, passing the mrc20Name parameter, which might be an empty string
	bars, err := store.OrdIdx.GetAddressMrc20Bar(address, mrc20Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetAddressMrc20BarResult{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		})
		return
	}

	// Create and send success response
	result := GetAddressMrc20BarResult{
		Code:    http.StatusOK,
		Message: "Success",
		Data: GetAddressMrc20BarData{
			Bars: bars,
		},
	}
	c.JSON(http.StatusOK, result)
}

// GetAddressMrc20ListResult defines the structure for the API response.
// This structure includes the status code, message, and data related to MRC-20 inscriptions.
type GetAddressMrc20ListResult struct {
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    GetAddressMrc20ListData `json:"data"`
}

// GetAddressMrc20ListData defines the structure for the data section of the API response.
// This structure encapsulates a list of MRC-20 inscriptions along with the total count.
type GetAddressMrc20ListData struct {
	List     []satmine.WebMrc20Inscription `json:"list"`
	AllCount int                           `json:"all_count"`
}

// GetAddressMrc20List godoc
// @Summary Retrieve a paginated list of MRC-20 inscriptions for a given address
// @Schemes
// @Description Retrieves a paginated list of MRC-20 inscriptions for a given address and MRC-20 name, sorted by inscription number
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address"
// @Param mrc20name query string false "MRC-20 Name"
// @Param pageIndex query int false "Page Index" default(0)
// @Param pageSize query int false "Page Size" default(100)
// @Success 200 {object} GetAddressMrc20ListResult "List of MRC-20 inscriptions"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressmrc20list [get]
func GetAddressMrc20List(c *gin.Context) {
	// Retrieve query parameters
	address := c.Query("address")
	mrc20Name := c.DefaultQuery("mrc20name", "")
	pageIndexStr := c.DefaultQuery("pageIndex", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "100")

	// Validate required parameters
	if address == "" {
		c.JSON(400, GetAddressMrc20ListResult{
			Code:    400,
			Message: "Address is required",
		})
		return
	}

	// Convert pageIndex and pageSize to integers
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil {
		c.JSON(400, GetAddressMrc20ListResult{
			Code:    400,
			Message: "Invalid page index",
		})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(400, GetAddressMrc20ListResult{
			Code:    400,
			Message: "Invalid page size",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAddressMrc20List method on the BTOrdIdx object of the store
	inscriptions, allCount, err := store.OrdIdx.GetAddressMrc20List(address, mrc20Name, pageIndex, pageSize)
	if err != nil {
		c.JSON(501, GetAddressMrc20ListResult{
			Code:    501,
			Message: err.Error(),
		})
		return
	}
	if inscriptions == nil {
		inscriptions = []satmine.WebMrc20Inscription{}
	}

	// Create and send success response
	result := GetAddressMrc20ListResult{
		Code:    200,
		Message: "Success",
		Data: GetAddressMrc20ListData{
			List:     inscriptions,
			AllCount: allCount,
		},
	}
	c.JSON(http.StatusOK, result)
}

// Define a struct to match the JSON structure for the GetAddressMrc721BarPlusResult
type GetAddressMrc721BarPlusResult struct {
	Code    int                         `json:"code"`
	Message string                      `json:"message"`
	Data    GetAddressMrc721BarPlusData `json:"data"`
}

type GetAddressMrc721BarPlusData struct {
	Bars []satmine.WebMrc721Bar `json:"bars"`
}

// GetAddressMrc721BarPlus godoc
// @Summary Retrieve statistics of MRC-721 inscriptions for a given address
// @Schemes
// @Description Retrieves statistics like total amount, total power, and total reward of MRC-721 inscriptions for a given address
// @Tags mrc20
// @Accept json
// @Produce json
// @Param address query string true "Address"
// @Success 200 {object} GetAddressMrc721BarPlusResult "Statistics of MRC-721 inscriptions"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressmrc721barplus [get]
func GetAddressMrc721BarPlus(c *gin.Context) {

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		// Handle the panic, convert it to error if needed
	// 		fmt.Println("GetAddressMrc721BarPlus err= ", r)
	// 		// 打印堆栈跟踪
	// 		fmt.Printf("Stack Trace:\n%s\n", debug.Stack())
	// 		os.Exit(1)
	// 	}
	// }()

	// Retrieve query parameter
	address := c.Query("address")

	// Validate required parameters
	if address == "" {
		c.JSON(400, GetAddressMrc721BarPlusResult{
			Code:    400,
			Message: "Address is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAddressMrc721BarPlus method on the BTOrdIdx object of the store
	webMrc721Bars, err := store.OrdIdx.GetAddressMrc721BarPlus(address)
	if err != nil {
		c.JSON(500, GetAddressMrc721BarPlusResult{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	if webMrc721Bars == nil {
		webMrc721Bars = []satmine.WebMrc721Bar{}
	}

	// Create and send success response
	result := GetAddressMrc721BarPlusResult{
		Code:    200,
		Message: "Success",
		Data: GetAddressMrc721BarPlusData{
			Bars: webMrc721Bars,
		},
	}
	c.JSON(http.StatusOK, result)
}

// Define a struct to match the JSON structure of the response
type GetAddressMrc721HoldersResult struct {
	Code    int                         `json:"code"`
	Message string                      `json:"message"`
	Data    GetAddressMrc721HoldersData `json:"data"`
}

type GetAddressMrc721HoldersData struct {
	Holders  []satmine.WebMrc721Holder `json:"holders"`
	AllCount int                       `json:"all_count"`
}

// GetAddressMrc721Holders godoc
// @Summary Retrieve a paginated list of MRC-721 holders for a given MRC-721 name
// @Schemes
// @Description Retrieves a paginated list of MRC-721 holders for a given MRC-721 name, sorted by the number of inscriptions per address in descending order
// @Tags mrc20
// @Accept json
// @Produce json
// @Param mrc721name query string true "MRC-721 Name"
// @Param pageIndex query int false "Page Index" default(0)
// @Param pageSize query int false "Page Size" default(100)
// @Success 200 {object} GetAddressMrc721HoldersResult "List of MRC-721 holders along with the total count"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/addressmrc721holders [get]
func GetAddressMrc721Holders(c *gin.Context) {
	// Retrieve query parameters
	mrc721Name := c.Query("mrc721name")
	if mrc721Name == "" {
		c.JSON(400, GetAddressMrc721HoldersResult{
			Code:    400,
			Message: "MRC-721 name is required",
		})
		return
	}

	pageIndexStr := c.DefaultQuery("pageIndex", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "100")

	// Convert pageIndex and pageSize to integers
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil {
		c.JSON(400, GetAddressMrc721HoldersResult{
			Code:    400,
			Message: "Invalid page index",
		})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(400, GetAddressMrc721HoldersResult{
			Code:    400,
			Message: "Invalid page size",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetAddressMrc721Holders method on the BTOrdIdx object of the store
	holders, allCount, err := store.OrdIdx.GetAddressMrc721Holders(mrc721Name, pageIndex, pageSize)
	if err != nil {
		c.JSON(501, GetAddressMrc721HoldersResult{
			Code:    501,
			Message: err.Error(),
		})
		return
	}
	if holders == nil {
		holders = []satmine.WebMrc721Holder{}
	}

	// Create and send success response
	result := GetAddressMrc721HoldersResult{
		Code:    200,
		Message: "Success",
		Data: GetAddressMrc721HoldersData{
			Holders:  holders,
			AllCount: allCount,
		},
	}
	c.JSON(http.StatusOK, result)
}

// ScanMissingBlocksResult defines the structure for the API response for the ScanMissingBlocks API.
type ScanMissingBlocksResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []int  `json:"data,omitempty"` // Omitempty will not show "data" in JSON if it's null
}

// ScanMissingBlocks godoc
// @Summary Scans and retrieves a list of missing block numbers
// @Schemes
// @Description Scans the blockchain database from a starting block to an ending block and retrieves a list of missing block numbers
// @Tags mrc20
// @Accept json
// @Produce json
// @Param begin query int true "Begin Block Number"
// @Param end query int true "End Block Number"
// @Success 200 {object} ScanMissingBlocksResult "List of missing block numbers"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/scanmissingblocks [get]
func ScanMissingBlocks(c *gin.Context) {
	// Retrieve query parameters
	beginStr := c.Query("begin")
	endStr := c.Query("end")

	// Convert beginStr and endStr to integers
	begin, err := strconv.Atoi(beginStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ScanMissingBlocksResult{
			Code:    400,
			Message: "Invalid begin block number",
		})
		return
	}
	end, err := strconv.Atoi(endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ScanMissingBlocksResult{
			Code:    400,
			Message: "Invalid end block number",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the ScanMissingBlocks method on the BTOrdIdx object of the store
	missingBlocks, err := store.OrdIdx.ScanMissingBlocks(begin, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ScanMissingBlocksResult{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// Ensure that Data is always an array, even if it's empty
	if missingBlocks == nil {
		missingBlocks = []int{}
	}
	fmt.Println("missingBlocks =", missingBlocks)

	// Create and send success response
	result := ScanMissingBlocksResult{
		Code:    200,
		Message: "Success not missing blocks",
		Data:    missingBlocks,
	}
	c.JSON(http.StatusOK, result)
}

// Define a struct to match the JSON structure for the GetGenesisDataResult
type GetGenesisDataResult struct {
	Code    int                       `json:"code"`    // Status code of the response
	Message string                    `json:"message"` // Description message about the result
	Data    satmine.Mrc721GenesisData `json:"data"`    // Genesis data of a MRC-721 name
}

// GetGenesisData godoc
// @Summary Retrieve genesis data for a given MRC-721 name
// @Schemes
// @Description Retrieves genesis data including identification details and statistics relevant to the inscription of a given MRC-721 name
// @Tags mrc20
// @Accept json
// @Produce json
// @Param mrc721name query string true "MRC-721 name"
// @Success 200 {object} GetGenesisDataResult "Genesis data of the specified MRC-721 name"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/genesisdata [get]
func GetGenesisData(c *gin.Context) {
	// Retrieve query parameter for MRC-721 name
	mrc721name := c.Query("mrc721name")

	// Validate required parameters
	if mrc721name == "" {
		c.JSON(400, GetGenesisDataResult{
			Code:    400,
			Message: "MRC-721 name is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetGenesisData method on the BTOrdIdx object of the store
	genesisData, err := store.OrdIdx.GetGenesisData(mrc721name)
	if err != nil {
		c.JSON(500, GetGenesisDataResult{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// Create and send success response
	result := GetGenesisDataResult{
		Code:    200,
		Message: "Success",
		Data:    genesisData,
	}
	c.JSON(http.StatusOK, result)
}

// Define a struct to match the JSON structure for the GetBurnInfoResult
type GetBurnInfoResult struct {
	Code    int                 `json:"code"`    // Response code
	Message string              `json:"message"` // Response message
	Data    satmine.WebBurnInfo `json:"data"`    // Burn information data
}

// GetBurnInfo godoc
// @Summary Retrieve burn information for a specific inscription
// @Schemes
// @Description Retrieves burn information including balance, power, burn amount, and other relevant details for a given inscription ID
// @Tags mrc20
// @Accept json
// @Produce json
// @Param inscriptionID query string true "Inscription ID"
// @Success 200 {object} GetBurnInfoResult "Burn information for the inscription"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/burninfo [get]
func GetBurnInfo(c *gin.Context) {
	// Retrieve query parameter
	inscriptionID := c.Query("inscriptionID")

	// Validate required parameters
	if inscriptionID == "" {
		c.JSON(400, GetBurnInfoResult{
			Code:    400,
			Message: "Inscription ID is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetBurnInfo method on the BTOrdIdx object of the store
	webBurnInfo, err := store.OrdIdx.GetBurnInfo(inscriptionID)
	if err != nil {
		c.JSON(500, GetBurnInfoResult{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	// Create and send success response
	result := GetBurnInfoResult{
		Code:    200,
		Message: "Success",
		Data:    webBurnInfo,
	}
	c.JSON(http.StatusOK, result)
}

// WebMrcAllInscriptionResult represents the data structure returned by the GetMrcAllInscription API endpoint
type WebMrcAllInscriptionResult struct {
	Code    int                          `json:"code"`
	Message string                       `json:"message"`
	Data    satmine.WebMrcAllInscription `json:"data"`
}

// GetMrcAllInscription godoc
// @Summary Retrieve detailed information about a specific MRC inscription
// @Schemes
// @Description associated inscription data, genesis data, and the relevant MRC-20 or MRC-721 protocol details. MrcType(mrc721,mrc20,unknown,inscription)
// @Tags mrc20
// @Accept json
// @Produce json
// @Param inscriptionId query string true "Inscription ID"
// @Success 200 {object} WebMrcAllInscriptionResult "Detailed information about the MRC inscription"
// @Failure 400 {object} string "Error message if the inscription ID is not provided or invalid"
// @Failure 501 {object} string "Error message if there's an internal error during data retrieval"
// @Router /mrc20/mrcallinscription [get]
func GetMrcAllInscription(c *gin.Context) {
	// Retrieve query parameters
	inscriptionId := c.Query("inscriptionId")

	// Validate required parameters
	if inscriptionId == "" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "Inscription ID is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetMrcAllInscription method on the BTOrdIdx object of the store
	mrcAllInscription, err := store.OrdIdx.GetMrcAllInscription(inscriptionId)
	if err != nil {
		c.JSON(501, gin.H{
			"code":    501,
			"message": err.Error(),
		})
		return
	}

	// Create and send success response
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Success",
		"data":    mrcAllInscription,
	})
}

// Define a struct to match the JSON structure for the GetLotteryListResult
type GetLotteryListResult struct {
	Code    int                `json:"code"`
	Message string             `json:"message"`
	Data    GetLotteryListData `json:"data"`
}

type GetLotteryListData struct {
	Lotteries []satmine.LotteryData `json:"lotteries"`
}

// GetLotteryList godoc
// @Summary Retrieve lottery data list for a given MRC-721 name
// @Schemes
// @Description Retrieves a list of lottery data including block height, address, inscription ID, and win amount for a specific MRC-721 name
// @Tags mrc20
// @Accept json
// @Produce json
// @Param mrc721name query string true "MRC-721 Name"
// @Success 200 {object} GetLotteryListResult "Lottery data list for the specified MRC-721 name"
// @Failure 400 {object} string "Error message if retrieval fails"
// @Router /mrc20/lotterylist [get]
func GetLotteryList(c *gin.Context) {
	// Retrieve query parameter
	mrc721name := c.Query("mrc721name")

	// Validate required parameters
	if mrc721name == "" {
		c.JSON(http.StatusBadRequest, GetLotteryListResult{
			Code:    400,
			Message: "MRC-721 Name is required",
		})
		return
	}

	// Retrieve the store instance from the global context
	store := store.Instance()

	// Call the GetLotteryList method on the BTOrdIdx object of the store
	lotteryData, err := store.OrdIdx.GetLotteryList(mrc721name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GetLotteryListResult{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	if lotteryData == nil {
		lotteryData = []satmine.LotteryData{}
	}

	// Create and send success response
	result := GetLotteryListResult{
		Code:    200,
		Message: "Success",
		Data: GetLotteryListData{
			Lotteries: lotteryData,
		},
	}
	c.JSON(http.StatusOK, result)
}
