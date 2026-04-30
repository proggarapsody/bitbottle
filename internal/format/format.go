package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/itchyny/gojq"

	"github.com/proggarapsody/bitbottle/internal/tableprinter"
)

// Field describes one output column and JSON key for type T.
type Field[T any] struct {
	Name    string
	Header  string
	Extract func(T) any
	// Aliases are alternative names accepted by --json; they do not appear in
	// the valid-fields error list so the canonical name is always preferred.
	Aliases []string
	// JSONOnly marks a field that is available via --json but omitted from the
	// default table output (e.g. UUID, webURL columns that clutter the TTY view).
	JSONOnly bool
}

// Printer renders a slice of T in the correct output mode.
type Printer[T any] struct {
	w          io.Writer
	isTTY      bool
	jsonFields string
	jqExpr     string
	singleItem bool
	fields     []Field[T]
	items      []T
}

// New constructs a Printer.
func New[T any](w io.Writer, isTTY bool, jsonFields, jqExpr string) *Printer[T] {
	return &Printer[T]{
		w:          w,
		isTTY:      isTTY,
		jsonFields: jsonFields,
		jqExpr:     jqExpr,
	}
}

// SetSingleItem marks the printer to emit a JSON object instead of an array.
func (p *Printer[T]) SetSingleItem() {
	p.singleItem = true
}

// AddField registers a field.
func (p *Printer[T]) AddField(f Field[T]) {
	p.fields = append(p.fields, f)
}

// AddItem enqueues one result item.
func (p *Printer[T]) AddItem(item T) {
	p.items = append(p.items, item)
}

// Render writes all items in the appropriate format.
func (p *Printer[T]) Render() error {
	if p.jqExpr != "" && p.jsonFields == "" {
		return fmt.Errorf("--jq requires --json")
	}

	if p.jsonFields != "" {
		return p.renderJSON()
	}

	return p.renderTable()
}

func (p *Printer[T]) renderTable() error {
	if len(p.items) == 0 {
		return nil
	}

	tableFields := make([]Field[T], 0, len(p.fields))
	for _, f := range p.fields {
		if !f.JSONOnly {
			tableFields = append(tableFields, f)
		}
	}

	headers := make([]string, len(tableFields))
	for i, f := range tableFields {
		headers[i] = f.Header
	}

	tp := tableprinter.New(p.w, p.isTTY, 0)
	if p.isTTY {
		tp.AddHeader(headers...)
	}
	for _, item := range p.items {
		for _, f := range tableFields {
			tp.AddField(fmt.Sprintf("%v", f.Extract(item)))
		}
		tp.EndRow()
	}
	return tp.Render()
}

func (p *Printer[T]) renderJSON() error {
	requested, err := p.resolveFields()
	if err != nil {
		return err
	}

	if p.singleItem {
		if len(p.items) == 0 {
			_, err := fmt.Fprintln(p.w, "{}")
			return err
		}
		obj := p.itemToMap(p.items[0], requested)
		if p.jqExpr != "" {
			return p.runJQ(obj)
		}
		return json.NewEncoder(p.w).Encode(obj)
	}

	objs := make([]any, len(p.items))
	for i, item := range p.items {
		objs[i] = p.itemToMap(item, requested)
	}

	if p.jqExpr != "" {
		return p.runJQ(objs)
	}

	enc := json.NewEncoder(p.w)
	enc.SetIndent("", "")
	return enc.Encode(objs)
}

func (p *Printer[T]) resolveFields() ([]Field[T], error) {
	names := strings.Split(p.jsonFields, ",")
	fieldByName := make(map[string]Field[T], len(p.fields))
	for _, f := range p.fields {
		fieldByName[f.Name] = f
		for _, alias := range f.Aliases {
			fieldByName[alias] = f
		}
	}

	result := make([]Field[T], 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		f, ok := fieldByName[name]
		if !ok {
			valid := make([]string, len(p.fields))
			for i, f := range p.fields {
				valid[i] = f.Name
			}
			return nil, fmt.Errorf("unknown field %q; valid fields: %s", name, strings.Join(valid, ", "))
		}
		// Use the requested name as the JSON key so that --json link produces
		// {"link": ...} and --jq '.[].link' works as expected.
		if name != f.Name {
			f.Name = name
		}
		result = append(result, f)
	}
	return result, nil
}

func (p *Printer[T]) itemToMap(item T, fields []Field[T]) map[string]any {
	m := make(map[string]any, len(fields))
	for _, f := range fields {
		m[f.Name] = f.Extract(item)
	}
	return m
}

func (p *Printer[T]) runJQ(input any) error {
	q, err := gojq.Parse(p.jqExpr)
	if err != nil {
		return fmt.Errorf("--jq: %w", err)
	}
	iter := q.Run(input)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return fmt.Errorf("--jq: %w", err)
		}
		out, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(p.w, "%s\n", out); err != nil {
			return err
		}
	}
	return nil
}
