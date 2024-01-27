package xkblayouts

import (
	"encoding/xml"
	"fmt"
	"os"
)

func ParseLayouts(path string) (*XkbConfigRegistry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	registry := &XkbConfigRegistry{}
	err = xml.NewDecoder(file).Decode(registry)
	if err != nil {
		return nil, fmt.Errorf("decode xml: %w", err)
	}

	return registry, nil
}

func (r *XkbConfigRegistry) GetLayoutPrettyName(layout, variant string) string {
	for _, l := range r.LayoutList.Layout {
		if l.ConfigItem.Name == layout {
			if variant == "" {
				return l.ConfigItem.Description
			}

			for _, v := range l.VariantList.Variant {
				if v.ConfigItem.Name == variant {
					return v.ConfigItem.Description
				}
			}
		}
	}

	return ""
}

func (r *XkbConfigRegistry) GetLayoutAndVariantFromPrettyName(prettyName string) (string, string) {
	for _, l := range r.LayoutList.Layout {
		if l.ConfigItem.Description == prettyName {
			return l.ConfigItem.Name, ""
		}

		for _, v := range l.VariantList.Variant {
			if v.ConfigItem.Description == prettyName {
				return l.ConfigItem.Name, v.ConfigItem.Name
			}
		}
	}

	return "", ""
}
