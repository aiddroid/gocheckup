package checkup

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// HttpChecker对象
type HttpChecker struct {
	// api 端点名称
	Name string `json:"endpoint_name"`

	// 端点URL
	URL string `json:"endpoint_url"`

	// 端点http状态码. 默认是 http.StatusOK.
	UpStatus int `json:"up_status,omitempty"`

	// 端点健康检查可容忍的最大RTT时间
	ThresholdRTT time.Duration `json:"threshold_rtt,omitempty"`

	// 健康检查时http 响应体必须包含的字符串，如果设置了，会接收全部 http 响应体进行比对
	MustContain string `json:"must_contain,omitempty"`

	// 健康检查时http 响应体必须不包含的字符串，如果设置了，会接收全部 http 响应体进行比对
	MustNotContain string `json:"must_not_contain,omitempty"`

	// 一次检查时，发送的请求数
	Attempts int `json:"attempts,omitempty"`

	// 请求之间的间隔时间
	AttemptSpacing time.Duration `json:"attempt_spacing,omitempty"`

	// http客户端，如果不设置，默认使用DefaultHTTPClient
	// used.
	Client *http.Client `json:"-"`

	// 发送http请求时附加的头信息
	Headers http.Header `json:"headers,omitempty"`
}

// HttpChecker实现Check接口
func (c HttpChecker) Check() (Result, error) {
	if c.Attempts < 1 {
		c.Attempts = 1
	}

	if c.Client == nil {
		c.Client = http.DefaultClient
	}

	if c.UpStatus == 0 {
		c.UpStatus = 200
	}

	result := Result{Title: c.Name, Endpoint: c.URL, Timestamp: Timestamp()}
	req, err := http.NewRequest("GET", c.URL, nil)
	if err != nil {
		return result, err
	}

	if c.Headers != nil {
		for key, header := range c.Headers {
			req.Header.Add(key, strings.Join(header, ", "))
		}
	}

	result.Times = c.doChecks(req)
	return c.conclude(result), nil
}

// 执行具体的Check操作
func (c HttpChecker) doChecks(req *http.Request) Attempts {
	checks := make(Attempts, c.Attempts)
	for i := 0; i < c.Attempts; i++ {
		start := time.Now()
		resp, err := c.Client.Do(req)
		checks[i].RTT = time.Since(start)
		if err != nil {
			checks[i].Error = err.Error()
			continue
		}
		err = c.checkDown(resp)
		if err != nil {
			checks[i].Error = err.Error()
		}
		resp.Body.Close()
		if c.AttemptSpacing > 0 {
			time.Sleep(c.AttemptSpacing)
		}
	}
	return checks
}

// 检查是否宕机
func (c HttpChecker) checkDown(resp *http.Response) error {
	// Check status code
	if resp.StatusCode != c.UpStatus {
		return fmt.Errorf("response status %s", resp.Status)
	}

	// Check response body
	if c.MustContain == "" && c.MustNotContain == "" {
		return nil
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}
	body := string(bodyBytes)
	if c.MustContain != "" && !strings.Contains(body, c.MustContain) {
		return fmt.Errorf("response does not contain '%s'", c.MustContain)
	}
	if c.MustNotContain != "" && strings.Contains(body, c.MustNotContain) {
		return fmt.Errorf("response contains '%s'", c.MustNotContain)
	}

	return nil
}

//获取检查结论
func (c HttpChecker) conclude(result Result) Result {
	result.ThresholdRTT = c.ThresholdRTT

	// Check errors (down)
	for i := range result.Times {
		if result.Times[i].Error != "" {
			result.Down = true
			return result
		}
	}

	// Check round trip time (degraded)
	if c.ThresholdRTT > 0 {
		stats := result.ComputeStats()
		if stats.Median > c.ThresholdRTT {
			result.Notice = fmt.Sprintf("Median RTT exceeded threshold (%s)", c.ThresholdRTT)
			result.Degraded = true
			return result
		}
	}

	result.Healthy = true
	return result
}
