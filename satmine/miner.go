package satmine

import (
	"fmt"
	"math/big"
	"sort"
)

// Mrc721GenesisData represents the information about a genesis inscription in the MRC-721 protocol.
// It includes identification details and statistics relevant to the inscription.
type Mrc721GenesisData struct {
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
	GenesisBlockHeight   string `json:"genesis_block_height"`    // Height of the block in the blockchain where the genesis transaction of the inscription was recorded
	GenesisTimestamp     int64  `json:"genesis_timestamp"`       // Timestamp of the block in the blockchain where the genesis transaction of the inscription was recorded
	EndID                string `json:"end_id"`                  // The ID of the inscription that ended the mining round
	EndBlockHeight       string `json:"end_block_height"`        // The block height of the inscription that ended the mining round
	EndTimestamp         int64  `json:"end_timestamp"`           // The timestamp of the block in the blockchain where the inscription that ended the mining round was recorded
	TotalPrizeRound      int    `json:"total_prize_round"`       // The total prize round
	TotalBurn            string `json:"total_burn"`              // The total burn
}

// MiningRewardCalculation contains the results of the CalculateMiningRewards method.
type MiningRewardCalculation struct {
	CurrentPrizePoolAllNum string `json:"current_prize_pool_all_num"` // Total amount of funds extracted in this round (to be subsequently added to the prize pool for accumulation)
	CurrentMiningAllNum    string `json:"current_mining_all_num"`     // Total amount of funds mined in this round (to be deducted from the total capital)
	IsMiningEnd            bool   `json:"is_mining_end"`              // Has the mining round concluded?
	EndReason              string `json:"end_reason"`                 // Reason for conclusion
}

type Mrc721MinerData struct {
	InscriptionsID     string  `json:"inscriptions_id"`
	InscriptionsNumber int     `json:"inscriptions_number"`
	Address            string  `json:"address"`
	BurnNum            string  `json:"burn_num"`
	Tick               string  `json:"tick"`
	MinedAmount        string  `json:"mined_amount"` // Number of final digs per inscription
	Power              big.Int // The arithmetic value of each inscription, determined by the BurnNum parameter, defaults to 1000.
}

type Mrc721MinerMap struct {
	Data map[string]*Mrc721MinerData
}

// CalculateMiningRewards computes the amount of tokens that each mining machine can earn in a single mining round.
// This function takes into account various factors such as the mining machine's performance, the current network difficulty,
// and any other relevant parameters to accurately calculate the mining yield. The result helps miners understand their potential
// earnings from mining activities during each round.
func CalculateMiningRewards(currentHeight string, genesisData *Mrc721GenesisData, firstMrc721 *MRC721Protocol, minerMap *Mrc721MinerMap) (calcResult MiningRewardCalculation, err error) {
	// Parameters:
	// genesisHeight is the block height at genesis
	// blockHeight is the current block height
	// total is the total number of tokens that can be mined
	// beg is the number of tokens mined per block
	// halv is the halving cycle from 1-n
	// dcr is the reduction ratio in mining amount after each halving cycle "0.000" - "1.000"
	// extr is the proportion of funds drawn into the prize pool each round "0.000" - "1.000"
	// minedTokens is the total number of tokens already mined
	// inscriptionsCount is the number of inscriptions participating in mining
	// Returns:
	// tokensPerInscription is the number of tokens each inscription can mine this time
	// eligibleInscriptions is the number of inscriptions that are allowed to receive tokens, during the last mining round not all inscriptions can be allocated, only a fixed number of inscriptions to avoid exceeding the total value
	// prizePoolTokens is the amount of funds that need to be put into the prize pool after drawing
	// currentMiningValue is the total amount of mining in this round

	isPrint := false

	// Convert parameters to big numbers
	genesisHeightBigInt, _ := new(big.Int).SetString(genesisData.BlockHeight, 10) // Block height at genesis
	currentHeightBigInt, _ := new(big.Int).SetString(currentHeight, 10)           // Current block height
	totalBigInt, _ := new(big.Int).SetString(firstMrc721.Token.Total, 10)         // Total number of tokens that can be mined
	beginBigInt, _ := new(big.Int).SetString(firstMrc721.Token.Beg, 10)           // Number of tokens mined per block
	halvBigInt, _ := new(big.Int).SetString(firstMrc721.Token.Halv, 10)           // Halving cycle
	dcrBigInt := stringToPercentageBigInt(firstMrc721.Token.Dcr)                  // Reduction ratio after each halving cycle "0.000" - "1.000"

	totalMinedTokensBigInt, _ := new(big.Int).SetString(genesisData.TotalMinedTokens, 10)         // Total mined tokens
	totalPrizePoolTokensBigInt, _ := new(big.Int).SetString(genesisData.TotalPrizePoolTokens, 10) // Total tokens in prize pool

	//fmt.Println("CalculateMiningRewards debug=", totalBigInt, totalMinedTokensBigInt, totalPrizePoolTokensBigInt)

	// Calculate the remaining unreleased tokens
	remainingTokens := new(big.Int).Sub(totalBigInt, new(big.Int).Add(totalMinedTokensBigInt, totalPrizePoolTokensBigInt))
	// Check if remaining tokens are less than zero, which is an error
	if remainingTokens.Cmp(big.NewInt(0)) < 0 {
		err = fmt.Errorf("error: negative value of remaining tokens")
		return
	}

	// If remaining tokens are zero, return early
	if remainingTokens.Cmp(big.NewInt(0)) == 0 {
		// Funds have been fully released, stop mining
		calcResult.CurrentMiningAllNum = "0"
		calcResult.CurrentPrizePoolAllNum = "0"
		calcResult.IsMiningEnd = true
		calcResult.EndReason = "FullRelease"
		return
	}

	blockCount := new(big.Int).Sub(currentHeightBigInt, genesisHeightBigInt)

	// Calculate the current round
	roundBigInt := new(big.Int).Div(blockCount, halvBigInt)
	// Print the current round for debugging purposes
	if isPrint {
		fmt.Printf("Current round: %s\n", roundBigInt.String())
	}

	currentBlockMining := new(big.Int).Set(beginBigInt)

	// Loop through each round
	for i := big.NewInt(0); i.Cmp(roundBigInt) < 0; i.Add(i, big.NewInt(1)) {
		// Update the current block mining total
		currentBlockMining.Mul(currentBlockMining, new(big.Int).Sub(big.NewInt(1000), dcrBigInt))
		currentBlockMining.Div(currentBlockMining, big.NewInt(1000))

	}

	// If the funds available are greater than the funds to be allocated, the maximum amount that can be allocated is the remaining funds.
	if currentBlockMining.Cmp(remainingTokens) > 0 {
		currentBlockMining.Set(remainingTokens)
	}

	// Halving has come to an end and does not free up funds
	if currentBlockMining.Cmp(big.NewInt(0)) == 0 {
		// Funds fully released, stop mining.
		calcResult.CurrentMiningAllNum = "0"
		calcResult.CurrentPrizePoolAllNum = "0"
		calcResult.IsMiningEnd = true
		calcResult.EndReason = "NotFullyReleased"
		return
	}

	// Print the final current block mining total
	if isPrint {
		fmt.Printf("Total number of coins out of the block after halving: %s\n", currentBlockMining.String())
	}
	// Pumping value per round
	calcResult.CurrentPrizePoolAllNum = "0"

	// Handling of raffle prizes
	if firstMrc721.Ltry != nil {
		poolBigInt := stringToPercentageBigInt(firstMrc721.Ltry.Pool)
		//intvlBigInt, _ := new(big.Int).SetString(firstMrc721.Ltry.Intvl, 10)

		// Calculation of pumping values per round
		prizePoolValue := new(big.Int).Mul(currentBlockMining, poolBigInt)
		prizePoolValue.Div(prizePoolValue, big.NewInt(1000))

		// Deduct the pumping value from the total amount of current block mined
		currentBlockMining.Sub(currentBlockMining, prizePoolValue)

		calcResult.CurrentPrizePoolAllNum = prizePoolValue.String()

	}

	// Increased arithmetic of burning
	if firstMrc721.Burn != nil {
		// Calculate the arithmetic value of each inscription
		unitBigInt, _ := new(big.Int).SetString(firstMrc721.Burn.Unit, 10)
		boostBigInt := stringToPercentageBigInt(firstMrc721.Burn.Boost)

		// Iterate over the minerMap
		for _, minerData := range minerMap.Data {
			//burnNumBigInt, _ := new(big.Int).SetString(minerData.BurnNum, 10)
			burnNumBigInt := new(big.Int).SetBytes([]byte(minerData.BurnNum))
			// Calculate the power value

			//fmt.Println("burnNumBigInt =", burnNumBigInt, "minerData.BurnNum=", minerData.BurnNum)
			powerValue := new(big.Int).Div(burnNumBigInt, unitBigInt)
			powerValue.Mul(powerValue, boostBigInt)
			// Add to the existing Power value
			minerData.Power.Add(&minerData.Power, powerValue)
			// If the power value is greater than 10000, set the power value to 100000 (that's 10 times the power, to avoid over-parameterization issues).
			if minerData.Power.Cmp(big.NewInt(11000)) > 0 {
				minerData.Power.SetInt64(11000)
			}

			//fmt.Println("minerData.Power=", minerData.Power)
			// Print each InscriptionsID's power value
			if isPrint {
				fmt.Printf("InscriptionsID: %s,BurnNum %s,unitBigInt %s, Power: %s, MinedAmount: %s\n", minerData.InscriptionsID, minerData.BurnNum, unitBigInt, minerData.Power.String(), minerData.MinedAmount)
			}
		}
	}

	//Calculate the bonus for each inscription
	residualFunds, _ := powerRewards(*currentBlockMining, minerMap)

	if isPrint {
		fmt.Printf("Funds remaining after mining: %s\n", residualFunds.String())

		//fmt.Println("currentBlockMining", currentBlockMining)
		fmt.Printf("currentBlockMining=%s \n", currentBlockMining.String())
	}

	// Update calcResult.CurrentMiningAllNum to the total number of mined blocks mined minus the remaining funds.
	calcResult.CurrentMiningAllNum = new(big.Int).Sub(currentBlockMining, &residualFunds).String()
	calcResult.IsMiningEnd = false
	return

}

// powerRewards distributes the mining rewards based on the power of each inscription in the minerMap.
// It returns the residual funds that couldn't be distributed due to rounding.
func powerRewards(currentMining big.Int, minerMap *Mrc721MinerMap) (residualFunds big.Int, err error) {

	// Generate a copy value
	var currentBlockMining big.Int
	currentBlockMining.Set(&currentMining)

	// If the mining money is zero
	if currentBlockMining.Cmp(big.NewInt(0)) == 0 {
		// If the total power is zero, then all bonuses are returned directly as residual funds
		return currentBlockMining, nil
	}

	// Calculate total force value
	var totalPower big.Int
	for _, minerData := range minerMap.Data {
		totalPower.Add(&totalPower, &minerData.Power)
	}

	// If the totalizing power is zero
	if totalPower.Cmp(big.NewInt(0)) == 0 {
		// If the total power is zero, then all bonuses are returned directly as residual funds
		return currentBlockMining, nil
	}

	// If the total math power is zero, or if the prize money is sufficient for distribution
	if currentBlockMining.Cmp(big.NewInt(int64(len(minerMap.Data)))) >= 0 {
		// Distribution of bonuses
		var allocated big.Int
		for _, minerData := range minerMap.Data {
			// Calculate the bonus for each inscription
			share := new(big.Int).Mul(&currentBlockMining, &minerData.Power)
			share.Div(share, &totalPower)
			// Distribution of bonuses
			minerData.MinedAmount = share.String()
			allocated.Add(&allocated, share)
		}

		// Calculation of remaining funds
		residualFunds.Sub(&currentBlockMining, &allocated)
	} else {
		// When the prize money is not enough to allocate to each inscription
		type minerDataSlice []*Mrc721MinerData
		var minerSlice minerDataSlice
		for _, v := range minerMap.Data {
			minerSlice = append(minerSlice, v)
		}

		// Then sort the slice
		sort.Slice(minerSlice, func(i, j int) bool {
			if minerSlice[i].Power.Cmp(&minerSlice[j].Power) == 0 {
				return minerSlice[i].InscriptionsNumber < minerSlice[j].InscriptionsNumber
			}
			return minerSlice[i].Power.Cmp(&minerSlice[j].Power) > 0
		})

		// Iterate, but use the new minerSlice
		for i := 0; i < len(minerSlice); i++ {
			minerSlice[i].MinedAmount = "1"
			currentBlockMining.Sub(&currentBlockMining, big.NewInt(1))
			if currentBlockMining.Cmp(big.NewInt(0)) == 0 {
				break
			}
		}
		residualFunds.Set(&currentBlockMining)
	}

	return residualFunds, nil
}
