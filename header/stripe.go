package header

/*********************************/

type StripeInfo struct {
	Subscriptionid string  `json:"subscriptionid"`
	Productid      string  `json:"productid"`
	Priceid        string  `json:"priceid"`
	Customerid     string  `json:"customerid"`
	Email          string  `json:"email"`
	Userkey        string  `json:"userkey"`
	Price          float64 `json:"price"`
	Sessionid      string  `json:"sessionid"`
	State          int64   `json:"state"`
	Updatetime     string  `json:"updatetime"`
}
