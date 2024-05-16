package satmine

// HookBlock represents a block with its inscriptions and transfers.
type HookBlock struct {
	BlockHeight  string `json:"block_height"`
	BlockHash    string `json:"block_hash"`
	Timestamp    int64  `json:"timestamp"`
	Inscriptions []HookInscription
	Transfers    []HookTransfer
}

// HookInscription represents a single inscription record from the JSON response.
type HookInscription struct {
	ID                      string  `json:"id"`
	Number                  int     `json:"number"`
	Address                 string  `json:"address"`
	Offset                  string  `json:"offset"`
	Sat                     int64   `json:"sat"`
	BlockHeight             int     `json:"block_height"`
	OrdinalHeight           int     `json:"ordinal_height"`
	ContentByte             *[]byte `json:"content_byte"` // binary data
	ContentType             string  `json:"content_type"`
	ContentLength           int     `json:"content_length"`
	CurseType               *string `json:"curse_type"`
	InscriptionFee          int     `json:"inscription_fee"`
	InscriptionInputIndex   int     `json:"inscription_input_index"`
	InscriptionOutputValue  int     `json:"inscription_output_value"`
	SatpointPostInscription string  `json:"satpoint_post_inscription"`
	TransfersPreInscription int     `json:"transfers_pre_inscription"`
	TxIndex                 int     `json:"tx_index"`
}

// HookTransfer represents a single transfer record from the JSON response.
type HookTransfer struct {
	ID                      string `json:"id"`
	Type                    string `json:"type"`
	ToAddress               string `json:"to_address"`
	PostTransferOutputValue int    `json:"post_transfer_output_value"`
	SatpointPostTransfer    string `json:"satpoint_post_transfer"`
	SatpointPreTransfer     string `json:"satpoint_pre_transfer"`
	TxIndex                 int    `json:"tx_index"`
}

type LotteryData struct {
	BlockHeight    string `json:"block_height"`
	BlockHash      string `json:"block_hash"`
	BlockTimestamp int64  `json:"timestamp"`
	Address        string `json:"address"`
	InscriptionID  string `json:"inscription_id"`
	Number         int    `json:"number"`
	Mrc721name     string `json:"mrc721name"`
	WinAmount      string `json:"win_amount"`
	JackpotAccum   string `json:"jackpot_accum"`
	Round          int    `json:"round"`
	Winp           string `json:"winp"`
	Dist           string `json:"dist"`
}

type SimpleHookBlock struct {
	BlockHeight string `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	Timestamp   int64  `json:"timestamp"`
}

type WebMrc721Bar struct {
	Mrc721name  string `json:"mrc721name"`
	Mrc20name   string `json:"mrc20name"`
	Amount      string `json:"amount"`
	TotalPower  string `json:"total_power"`
	TotalReward string `json:"total_reward"`
	First721ID  string `json:"first_721_id"`
	BlockHeight int    `json:"block_height"`
}

type WebInscription struct {
	Inscription HookInscription `json:"inscription"`
	Mrc721name  string          `json:"mrc721name"`
	Mrc20name   string          `json:"mrc20name"`
	MinedAmount string          `json:"mined_amount"` // Number of final digs per inscription
	Power       string          `json:"power"`        // The arithmetic value of each inscription, determined by the BurnNum parameter, defaults to 1000.
}

type WebCollections struct {
	Id   string            `json:"id"`
	Meta WebCollectionItem `json:"meta"`
}

type WebCollectionItem struct {
	Name string `json:"name"`
}

type WebMrc20Bar struct {
	Mrc20name    string `json:"mrc20name"`
	Balance      string `json:"balance"`
	Avaliable    string `json:"avaliable"`
	Transferable string `json:"transferable"`
}

type WebMrc20Inscription struct {
	Inscription HookInscription `json:"inscription"`
	Mrc20name   string          `json:"mrc20name"`
	Amount      string          `json:"amount"`
}

type WebMrc721Holder struct {
	Rank       string `json:"rank"`
	Address    string `json:"address"`
	Percentage string `json:"percentage"`
	Amount     string `json:"amount"`
}

type WebBurnInfo struct {
	Balance    string `json:"balance"`
	Mrc721name string `json:"mrc721name"`
	Mrc20name  string `json:"mrc20name"`
	Power      string `json:"power"`
	BurnAmount string `json:"burn_amount"`
	BurnMax    string `json:"burn_max"`
	Unit       string `json:"unit"`
	Boost      string `json:"boost"`
}

type WebMrcAllInscription struct {
	MrcType     string            `json:"mrc_type"`
	Inscription HookInscription   `json:"inscription"`
	GenesisData Mrc721GenesisData `json:"genesis_data"`
	Mrc20P      *MRC20Protocol    `json:"mrc20_protocol"`
	Mrc721P     *MRC721Protocol   `json:"mrc721_protocol"`
}
