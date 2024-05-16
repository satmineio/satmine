package rpc

import (
	"github.com/gin-gonic/gin"
)

// Registering routes to the gin engine
func RegisterRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		eg := v1.Group("mrc20")
		{
			eg.GET("/latestblock", GetLatestBlock)
			eg.GET("/blockbyheight", GetBlockByHeight)
			eg.GET("/blockbyhash", GetBlockByHash)
			eg.GET("/blocks", GetBlocks)
			eg.GET("/addressinfo", GetAddressInfo)
			eg.GET("/addressbalance", GetAddressBalance)
			eg.GET("/addressbalances", GetAddressBalances)
			eg.GET("/inscription", GetInscription)
			eg.POST("/miningprofitchart", MiningProfitChart)
			eg.GET("/inscriptionplus", GetInscriptionPlus)
			eg.GET("/allmrc721", GetAllMrc721)
			eg.GET("/onemrc721", GetOneMrc721)
			eg.GET("/addressmrc721list", GetAddressMrc721List)
			eg.GET("/addressmrc721bar", GetAddressMrc721Bar)
			eg.GET("/mrc721collections", GetMrc721CollectionsHandler)
			eg.GET("/validatename", ValidateMRCName)
			eg.GET("/genesisprotocol", GetGenesisMRC721ProtocolHandler)
			eg.GET("/addressmrc20bar", GetAddressMrc20Bar)
			eg.GET("/addressmrc20list", GetAddressMrc20List)
			eg.GET("/addressmrc721barplus", GetAddressMrc721BarPlus)
			eg.GET("/addressmrc721holders", GetAddressMrc721Holders)
			eg.GET("/scanmissingblocks", ScanMissingBlocks)
			eg.GET("/genesisdata", GetGenesisData)
			eg.GET("/burninfo", GetBurnInfo)
			eg.GET("/mrcallinscription", GetMrcAllInscription)
			eg.GET("/lotterylist", GetLotteryList)

			eg.POST("/postrecord", PostRecord)
			eg.GET("/getrecords", GetRecords)

			eg.POST("/hookevents", ordHookEvents)

			// eg.POST("/GetMockBlock", GetMockBlock)
			// eg.POST("/WriteMockBlock", WriteMockBlock)
			// eg.POST("/JoinMockTransfer", JoinMockTransfer)
			// eg.POST("/JoinMockInscription", JoinMockInscription)
			// eg.POST("/GoMockFullBlock", GoMockFullBlock)

		}
	}
}
