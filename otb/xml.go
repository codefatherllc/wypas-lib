package otb

type xmlItems struct {
	Items []xmlItem `xml:"item"`
}

type xmlItem struct {
	ID      uint16    `xml:"id,attr"`
	FromID  uint16    `xml:"fromid,attr"`
	ToID    uint16    `xml:"toid,attr"`
	Name    string    `xml:"name,attr"`
	Article string    `xml:"article,attr"`
	Plural  string    `xml:"plural,attr"`
	Attrs   []xmlAttr `xml:"attribute"`
}

type xmlAttr struct {
	Key      string    `xml:"key,attr"`
	Value    string    `xml:"value,attr"`
	Children []xmlAttr `xml:"attribute"`
}
