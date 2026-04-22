package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/mssantosdev/norn/internal/norn"
	"github.com/mssantosdev/norn/internal/spindle"
	"github.com/mssantosdev/norn/internal/ui/logger"
	"github.com/mssantosdev/norn/internal/ui/styles"
	"github.com/mssantosdev/norn/internal/warps"
	"gopkg.in/yaml.v3"
)

func runWarps(args []string) error {
	if showHelp(warpsHelp(), args) {
		return nil
	}
	w, err := norn.Load(".")
	if err != nil {
		return err
	}
	root := norn.SpindleRoot(w)
	if len(args) == 0 || args[0] == "list" {
		if len(args) == 2 && args[1] == "--view=runtime" {
			return runWarpRuntimeView(root)
		}
		items, err := warps.List(root)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render("Warps"))
		if len(items) == 0 {
			logger.Print(styles.KV("active", "none"))
			return nil
		}
		for _, item := range items {
			label := item.Title
			if strings.TrimSpace(item.Status) != "" {
				label = fmt.Sprintf("%s [%s]", label, item.Status)
			}
			logger.Print(styles.KV(item.ID, label))
		}
		return nil
	}
	if len(args) >= 2 && args[0] == "assign" {
		if len(args) == 1 {
			assignment, err := promptWarpAssignment(root)
			if err != nil {
				return err
			}
			return spindle.SaveAssignment(root, assignment)
		}
		assignment, err := parseWarpAssignArgs(args[1:])
		if err != nil {
			return err
		}
		return spindle.SaveAssignment(root, assignment)
	}
	if len(args) >= 2 && args[0] == "assignment" {
		switch args[1] {
		case "show":
			if len(args) != 4 {
				return fmt.Errorf("usage: norn warps assignment show <weave|thread> <id>")
			}
			item, err := spindle.LoadAssignment(root, args[2], args[3])
			if err != nil {
				return err
			}
			data, err := yaml.Marshal(item)
			if err != nil {
				return err
			}
			logger.Print(styles.Title.Render(fmt.Sprintf("%s %s", item.Kind, item.ID)))
			logger.Print(string(data))
			return nil
		case "remove":
			if len(args) != 4 {
				return fmt.Errorf("usage: norn warps assignment remove <weave|thread> <id>")
			}
			return spindle.DeleteAssignment(root, args[2], args[3])
		default:
			return fmt.Errorf("usage: norn warps assignment <show|remove>")
		}
	}
	if len(args) >= 3 && args[0] == "add" {
		warp, err := parseWarpAddArgs(args[1:])
		if err != nil {
			return err
		}
		return warps.Save(root, warp)
	}
	if len(args) == 1 && args[0] == "add" {
		warp, err := promptWarpCreation()
		if err != nil {
			return err
		}
		return warps.Save(root, warp)
	}
	if len(args) == 2 && args[0] == "show" {
		item, err := warps.Load(root, args[1])
		if err != nil {
			return err
		}
		data, err := yaml.Marshal(item)
		if err != nil {
			return err
		}
		logger.Print(styles.Title.Render(item.Title))
		logger.Print(string(data))
		return nil
	}
	if len(args) == 2 && args[0] == "remove" {
		return warps.Delete(root, args[1])
	}
	return fmt.Errorf("usage: norn warps <list|add|assign|assignment|show|remove>")
}

func runWarpRuntimeView(root string) error {
	views, err := spindle.BuildWarpViews(root)
	if err != nil {
		return err
	}
	logger.Print(styles.Title.Render("Warp Runtime View"))
	if len(views) == 0 {
		logger.Print(styles.KV("active", "none"))
		return nil
	}
	for _, view := range views {
		summary := fmt.Sprintf("%s [%s]", view.Warp.Title, view.Warp.Status)
		logger.Print(styles.KV(view.Warp.ID, summary))
		if len(view.Weaves) > 0 {
			for _, item := range view.Weaves {
				logger.Print(styles.KV("  weave", fmt.Sprintf("%s (%s) owner=%s state=%s", item.ID, view.Warp.ID, item.Owner, item.State)))
			}
		}
		if len(view.Threads) > 0 {
			for _, item := range view.Threads {
				logger.Print(styles.KV("  thread", fmt.Sprintf("%s (%s) owner=%s state=%s", item.ID, view.Warp.ID, item.Owner, item.State)))
			}
		}
	}
	return nil
}

func parseWarpAddArgs(args []string) (norn.Warp, error) {
	warp := norn.Warp{}
	remaining := make([]string, 0, len(args))
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--id="):
			warp.ID = strings.TrimPrefix(arg, "--id=")
		case strings.HasPrefix(arg, "--status="):
			warp.Status = strings.TrimPrefix(arg, "--status=")
		case strings.HasPrefix(arg, "--owner="):
			warp.Owner = strings.TrimPrefix(arg, "--owner=")
		case strings.HasPrefix(arg, "--root="):
			warp.Root = strings.TrimPrefix(arg, "--root=")
		case strings.HasPrefix(arg, "--branch="):
			warp.Branch = strings.TrimPrefix(arg, "--branch=")
		case strings.HasPrefix(arg, "--weaves="):
			warp.WeaveIDs = splitCSV(strings.TrimPrefix(arg, "--weaves="))
		case strings.HasPrefix(arg, "--threads="):
			warp.ThreadIDs = splitCSV(strings.TrimPrefix(arg, "--threads="))
		case strings.HasPrefix(arg, "--notes="):
			warp.Notes = strings.TrimPrefix(arg, "--notes=")
		default:
			remaining = append(remaining, arg)
		}
	}
	if len(remaining) < 2 {
		return norn.Warp{}, fmt.Errorf("usage: norn warps add [--id=...] [--status=...] [--owner=...] [--root=...] [--branch=...] [--weaves=a,b] [--threads=a,b] [--notes=...] <title> <summary>")
	}
	warp.Title = remaining[0]
	warp.Summary = strings.Join(remaining[1:], " ")
	if strings.TrimSpace(warp.Notes) == "" {
		warp.Notes = warps.DefaultNotes(warp.Summary)
	}
	return warp, nil
}

func parseWarpAssignArgs(args []string) (norn.RuntimeAssignment, error) {
	assignment := norn.RuntimeAssignment{}
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "--kind="):
			assignment.Kind = strings.TrimPrefix(arg, "--kind=")
		case strings.HasPrefix(arg, "--id="):
			assignment.ID = strings.TrimPrefix(arg, "--id=")
		case strings.HasPrefix(arg, "--warp="):
			assignment.WarpID = strings.TrimPrefix(arg, "--warp=")
		case strings.HasPrefix(arg, "--owner="):
			assignment.Owner = strings.TrimPrefix(arg, "--owner=")
		case strings.HasPrefix(arg, "--state="):
			assignment.State = strings.TrimPrefix(arg, "--state=")
		case strings.HasPrefix(arg, "--notes="):
			assignment.Notes = strings.TrimPrefix(arg, "--notes=")
		default:
			return norn.RuntimeAssignment{}, fmt.Errorf("unknown assign argument: %s", arg)
		}
	}
	if assignment.Kind == "" || assignment.ID == "" || assignment.WarpID == "" {
		return norn.RuntimeAssignment{}, fmt.Errorf("usage: norn warps assign --kind=weave|thread --id=<artifact-id> --warp=<warp-id> [--owner=...] [--state=...] [--notes=...]")
	}
	if assignment.Kind != "weave" && assignment.Kind != "thread" {
		return norn.RuntimeAssignment{}, fmt.Errorf("invalid assignment kind %q; expected weave or thread", assignment.Kind)
	}
	if assignment.State == "" {
		assignment.State = "active"
	}
	return assignment, nil
}

func promptWarpAssignment(root string) (norn.RuntimeAssignment, error) {
	warpsList, err := warps.List(root)
	if err != nil {
		return norn.RuntimeAssignment{}, err
	}
	if len(warpsList) == 0 {
		return norn.RuntimeAssignment{}, fmt.Errorf("no warps available; create a warp first")
	}
	warpOptions := make([]huh.Option[string], 0, len(warpsList))
	selectedWarp := warpsList[0].ID
	for _, item := range warpsList {
		warpOptions = append(warpOptions, huh.NewOption(fmt.Sprintf("%s (%s)", item.Title, item.ID), item.ID))
	}
	kind := "thread"
	id := ""
	owner := ""
	state := "active"
	notes := ""
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Kind").Options(
				huh.NewOption("Thread", "thread"),
				huh.NewOption("Weave", "weave"),
			).Value(&kind),
		),
		huh.NewGroup(huh.NewInput().Title("Artifact ID").Value(&id)),
		huh.NewGroup(huh.NewSelect[string]().Title("Warp").Options(warpOptions...).Value(&selectedWarp)),
		huh.NewGroup(huh.NewInput().Title("Owner").Value(&owner)),
		huh.NewGroup(
			huh.NewSelect[string]().Title("State").Options(
				huh.NewOption("Active", "active"),
				huh.NewOption("Review", "review"),
				huh.NewOption("Blocked", "blocked"),
				huh.NewOption("Done", "done"),
			).Value(&state),
		),
		huh.NewGroup(huh.NewText().Title("Notes").Value(&notes)),
	)
	if err := form.Run(); err != nil {
		return norn.RuntimeAssignment{}, err
	}
	preview := fmt.Sprintf("Kind: %s\nID: %s\nWarp: %s\nOwner: %s\nState: %s\nNotes: %s", kind, id, selectedWarp, owner, state, notes)
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Preview").Description(preview), huh.NewConfirm().Title("Save runtime assignment?").Value(&confirmed)))
	if err := confirm.Run(); err != nil {
		return norn.RuntimeAssignment{}, err
	}
	if !confirmed {
		return norn.RuntimeAssignment{}, fmt.Errorf("warp assignment cancelled")
	}
	return norn.RuntimeAssignment{Kind: kind, ID: id, WarpID: selectedWarp, Owner: owner, State: state, Notes: notes}, nil
}

func promptWarpCreation() (norn.Warp, error) {
	title := ""
	id := ""
	summary := ""
	status := "active"
	owner := ""
	root := ""
	branch := ""
	weavesValue := ""
	threadsValue := ""
	notes := ""
	form := huh.NewForm(
		huh.NewGroup(huh.NewInput().Title("Title").Value(&title)),
		huh.NewGroup(huh.NewInput().Title("ID").Description("Leave empty to derive from title").Value(&id)),
		huh.NewGroup(huh.NewText().Title("Summary").Value(&summary)),
		huh.NewGroup(
			huh.NewSelect[string]().Title("Status").Options(
				huh.NewOption("Active", "active"),
				huh.NewOption("Paused", "paused"),
				huh.NewOption("Review", "review"),
				huh.NewOption("Done", "done"),
			).Value(&status),
		),
		huh.NewGroup(huh.NewInput().Title("Owner").Value(&owner)),
		huh.NewGroup(huh.NewInput().Title("Root path").Value(&root)),
		huh.NewGroup(huh.NewInput().Title("Branch").Value(&branch)),
		huh.NewGroup(huh.NewInput().Title("Weaves").Description("Comma-separated weave ids").Value(&weavesValue)),
		huh.NewGroup(huh.NewInput().Title("Threads").Description("Comma-separated thread ids").Value(&threadsValue)),
		huh.NewGroup(huh.NewText().Title("Notes").Description("Blank uses a default runtime note scaffold").Value(&notes)),
	)
	if err := form.Run(); err != nil {
		return norn.Warp{}, err
	}
	if id == "" {
		id = slug(title)
	}
	if strings.TrimSpace(notes) == "" {
		notes = warps.DefaultNotes(summary)
	}
	preview := fmt.Sprintf("ID: %s\nStatus: %s\nOwner: %s\nRoot: %s\nBranch: %s\nWeaves: %s\nThreads: %s\n\n%s", id, status, owner, root, branch, weavesValue, threadsValue, notes)
	confirmed := true
	confirm := huh.NewForm(huh.NewGroup(huh.NewNote().Title("Preview").Description(preview), huh.NewConfirm().Title("Create warp with this content?").Value(&confirmed)))
	if err := confirm.Run(); err != nil {
		return norn.Warp{}, err
	}
	if !confirmed {
		return norn.Warp{}, fmt.Errorf("warp creation cancelled")
	}
	return norn.Warp{ID: id, Title: title, Summary: summary, Status: status, Owner: owner, Root: root, Branch: branch, WeaveIDs: splitCSV(weavesValue), ThreadIDs: splitCSV(threadsValue), Notes: notes}, nil
}

func warpsHelp() HelpTopic {
	return HelpTopic{
		Name:        "norn warps",
		Description: "Manage runtime warp lanes",
		Usage:       "norn warps <command> [flags]",
		Commands: []CommandHelp{
			{
				Name:        "list",
				Description: "List warp lanes",
				Usage:       "norn warps list [--view=runtime]",
				Flags: []FlagHelp{
					{Name: "--view=runtime", Description: "Show runtime ownership view"},
				},
				Examples: []string{
					"norn warps list",
					"norn warps list --view=runtime",
				},
			},
			{
				Name:        "add",
				Description: "Add a new warp lane",
				Usage:       "norn warps add [--id=...] [--status=...] [--owner=...] [--root=...] [--branch=...] [--weaves=a,b] [--threads=a,b] [--notes=...] <title> <summary>",
				Flags: []FlagHelp{
					{Name: "--id=<id>", Description: "Warp ID (derived from title if omitted)"},
					{Name: "--status=<status>", Description: "Warp status"},
					{Name: "--owner=<owner>", Description: "Warp owner"},
					{Name: "--root=<path>", Description: "Root path"},
					{Name: "--branch=<branch>", Description: "Git branch"},
					{Name: "--weaves=a,b", Description: "Associated weave IDs"},
					{Name: "--threads=a,b", Description: "Associated thread IDs"},
					{Name: "--notes=<text>", Description: "Warp notes"},
				},
				Examples: []string{
					"norn warps add \"API Warp\" \"Runtime lane for API\"",
					"norn warps add --status=active --owner=marcus \"API Warp\" \"Runtime lane for API\"",
				},
			},
			{
				Name:        "assign",
				Description: "Assign a weave or thread to a warp",
				Usage:       "norn warps assign --kind=weave|thread --id=<artifact-id> --warp=<warp-id> [--owner=...] [--state=...] [--notes=...]",
				Flags: []FlagHelp{
					{Name: "--kind=weave|thread", Description: "Artifact kind"},
					{Name: "--id=<id>", Description: "Artifact ID"},
					{Name: "--warp=<warp-id>", Description: "Target warp ID"},
					{Name: "--owner=<owner>", Description: "Assignment owner"},
					{Name: "--state=<state>", Description: "Assignment state"},
					{Name: "--notes=<text>", Description: "Assignment notes"},
				},
			},
			{
				Name:        "assignment",
				Description: "Show or remove runtime assignments",
				Usage:       "norn warps assignment <show|remove> <weave|thread> <id>",
			},
			{
				Name:        "show",
				Description: "Show warp details",
				Usage:       "norn warps show <warp-id>",
			},
			{
				Name:        "remove",
				Description: "Remove a warp",
				Usage:       "norn warps remove <warp-id>",
			},
		},
	}
}
