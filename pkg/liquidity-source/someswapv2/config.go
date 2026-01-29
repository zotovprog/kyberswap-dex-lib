package someswapv2

type Config struct {
	// Optional: kept for consistency with other liquidity sources.
	DexID string `json:"dexID,omitempty"`

	// Matches the JSON you provided.
	Factory              string `json:"factory"`
	Router               string `json:"router"`
	Quoter               string `json:"quoter"`
	LPFeeManager         string `json:"lpFeeManager"`
	LiquidityLocker      string `json:"liquidityLocker"`
	CoreModule           string `json:"coreModule"`
	PermissionsRegistry  string `json:"permissionsRegistry"`

	// Optional: used by some list updaters; not in your JSON.
	NewPoolLimit int `json:"newPoolLimit,omitempty"`
}

