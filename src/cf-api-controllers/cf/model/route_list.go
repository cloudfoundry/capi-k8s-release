package model

type RouteList struct {
	Pagination struct {
		TotalPages int `json:"total_pages"`
	}
	Resources []Route           `json:"resources"`
	Included  RouteListIncluded `json:"included"`
}

type RouteListIncluded struct {
	Spaces  []Space  `json:"spaces"`
	Domains []Domain `json:"domains"`
}
