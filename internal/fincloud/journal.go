package fincloud

import (
	"context"
	"encoding/json"
	"net/http"
)

func (c *Client) FetchJournalTrxTypes(ctx context.Context) (map[string]string, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		return c.NewRequest(ctx, http.MethodGet, "/bukuBesar/laporan/mutasiAkun//listvalues", nil)
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var intermediate struct {
		Data struct {
			Result struct {
				NoAkun []struct {
					ID          string `json:"id"`
					Description string `json:"descr"`
				} `json:"noakun"`
			} `json:"result"`
		} `json:"data"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&intermediate); err != nil {
		return nil, err
	}

	if intermediate.Status != "ok" {
		return nil, ErrDataFetchFailed
	}

	return func() map[string]string {
		m := make(map[string]string, len(intermediate.Data.Result.NoAkun))
		for _, entry := range intermediate.Data.Result.NoAkun {
			m[entry.ID] = entry.Description
		}
		return m
	}(), nil
}
