// Package ini is an LL(1) parser for con***REMOVED***guration ***REMOVED***les.
//
//	Example:
//	sections, err := ini.OpenFile("/path/to/***REMOVED***le")
//	if err != nil {
//		panic(err)
//	}
//
//	pro***REMOVED***le := "foo"
//	section, ok := sections.GetSection(pro***REMOVED***le)
//	if !ok {
//		fmt.Printf("section %q could not be found", pro***REMOVED***le)
//	}
//
// Below is the BNF that describes this parser
//	Grammar:
//	stmt -> value stmt'
//	stmt' -> epsilon | op stmt
//	value -> number | string | boolean | quoted_string
//
//	section -> [ section'
//	section' -> value section_close
//	section_close -> ]
//
//	SkipState will skip (NL WS)+
//
//	comment -> # comment' | ; comment'
//	comment' -> epsilon | value
package ini
