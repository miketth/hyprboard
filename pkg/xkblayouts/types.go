package xkblayouts

import "encoding/xml"

type XkbConfigRegistry struct {
	XMLName    xml.Name   `xml:"xkbConfigRegistry"`
	LayoutList LayoutList `xml:"layoutList"`
}

type ConfigItem struct {
	Name        string `xml:"name"`
	Description string `xml:"description"`
}

type Variant struct {
	ConfigItem ConfigItem `xml:"configItem"`
}

type VariantList struct {
	Variant []Variant `xml:"variant"`
}

type Layout struct {
	ConfigItem  ConfigItem  `xml:"configItem"`
	VariantList VariantList `xml:"variantList"`
}

type LayoutList struct {
	Layout []Layout `xml:"layout"`
}
