package SteamHttp

//Object that is passed back from a market search on the steam community
type marketSearch struct {
	Success      bool   `json:"success"`
	Start        int    `json:"start"`
	Pagesize     int    `json:"pagesize"`
	Total_count  int    `json:"total_count"`
	Results_html string `json:"results_html"`
}
