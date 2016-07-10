package downloader

type Downloaded interface {
	Filename() string
	Data() []byte
}

type downloaded struct {
	name string
	data []byte
}

func (d *downloaded) Filename() string {
	return d.name
}

func (d *downloaded) Data() []byte {
	return d.data
}

func NewDownloaded(name string, data []byte) Downloaded {
	return &downloaded{
		name: name,
		data: data,
	}
}
