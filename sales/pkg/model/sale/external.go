package sale

type ExternalSale struct {
	ID        string              `json:"id"`
	LineItems []*ExternalLineItem `json:"line_items"`
}

type ExternalLineItem struct {
	ID        string `json:"id"`
	ProductID string `json:"productID" obfuscate:"true"`
	Quantity  int    `json:"quantity"`
	UnitPrice int    `json:"unit_price"`
}
