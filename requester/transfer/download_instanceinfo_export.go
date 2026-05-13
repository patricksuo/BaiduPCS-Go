package transfer

type (
	// RangeGenMode 线程分配方式
	RangeGenMode int32

	//Range 请求范围
	Range struct {
		Begin int64 `json:"begin,omitempty"`
		End   int64 `json:"end,omitempty"`
	}

	// DownloadInstanceInfoExport 断点续传信息, 用于序列化到本地状态文件
	DownloadInstanceInfoExport struct {
		RangeGenMode RangeGenMode `json:"range_gen_mode,omitempty"`
		TotalSize    int64        `json:"total_size,omitempty"`
		GenBegin     int64        `json:"gen_begin,omitempty"`
		BlockSize    int64        `json:"block_size,omitempty"`
		Ranges       RangeList    `json:"ranges,omitempty"`
	}
)

const (
	// RangeGenMode_Default 根据parallel平均生成
	RangeGenMode_Default RangeGenMode = 0
	// RangeGenMode_BlockSize 根据blockSize生成
	RangeGenMode_BlockSize RangeGenMode = 1
)
