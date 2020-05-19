package checkup

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Fs 对象存储
type Fs struct {
	Dir      string `json:"dir"`
	Filename string `json:"filename"`
}

// 实现Store接口进行文件存储
func (fs Fs) Store(results []Result) error {
	// 打开文件
	fpath := filepath.Join(fs.Dir, fs.Filename)
	f, err := os.OpenFile(fpath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0775)
	if err != nil {
		return err
	}

	// 写入检查结果
	bs, err := json.Marshal(results)
	if err != nil {
		return err
	}

	f.Write(bs)
	f.Close()
	return nil
}
