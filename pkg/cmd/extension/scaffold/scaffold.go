// Package scaffold implements the `bitbottle extension scaffold` command.
package scaffold

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

//go:embed templates
var templateFS embed.FS

// templateData holds the values substituted into every template file.
type templateData struct {
	Name       string
	BinaryName string
	Year       int
}

// NewCmdScaffold returns the `extension scaffold` subcommand.
func NewCmdScaffold(f *factory.Factory) *cobra.Command {
	var lang string
	var dir string

	cmd := &cobra.Command{
		Use:   "scaffold NAME",
		Short: "Generate a new bitbottle extension project from a template",
		Long: `Generate a new bitbottle extension project from a template.

The scaffold is created at <dir>/bitbottle-NAME/ and includes all files
needed to build and release a binary extension compatible with
'bitbottle extension install'.

Examples:
  bitbottle extension scaffold myplugin
  bitbottle extension scaffold myplugin --lang bash
  bitbottle extension scaffold myplugin --dir ~/projects`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return runScaffold(f, name, lang, dir)
		},
	}

	cmd.Flags().StringVar(&lang, "lang", "go", "Language template to use: go or bash")
	cmd.Flags().StringVar(&dir, "dir", ".", "Directory in which to create the extension project")

	return cmd
}

func runScaffold(f *factory.Factory, name, lang, dir string) error {
	if lang != "go" && lang != "bash" {
		return fmt.Errorf("unsupported language %q: must be go or bash", lang)
	}

	binaryName := "bitbottle-" + name
	destRoot := filepath.Join(dir, binaryName)

	if _, err := os.Stat(destRoot); err == nil {
		return fmt.Errorf("directory %s already exists", destRoot)
	}

	data := templateData{
		Name:       name,
		BinaryName: binaryName,
		Year:       time.Now().Year(),
	}

	if err := os.MkdirAll(filepath.Join(destRoot, ".github", "workflows"), 0o755); err != nil {
		return fmt.Errorf("creating scaffold directory: %w", err)
	}

	// Render shared templates (README.md, LICENSE, .github/workflows/release.yml).
	if err := renderDir(templateFS, "templates/shared", destRoot, data, sharedDestPath); err != nil {
		return err
	}

	// Render language-specific templates.
	langDir := "templates/" + lang
	if err := renderDir(templateFS, langDir, destRoot, data, langDestPath(lang, name)); err != nil {
		return err
	}

	fmt.Fprintf(f.IOStreams.Out, "✓ Created extension project %s\n", destRoot)
	return nil
}

// sharedDestPath maps a shared template path to its output path relative to destRoot.
func sharedDestPath(tmplPath string) string {
	// tmplPath example: "templates/shared/README.md.tmpl"
	base := filepath.Base(tmplPath)
	out := strings.TrimSuffix(base, ".tmpl")
	switch out {
	case "release.yml":
		return filepath.Join(".github", "workflows", "release.yml")
	default:
		return out
	}
}

// langDestPath maps a language-specific template path to its output path.
func langDestPath(lang, name string) func(string) string {
	return func(tmplPath string) string {
		base := filepath.Base(tmplPath)
		out := strings.TrimSuffix(base, ".tmpl")
		if lang == "bash" && out == "run.sh" {
			// The bash entry point is the binary itself: bitbottle-NAME
			return "bitbottle-" + name
		}
		return out
	}
}

// renderDir walks srcDir inside fsys, renders each .tmpl file via text/template,
// and writes the result to destRoot using destPath to compute the relative output path.
func renderDir(fsys embed.FS, srcDir, destRoot string, data templateData, destPath func(string) string) error {
	return fs.WalkDir(fsys, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		raw, err := fsys.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		// Use [[ ]] delimiters so templates can contain ${{ }} GitHub Actions
		// expressions without conflicting with Go's default {{ }} syntax.
		tmpl, err := template.New(path).Delims("[[", "]]").Parse(string(raw))
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		rel := destPath(path)
		outPath := filepath.Join(destRoot, rel)

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("creating directory for %s: %w", outPath, err)
		}

		perm := fs.FileMode(0o644)
		// The bash entry point must be executable.
		if filepath.Base(outPath) == "bitbottle-"+data.Name {
			perm = 0o755
		}

		outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", outPath, err)
		}

		if err := tmpl.Execute(outFile, data); err != nil {
			_ = outFile.Close()
			return fmt.Errorf("rendering template %s: %w", path, err)
		}
		if err := outFile.Close(); err != nil {
			return fmt.Errorf("closing file %s: %w", outPath, err)
		}
		return nil
	})
}
