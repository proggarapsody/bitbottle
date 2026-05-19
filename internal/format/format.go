package format

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/internal/tableprinter"
)

// OutputFormat enumerates supported output modes.
type OutputFormat string

const (
	// FormatTable is the default human-readable table output.
	FormatTable OutputFormat = ""
	// FormatJSON emits JSON, optionally piped through jq.
	FormatJSON OutputFormat = "json"
	// FormatYAML emits YAML.
	FormatYAML OutputFormat = "yaml"
	// FormatTemplate renders results through a Go text/template.
	FormatTemplate OutputFormat = "template"
)

// OutputConfig captures the user-facing output-format selection — exactly
// one Format is active per invocation. JQExpr is only meaningful when
// Format == FormatJSON; Template only when Format == FormatTemplate.
// JSONFields, when non-nil, restricts JSON output to only the named fields;
// nil means all fields (backward compatible).
type OutputConfig struct {
	Format     OutputFormat
	JQExpr     string
	Template   string
	JSONFields []string // nil = all fields; non-nil = only these fields
}

// RegisterOutputFlags registers --json, --yaml, --jq and --template as
// persistent flags on cmd. Use this on the root command; subcommands
// inherit the flags via cobra's persistent-flag merging.
//
// Exposed so tests that instantiate individual subcommands in isolation
// (without going through the root) can still parse output-format flags
// the same way production does.
// jsonNoArgSentinel is the value set when --json is used without an argument
// (e.g. --json alone). It signals "all fields" and must be non-empty so that
// pflag's NoOptDefVal mechanism activates.
const jsonNoArgSentinel = "*"

func RegisterOutputFlags(cmd *cobra.Command) {
	f := cmd.PersistentFlags()
	f.String("json", "", "Output as JSON (optionally select fields: --json field1,field2)")
	f.Lookup("json").NoOptDefVal = jsonNoArgSentinel
	f.Bool("yaml", false, "Output as YAML")
	f.String("jq", "", "Filter JSON output with a jq expression")
	f.String("template", "", "Format output with a Go template")
}

// ConfigFromCmd reads the four output-format persistent flags from any
// cobra.Command. Works from any subcommand because Cobra merges persistent
// flags from ancestors. Falls back to FormatTable when the flags are not
// registered (e.g. a test calling a subcommand without a parent).
func ConfigFromCmd(cmd *cobra.Command) OutputConfig {
	jsonChanged := lookupChanged(cmd, "json")
	jsonValue := lookupString(cmd, "json")
	jqExpr := lookupString(cmd, "jq")
	yamlMode := lookupBool(cmd, "yaml")
	tmpl := lookupString(cmd, "template")
	switch {
	case yamlMode:
		return OutputConfig{Format: FormatYAML}
	case tmpl != "":
		return OutputConfig{Format: FormatTemplate, Template: tmpl}
	case jsonChanged:
		var jsonFields []string
		if jsonValue != "" && jsonValue != jsonNoArgSentinel {
			jsonFields = splitFields(jsonValue)
		}
		return OutputConfig{Format: FormatJSON, JQExpr: jqExpr, JSONFields: jsonFields}
	default:
		return OutputConfig{Format: FormatTable, JQExpr: jqExpr}
	}
}

func splitFields(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if f := strings.TrimSpace(p); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func lookupChanged(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Lookup(name) == nil {
		return false
	}
	return cmd.Flags().Changed(name)
}

func lookupBool(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Lookup(name) == nil {
		return false
	}
	v, _ := cmd.Flags().GetBool(name)
	return v
}

func lookupString(cmd *cobra.Command, name string) string {
	if cmd.Flags().Lookup(name) == nil {
		return ""
	}
	v, _ := cmd.Flags().GetString(name)
	return v
}

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
	// ColorFunc, if set, is applied to the rendered string in TTY table mode
	// only. It is intentionally NOT applied to JSON output, where consumers
	// expect raw values, nor in non-TTY table mode (piped output).
	ColorFunc func(string) string
}

// Printer renders a slice of T in the correct output mode.
type Printer[T any] struct {
	w          io.Writer
	isTTY      bool
	cfg        OutputConfig
	singleItem bool
	fields     []Field[T]
	items      []T
}

// New constructs a Printer.
func New[T any](w io.Writer, isTTY bool, cfg OutputConfig) *Printer[T] {
	return &Printer[T]{
		w:     w,
		isTTY: isTTY,
		cfg:   cfg,
	}
}

// SetSingleItem marks the printer to emit a single object (JSON/YAML/template)
// instead of an array.
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

// ValidateJSONFields returns an error if any requested field is unknown.
// Only called when cfg.JSONFields is non-nil (i.e. --json field1,field2 was used).
func (p *Printer[T]) ValidateJSONFields() error {
	if len(p.cfg.JSONFields) == 0 {
		return nil
	}
	valid := make(map[string]bool)
	for _, f := range p.fields {
		valid[f.Name] = true
		for _, a := range f.Aliases {
			valid[a] = true
		}
	}
	var unknown []string
	for _, req := range p.cfg.JSONFields {
		if !valid[req] {
			unknown = append(unknown, req)
		}
	}
	if len(unknown) > 0 {
		available := make([]string, 0, len(p.fields))
		for _, f := range p.fields {
			available = append(available, f.Name)
		}
		sort.Strings(available)
		return fmt.Errorf("unknown JSON field(s): %s; available: %s",
			strings.Join(unknown, ", "),
			strings.Join(available, ", "))
	}
	return nil
}

// Render writes all items in the appropriate format.
func (p *Printer[T]) Render() error {
	if p.cfg.JQExpr != "" && p.cfg.Format != FormatJSON {
		return fmt.Errorf("--jq requires --json")
	}

	if p.cfg.Format == FormatJSON && p.cfg.JSONFields != nil {
		if err := p.ValidateJSONFields(); err != nil {
			return err
		}
	}

	switch p.cfg.Format {
	case FormatJSON:
		return p.renderJSON()
	case FormatYAML:
		return p.renderYAML()
	case FormatTemplate:
		if p.cfg.Template == "" {
			return fmt.Errorf("--template: expression is empty")
		}
		return p.renderTemplate()
	default:
		return p.renderTable()
	}
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
			val := fmt.Sprintf("%v", f.Extract(item))
			if p.isTTY && f.ColorFunc != nil {
				val = f.ColorFunc(val)
			}
			tp.AddField(val)
		}
		tp.EndRow()
	}
	return tp.Render()
}

func (p *Printer[T]) renderJSON() error {
	if p.singleItem {
		if len(p.items) == 0 {
			_, err := fmt.Fprintln(p.w, "{}")
			return err
		}
		obj := p.itemToMap(p.items[0])
		if p.cfg.JQExpr != "" {
			return p.runJQ(obj)
		}
		return json.NewEncoder(p.w).Encode(obj)
	}

	objs := make([]any, len(p.items))
	for i, item := range p.items {
		objs[i] = p.itemToMap(item)
	}

	if p.cfg.JQExpr != "" {
		return p.runJQ(objs)
	}

	enc := json.NewEncoder(p.w)
	enc.SetIndent("", "")
	return enc.Encode(objs)
}

func (p *Printer[T]) renderYAML() error {
	if p.singleItem {
		if len(p.items) == 0 {
			return WriteYAML(p.w, map[string]any{})
		}
		return WriteYAML(p.w, p.itemToMap(p.items[0]))
	}
	objs := make([]map[string]any, len(p.items))
	for i, item := range p.items {
		objs[i] = p.itemToMap(item)
	}
	return WriteYAML(p.w, objs)
}

func (p *Printer[T]) renderTemplate() error {
	if p.singleItem {
		var item map[string]any
		if len(p.items) > 0 {
			item = p.itemToMap(p.items[0])
		}
		return WriteTemplate(p.w, p.cfg.Template, item)
	}
	objs := make([]map[string]any, len(p.items))
	for i, item := range p.items {
		objs[i] = p.itemToMap(item)
	}
	return WriteTemplate(p.w, p.cfg.Template, objs)
}

func (p *Printer[T]) itemToMap(item T) map[string]any {
	wantAll := p.cfg.JSONFields == nil
	wantSet := make(map[string]bool, len(p.cfg.JSONFields))
	for _, f := range p.cfg.JSONFields {
		wantSet[f] = true
	}
	m := make(map[string]any, len(p.fields))
	for _, f := range p.fields {
		if !wantAll {
			wanted := wantSet[f.Name]
			if !wanted {
				for _, a := range f.Aliases {
					if wantSet[a] {
						wanted = true
						break
					}
				}
			}
			if !wanted {
				continue
			}
		}
		v := f.Extract(item)
		// Honour an "unknown" / "not applicable" signal from the field by
		// omitting the key entirely when Extract returns nil — mirroring
		// the omitempty pointer convention on the underlying struct so
		// callers cannot misread null-as-zero.
		if v == nil {
			continue
		}
		m[f.Name] = v
	}
	return m
}

func (p *Printer[T]) runJQ(input any) error {
	q, err := gojq.Parse(p.cfg.JQExpr)
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
