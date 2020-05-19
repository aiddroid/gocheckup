/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package checkup

import (
	"time"
)

// Checkup 对象， 管理一系列checker
type Checkup struct {
	// 检查项
	Checkers []Checker `json:"checkers,omitempty"`

	// 时间戳
	Timestamp time.Time `json:"timestamp,omitempty"`

	// 检查结果存储方式
	Storage Storage `json:"storage,omitempty"`

	// 检查结果通知方式
	Notifier Notifier `json:notifier,omitempty`
}

// Checker检查后的结果
type Result struct {
	// 标题
	Title string `json:"title,omitempty"`

	// api端点
	Endpoint string `json:"endpoint,omitempty"`

	// 时间戳
	Timestamp int64 `json:"timestamp,omitempty"`

	// ThresholdRTT，降级前的最大 RTT. 设置为0表示忽略
	ThresholdRTT time.Duration `json:"threshold,omitempty"`

	// 多个独立检查企图的列表
	Times Attempts `json:"times,omitempty"`

	// 是否健康
	Healthy bool `json:"healthy,omitempty"`

	// 是否降级
	Degraded bool `json:"degraded,omitempty"`

	// 是否宕机
	Down bool `json:"down,omitempty"`

	// 提示信息
	Notice string `json:"notice,omitempty"`

	// 消息
	Message string `json:"notice,omitempty"`
}

// Checker检查接口
type Checker interface {
	Check() (Result, error)
}

// Storage存储接口
type Storage interface {
	Store([]Result) error
}

// Notifier 通知接口
type Notifier interface {
	Notify([]Result) error
}

// 与端点进行通信的检查企图
type Attempt struct {
	RTT   time.Duration `json:"rtt"`
	Error string        `json:"error,omitempty"`
}

// 根据 RTT进行排序的检查企图列表
type Attempts []Attempt

func Timestamp() int64 {
	return time.Now().UTC().UnixNano()
}

// 对Checkup对象执行Check操作
func (c Checkup) Check() ([]Result, error) {
	results := make([]Result, len(c.Checkers))
	errors := make([]error, len(c.Checkers))

	for i, checker := range c.Checkers {
		results[i], errors[i] = checker.Check()
	}

	return results, nil
}

// 检查结果的统计信息，尤其是有多个检查企图时
type Stats struct {
	Total  time.Duration `json:"total,omitempty"`
	Mean   time.Duration `json:"mean,omitempty"`
	Median time.Duration `json:"median,omitempty"`
	Min    time.Duration `json:"min,omitempty"`
	Max    time.Duration `json:"max,omitempty"`
}

// ComputeStats 计算检查结果的统计信息
func (r Result) ComputeStats() Stats {
	var s Stats

	for _, a := range r.Times {
		s.Total += a.RTT
		if a.RTT < s.Min || s.Min == 0 {
			s.Min = a.RTT
		}
		if a.RTT > s.Max || s.Max == 0 {
			s.Max = a.RTT
		}
	}
	sorted := make(Attempts, len(r.Times))

	half := len(sorted) / 2
	if len(sorted)%2 == 0 {
		s.Median = (sorted[half-1].RTT + sorted[half].RTT) / 2
	} else {
		s.Median = sorted[half].RTT
	}

	s.Mean = time.Duration(int64(s.Total) / int64(len(r.Times)))

	return s
}
