package task

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/proggarapsody/bitbottle/api/backend"
	"github.com/proggarapsody/bitbottle/internal/bbrepo"
	"github.com/proggarapsody/bitbottle/internal/format"
	"github.com/proggarapsody/bitbottle/pkg/cmd/factory"
)

// NewCmdPRTask is the parent `pr task` subcommand.
func NewCmdPRTask(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage pull-request tasks (BLOCKER comments on Bitbucket Server / DC)",
	}
	cmd.AddCommand(NewCmdPRTaskList(f))
	cmd.AddCommand(NewCmdPRTaskCreate(f))
	cmd.AddCommand(NewCmdPRTaskResolve(f))
	cmd.AddCommand(NewCmdPRTaskReopen(f))
	return cmd
}

// NewCmdPRTaskList lists BLOCKER comments (tasks) on a pull request.
func NewCmdPRTaskList(f *factory.Factory) *cobra.Command {
	var stateFilter, hostname string

	cmd := &cobra.Command{
		Use:   "list PR_ID",
		Short: "List tasks on a pull request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			cmts, err := client.ListPRComments(ref.Project, ref.Slug, prID)
			if err != nil {
				return err
			}

			tasks := filterTasks(cmts, stateFilter)

			p := taskFields(f, format.ConfigFromCmd(cmd))
			for _, t := range tasks {
				p.AddItem(t)
			}
			return p.Render()
		},
	}
	cmd.Flags().StringVar(&stateFilter, "state", "open", "Filter by task state: open, resolved, all")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	format.RegisterOutputFlags(cmd)
	return cmd
}

// NewCmdPRTaskCreate posts a BLOCKER comment (task) on a pull request.
func NewCmdPRTaskCreate(f *factory.Factory) *cobra.Command {
	var body, hostname string
	var parent int

	cmd := &cobra.Command{
		Use:   "create PR_ID",
		Short: "Create a task on a pull request (Bitbucket Server / DC only)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if body == "" {
				return fmt.Errorf("--body is required")
			}
			ref, prID, client, err := resolveTarget(f, args, hostname)
			if err != nil {
				return err
			}

			in := backend.AddPRCommentInput{
				Text:     body,
				Severity: "BLOCKER",
			}
			if parent != 0 {
				p := parent
				in.Parent = &p
			}

			c, err := client.AddPRComment(ref.Project, ref.Slug, prID, in)
			if err != nil {
				return err
			}
			fmt.Fprintf(f.IOStreams.Out, "Created task #%d on pull request #%d\n", c.ID, prID)
			return nil
		},
	}
	cmd.Flags().StringVar(&body, "body", "", "Task body (required)")
	cmd.Flags().IntVar(&parent, "parent", 0, "Anchor task as reply under an existing comment by its ID")
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// NewCmdPRTaskResolve marks a task as RESOLVED.
func NewCmdPRTaskResolve(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "resolve PR_ID TASK_ID",
		Short: "Resolve a task on a pull request (Bitbucket Server / DC only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolveTarget(f, args[:1], hostname)
			if err != nil {
				return err
			}
			taskID, err := parseCommentID(args[1])
			if err != nil {
				return err
			}
			setter, err := backend.AsPRCommentStateSetter(client, ref.Host)
			if err != nil {
				return err
			}
			return setter.SetPRCommentState(ref.Project, ref.Slug, prID, taskID, "RESOLVED")
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// NewCmdPRTaskReopen marks a task as OPEN.
func NewCmdPRTaskReopen(f *factory.Factory) *cobra.Command {
	var hostname string

	cmd := &cobra.Command{
		Use:   "reopen PR_ID TASK_ID",
		Short: "Reopen a resolved task on a pull request (Bitbucket Server / DC only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ref, prID, client, err := resolveTarget(f, args[:1], hostname)
			if err != nil {
				return err
			}
			taskID, err := parseCommentID(args[1])
			if err != nil {
				return err
			}
			setter, err := backend.AsPRCommentStateSetter(client, ref.Host)
			if err != nil {
				return err
			}
			return setter.SetPRCommentState(ref.Project, ref.Slug, prID, taskID, "OPEN")
		},
	}
	cmd.Flags().StringVar(&hostname, "hostname", "", "Bitbucket hostname")
	return cmd
}

// resolveTarget resolves the PR target (host, project, slug) and PR ID.
// args[0] must be the PR ID string.
func resolveTarget(f *factory.Factory, args []string, hostnameFlag string) (bbrepo.RepoRef, int, backend.Client, error) {
	prID, err := parsePRID(args[0])
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}
	ref, err := factory.ResolveTarget(f, nil, hostnameFlag)
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}
	ref.Project = strings.ToUpper(ref.Project)
	client, err := f.Backend(ref.Host)
	if err != nil {
		return bbrepo.RepoRef{}, 0, nil, err
	}
	return ref, prID, client, nil
}

// filterTasks filters comments to BLOCKER-severity ones, optionally by state.
// stateFilter: "open" (default), "resolved", "all".
func filterTasks(cmts []backend.PRComment, stateFilter string) []backend.PRComment {
	sf := strings.ToLower(stateFilter)
	out := make([]backend.PRComment, 0, len(cmts))
	for _, c := range cmts {
		if c.Severity != "BLOCKER" {
			continue
		}
		switch sf {
		case "resolved":
			if strings.EqualFold(c.State, "RESOLVED") {
				out = append(out, c)
			}
		case "all":
			out = append(out, c)
		default: // "open" or unrecognised
			if c.State == "" || strings.EqualFold(c.State, "OPEN") {
				out = append(out, c)
			}
		}
	}
	return out
}

func taskFields(f *factory.Factory, cfg format.OutputConfig) *format.Printer[backend.PRComment] {
	isTTY := f.IOStreams.IsStdoutTTY()
	p := format.New[backend.PRComment](f.IOStreams.Out, isTTY, cfg)
	p.AddField(format.Field[backend.PRComment]{Name: "id", Header: "ID", Extract: func(c backend.PRComment) any { return c.ID }})
	p.AddField(format.Field[backend.PRComment]{Name: "state", Header: "STATE", Extract: func(c backend.PRComment) any {
		if c.State == "" {
			return "OPEN"
		}
		return c.State
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "author", Header: "AUTHOR", Extract: func(c backend.PRComment) any {
		if c.Author.Slug != "" {
			return c.Author.Slug
		}
		return c.Author.DisplayName
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "body", Header: "BODY", Extract: func(c backend.PRComment) any {
		if isTTY && len(c.Text) > 60 {
			return c.Text[:60]
		}
		return c.Text
	}})
	// JSON-only fields
	p.AddField(format.Field[backend.PRComment]{Name: "severity", Header: "SEVERITY", JSONOnly: true, Extract: func(c backend.PRComment) any { return c.Severity }})
	p.AddField(format.Field[backend.PRComment]{Name: "version", Header: "VERSION", JSONOnly: true, Extract: func(c backend.PRComment) any { return c.Version }})
	p.AddField(format.Field[backend.PRComment]{Name: "createdAt", Header: "CREATED", JSONOnly: true, Extract: func(c backend.PRComment) any {
		return c.CreatedAt.Format(time.RFC3339)
	}})
	p.AddField(format.Field[backend.PRComment]{Name: "parentId", Header: "PARENT", JSONOnly: true, Extract: func(c backend.PRComment) any { return c.ParentID }})
	return p
}

func parsePRID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid PR ID %q: must be a positive integer", arg)
	}
	return id, nil
}

func parseCommentID(arg string) (int, error) {
	id, err := strconv.Atoi(arg)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid TASK_ID %q: must be a positive integer", arg)
	}
	return id, nil
}
