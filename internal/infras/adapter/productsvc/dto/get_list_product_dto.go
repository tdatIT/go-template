package dto

const getListProductURL = "/api/v1/mock-svc/GetMockData"

type GetListProductReq struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

func (GetListProductReq) URL() string {
	return getListProductURL
}

type GetListProductResp struct {
	Products []*struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		PackageID string `json:"package_id"`
		Thumbnail string `json:"thumbnail"`
		Price     string `json:"price"`
		CreatedAt string `json:"created_at"`
		IsActive  bool   `json:"is_active"`
	}
}
