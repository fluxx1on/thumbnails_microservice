package serial

type ThumbnailData []byte

type SnippetSerializer struct {
	ChannelTitle string `json:"channelTitle"`
	Title        string `json:"title"`
	Thumbnails   struct {
		Maxres struct {
			Url    string `json:"url"`
			Width  int32  `json:"width"`
			Height int32  `json:"height"`
		} `json:"maxres"`
	} `json:"thumbnails"`
}

type Item struct {
	Id      string            `json:"id"`
	Snippet SnippetSerializer `json:"snippet"`
}

func (i *Item) GetId() string {
	return i.Id
}

func (i *Item) GetChannelTitle() string {
	return i.Snippet.ChannelTitle
}

func (i *Item) GetTitle() string {
	return i.Snippet.Title
}

func (i *Item) GetUrl() string {
	return i.Snippet.Thumbnails.Maxres.Url
}

func (i *Item) GetWidth() int32 {
	return i.Snippet.Thumbnails.Maxres.Width
}

func (i *Item) GetHeight() int32 {
	return i.Snippet.Thumbnails.Maxres.Height
}

type ListVideoSerializer struct {
	Items []Item `json:"items"`
}

func (l *ListVideoSerializer) IsEmpty() bool {
	if l.Items == nil {
		return true
	}
	return len(l.Items) == 0
}

type Video struct {
	I    *Item
	Data ThumbnailData
}

func (video *Video) GetData() ThumbnailData {
	return video.Data
}
