package mfd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"sort"
)

type XMLMap struct {
	elements []xmlMapElement
}

type xmlMapElement struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func NewXMLMap(init map[string]string) *XMLMap {
	xmlMap := &XMLMap{
		elements: []xmlMapElement{},
	}

	keys := make([]string, 0, len(init))
	for k := range init {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		xmlMap.Append(k, init[k])
	}

	return xmlMap
}

func (m *XMLMap) MarshalJSON() ([]byte, error) {
	jsM := make(map[string]interface{})
	for _, i := range m.elements {
		jsM[i.XMLName.Local] = i.Value
	}
	return json.Marshal(jsM)
}

func (m *XMLMap) Append(key, value string) {
	if m.elements == nil {
		m.elements = []xmlMapElement{{
			XMLName: xml.Name{Local: key},
			Value:   value,
		}}
		return
	}

	if m.Index(key) == -1 {
		m.elements = append(m.elements, xmlMapElement{
			XMLName: xml.Name{Local: key},
			Value:   value,
		})
	}
}

func (m *XMLMap) Delete(key string) {
	if i := m.Index(key); i != -1 {
		m.elements = append(m.elements[:i], m.elements[i+1:]...)
	}
}

func (m *XMLMap) Index(key string) int {
	for i, el := range m.elements {
		if el.XMLName.Local == key {
			return i
		}
	}

	return -1
}

// MarshalXML marshals the map to XML, with each key in the map being a
// tag and it's corresponding value being it's contents.
func (m *XMLMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m.elements) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for _, element := range m.elements {
		_ = e.Encode(element)
	}

	return e.EncodeToken(start.End())
}

// UnmarshalXML unmarshals the XML into a map of string to strings,
// creating a key in the map for each tag and setting it's value to the
// tags contents.
//
// The fact this function is on the pointer of XMLMap is important, so that
// if m is nil it can be initialized, which is often the case if m is
// nested in another xml structure. This is also why the first thing done
// on the first line is initialize it.
func (m *XMLMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = XMLMap{
		elements: []xmlMapElement{},
	}
	for {
		var e xmlMapElement

		if err := d.Decode(&e); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}

		m.elements = append(m.elements, e)
	}
	return nil
}
