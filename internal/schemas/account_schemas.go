package schemas

type AccountIDRequest struct {
	AccountID int `json:"account_id"`
}

type AccountAccountProfileDirRequest struct {
	AccountIDs []int `json:"account_ids"`
}
