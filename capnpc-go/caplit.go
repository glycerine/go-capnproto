package main

import (
	"fmt"
	"io"
	"strings"
)

// define the Caplit capnproto literal representation
// producing functions. Initially adapted from the WriteJSON implementation.

func (n *node) defineTypeCaplitFuncs(w io.Writer) {
	g_imported["io"] = true
	g_imported["bufio"] = true
	g_imported["bytes"] = true

	fprintf(w, "func (s %s) WriteCapLit(w io.Writer) error {\n", n.name)
	fprintf(w, "b := bufio.NewWriter(w);")
	fprintf(w, "var err error;")
	fprintf(w, "var buf []byte;")
	fprintf(w, "_ = buf;")

	switch n.Which() {
	case NODE_ENUM:
		n.caplitEnum(w)
	case NODE_STRUCT:
		n.caplitStruct(w)
	}

	fprintf(w, "err = b.Flush(); return err\n};\n")

	fprintf(w, "func (s %s) MarshalCapLit() ([]byte, error) {\n", n.name)
	fprintf(w, "b := bytes.Buffer{}; err := s.WriteCapLit(&b); return b.Bytes(), err };")
}

func (n *node) caplitEnum(w io.Writer) {
	fprintf(w, "_, err = b.WriteString(s.String());")
	writeErrCheck(w)
}

// Write statements that will write a caplit struct
func (n *node) caplitStruct(w io.Writer) {
	fprintf(w, `err = b.WriteByte('(');`)
	writeErrCheck(w)
	for i, f := range n.codeOrderFields() {
		if f.DiscriminantValue() != 0xFFFF {
			enumname := fmt.Sprintf("%s_%s", strings.ToUpper(n.name), strings.ToUpper(f.Name()))
			fprintf(w, "if s.Which() == %s {", enumname)
		} else if i != 0 {
			fprintf(w, `
					_, err = b.WriteString(", ");
				`)
			writeErrCheck(w)
		}

		fprintf(w, `_, err = b.WriteString("%s = ");`, f.Name())
		writeErrCheck(w)
		f.caplit(w)
		if f.DiscriminantValue() != 0xFFFF {
			fprintf(w, "};")
		}
	}
	fprintf(w, `err = b.WriteByte(')');`)
	writeErrCheck(w)
}

// This function writes statements that write the field's caplit representation to the bufio.
func (f *Field) caplit(w io.Writer) {

	switch f.Which() {
	case FIELD_SLOT:
		fs := f.Slot()
		// we don't generate setters for Void fields
		if fs.Type().Which() == TYPE_VOID {
			fs.Type().caplit(w)
			return
		}
		fprintf(w, "{ s := s.%s(); ", title(f.Name()))
		fs.Type().caplit(w)
		fprintf(w, "}; ")
	case FIELD_GROUP:
		tid := f.Group().TypeId()
		n := findNode(tid)
		fprintf(w, "{ s := s.%s();", title(f.Name()))

		n.caplitStruct(w)
		fprintf(w, "};")
	}
}

func (t Type) caplit(w io.Writer) {
	switch t.Which() {
	case TYPE_UINT8, TYPE_UINT16, TYPE_UINT32, TYPE_UINT64,
		TYPE_INT8, TYPE_INT16, TYPE_INT32, TYPE_INT64,
		TYPE_FLOAT32, TYPE_FLOAT64, TYPE_BOOL, TYPE_TEXT, TYPE_DATA:
		g_imported["encoding/json"] = true
		fprintf(w, "buf, err = json.Marshal(s);")
		writeErrCheck(w)
		fprintf(w, "_, err = b.Write(buf);")
		writeErrCheck(w)
	case TYPE_ENUM, TYPE_STRUCT:
		// since we handle groups at the field level, only named struct types make it in here
		// so we can just call the named structs caplit dumper
		fprintf(w, "err = s.WriteCapLit(b);")
		writeErrCheck(w)
	case TYPE_LIST:
		typ := t.List().ElementType()
		which := typ.Which()
		if which == TYPE_LIST || which == TYPE_ANYPOINTER {
			// untyped list, cant do anything but report
			// that a field existed.
			//
			// s will be unused in this case, so ignore
			fprintf(w, `_ = s;`)
			fprintf(w, `_, err = b.WriteString("\"untyped list\"");`)
			writeErrCheck(w)
			return
		}
		fprintf(w, "{ err = b.WriteByte('[');")
		writeErrCheck(w)
		fprintf(w, "for i, s := range s.ToArray() {")
		fprintf(w, `if i != 0 { _, err = b.WriteString(", "); };`)
		writeErrCheck(w)
		typ.caplit(w)
		fprintf(w, "}; err = b.WriteByte(']'); };")
		writeErrCheck(w)
	case TYPE_VOID:
		fprintf(w, `_ = s;`)
		fprintf(w, `_, err = b.WriteString("null");`)
		writeErrCheck(w)
	}
}
