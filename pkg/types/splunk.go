package splunk

//This file includes types for un-marshalling Splunk XML responses

// The XML response from Splunk is wrapped in a <results> tag
type Results struct {
	// XMLName xml.Name `xml:"results"`
	Results []Result `xml:"result"`
	Preview string   `xml:"preview,attr"`
}

type Result struct {
	// XMLName xml.Name `xml:"result"`
	Fields []Field `xml:"field"`
	Offset string  `xml:"offset,attr"`
}

type Field struct {
	// XMLName xml.Name `xml:"field"`
	Value Value  `xml:"value"`
	Key   string `xml:"k,attr"`
	V     string `xml:"v"`
}

type Value struct {
	// XMLName xml.Name `xml:"value"`
	Text string `xml:"text"`
}
