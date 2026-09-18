package fincloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (c *Client) DownloadReport(ctx context.Context, name string, params ...any) (string, error) {
	p := make([]string, len(params))
	for i, v := range params {
		switch t := v.(type) {
		case string:
			p[i] = t
		case time.Time:
			p[i] = t.Format("2006-01-02")
		default:
			p[i] = fmt.Sprint(t)
		}
	}

	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	q := url.Values{}
	q.Set("nm", name)
	q.Set("type", "csv")
	q.Set("p", string(b))

	path := "/system/laporanUmum/data/lap?" + q.Encode()
	path = strings.ReplaceAll(path, "+", "%20") // space encoding

	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		form := url.Values{}
		form.Set("sessionId", c.sessionIDValue())

		req, err := c.NewRequest(ctx, http.MethodGet, path, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrDataFetchFailed
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("\uFEFF")) // remove BOM if exists

	return string(data), nil
}

func (c *Client) DownloadReportFromMaintenance(ctx context.Context, file, path string) (string, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodGet, "/system/downloaderlaporan/download.php", nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("file", file)
		q.Set("path", path)
		q.Set("sessionId", c.sessionIDValue())
		req.URL.RawQuery = q.Encode()
		return req, nil
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrDataFetchFailed
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	data = bytes.TrimPrefix(data, []byte("\uFEFF")) // remove BOM if exists

	return string(data), nil
}

func (c *Client) ListMaintenanceReportFiles(ctx context.Context, folder string) ([]string, error) {
	return c.listMaintenanceDir(ctx, folder, "/app/report")
}

func (c *Client) listMaintenanceDir(ctx context.Context, file, dir string) ([]string, error) {
	resp, err := c.DoRequest(ctx, func() (*http.Request, error) {
		req, err := c.NewRequest(ctx, http.MethodGet, "/system/downloaderlaporan/pembuatan/loadorDownload", nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Set("file", file)
		q.Set("jenis", "Folder")
		q.Set("pathfolder", dir)
		req.URL.RawQuery = q.Encode()
		return req, nil
	})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var intermediate struct {
		Data struct {
			Result struct {
				PathFolder string `json:"pathfolder"`
				List       []struct {
					File  string `json:"file"`
					Jenis string `json:"jenis"`
				} `json:"list"`
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

	var files []string

	for _, item := range intermediate.Data.Result.List {
		switch item.Jenis {
		case "Folder":
			childFiles, err := c.listMaintenanceDir(ctx, item.File, intermediate.Data.Result.PathFolder)
			if err != nil {
				return nil, ErrDataFetchFailed
			}
			files = append(files, childFiles...)
		case "File":
			files = append(files, intermediate.Data.Result.PathFolder+"/"+item.File)
		}
	}

	return files, nil
}
