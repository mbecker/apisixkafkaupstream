package utils

import (
	"testing"
)

var jsonObject = `
{
    "glossary": {
        "title": "example glossary",
		"GlossDiv": {
            "title": "The Kafka Message ID configured globally by jsonkey",
			"GlossList": {
                "GlossEntry": {
                    "ID": "SGML",
					"SortAs": "SGML",
					"GlossTerm": "Standard Generalized Markup Language",
					"Acronym": "SGML",
					"Abbrev": "ISO 8879:1986",
					"GlossDef": {
                        "para": "A meta-markup language, used to create markup languages such as DocBook.",
						"GlossSeeAlso": ["GML", "XML"]
                    },
					"GlossSee": "markup"
                }
            }
        }
    }
}
`

func BenchmarkJsonCompact(b *testing.B) {
	data := []byte(jsonObject)
	for i := 0; i < b.N; i++ {
		_, err := JsonEncodingCompact(data)
		if err != nil {
			b.Fail()
			continue
		}
	}
}

func BenchmarkJsonUnmarshal(b *testing.B) {
	data := []byte(jsonObject)
	for i := 0; i < b.N; i++ {
		_, err := JsonEncoding(data)
		if err != nil {
			b.Fail()
			continue
		}
	}
}
