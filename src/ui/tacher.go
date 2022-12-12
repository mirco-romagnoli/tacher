package ui

import (
	"fmt"
	"os"
	"path"
	"sort"
	"tacher/src/client"
	"tacher/src/model"
	"tacher/src/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const PAGE_INTRO = "Intro"
const PAGE_PRJ_META = "Project Metadata"
const PAGE_DEPENDENCIES = "Dependencies"
const PAGE_PRJ_PATH = "Project Path"
const INITIAL_PAGE = PAGE_INTRO

func RunUI(group, artifact, name, description, pkg string) error {
	// init app's state and retrieve options from Spring initializer
	state := new(model.AppState)
	err := client.GetOptions(state)
	if err != nil {
		return err
	}

	// init data from parameters
	data := new(model.AppData)
	data.Group = utils.NonNullOrElse(group, state.DefaultGroupId)
	data.Artifact = utils.NonNullOrElse(artifact, state.DefaultArtifactId)
	data.Name = utils.NonNullOrElse(artifact, state.DefaultName)
	data.Description = utils.NonNullOrElse(description, state.DefaultDescription)
	data.Pkg = utils.NonNullOrElse(pkg, state.DefaultPackageName)

	// init app gui
	state.App = tview.NewApplication()
	state.Pages = tview.NewPages()
	state.Pages.AddPage(PAGE_INTRO, buildIntroForm(state, data), true, false)
	state.Pages.AddPage(PAGE_PRJ_META, buildProjectMetadataForm(state, data), true, false)
	state.Pages.AddPage(PAGE_DEPENDENCIES, buildDependenciesPage(state, data), true, false)
	state.Pages.AddPage(PAGE_PRJ_PATH, buildProjectPathPage(state, data), true, false)
	state.Pages.SwitchToPage(INITIAL_PAGE)

	// run gui
	if err := state.App.SetRoot(state.Pages, true).SetFocus(state.Pages).Run(); err != nil {
		return err
	}
	return nil
}

func buildIntroForm(state *model.AppState, data *model.AppData) *tview.Form {
	// map values into dropdown options
	buildTools := utils.Map(state.SpringBuildTools, func(st model.ValueWithDesc) string { return st.Name })
	languages := utils.Map(state.Languages, func(l model.Value) string { return l.Name })
	springBootVersions := utils.Map(state.SpringVersions, func(v model.Value) string { return v.Name })

	// build intro form
	form := tview.NewForm().
		AddDropDown("Project", buildTools, state.DefaultSpringBuildTool, func(option string, optionIndex int) { data.SpringBuildTool = state.SpringBuildTools[optionIndex].ID }).
		AddDropDown("Language", languages, state.DefaultLanguage, func(option string, optionIndex int) { data.Language = state.Languages[optionIndex].ID }).
		AddDropDown("Spring Boot", springBootVersions, state.DefaultSpringVersion, func(option string, optionIndex int) { data.SpringBootVersion = state.SpringVersions[optionIndex].ID }).
		AddButton("Next", func() { state.Pages.SwitchToPage(PAGE_PRJ_META) }).
		AddButton("Quit", func() { state.App.Stop() })
	form.SetBorder(true).SetTitle("Project").SetTitleAlign(tview.AlignLeft)
	return form
}

func buildProjectMetadataForm(state *model.AppState, data *model.AppData) *tview.Form {
	// map values into dropdown options
	packagings := utils.Map(state.Packaging, func(p model.Value) string { return p.Name })
	javaVersions := utils.Map(state.JavaVersions, func(v model.Value) string { return v.Name })

	// build project metadata form
	form := tview.NewForm().
		AddInputField("Group", data.Group, 200, nil, func(text string) { data.Group = text }).
		AddInputField("Artifact", data.Artifact, 200, nil, func(text string) { data.Artifact = text }).
		AddInputField("Name", data.Name, 200, nil, func(text string) { data.Name = text }).
		AddInputField("Description", data.Description, 200, nil, func(text string) { data.Description = text }).
		AddInputField("Package name", data.Pkg, 200, nil, func(text string) { data.Pkg = text }).
		AddDropDown("Packaging", packagings, state.DefaultPackaging, func(option string, optionIndex int) { data.Packaging = state.Packaging[optionIndex].ID }).
		AddDropDown("Java", javaVersions, state.DefaultJavaVersion, func(option string, optionIndex int) { data.JavaVersion = state.JavaVersions[optionIndex].ID }).
		AddButton("Next", func() { state.Pages.SwitchToPage(PAGE_DEPENDENCIES) }).
		AddButton("Back", func() { state.Pages.SwitchToPage(PAGE_INTRO) }).
		AddButton("Quit", func() { state.App.Stop() })
	form.SetBorder(true).SetTitle("Project Metadata").SetTitleAlign(tview.AlignLeft)
	return form
}

func buildDependenciesPage(state *model.AppState, data *model.AppData) *tview.Grid {
	grid := tview.NewGrid().
		SetRows(-1, -1, -1, 1).SetColumns(0, 0, 0)

	// init treeview
	root := tview.NewTreeNode(".")
	tree := tview.NewTreeView().SetRoot(root).SetCurrentNode(root)
	tree.SetBorder(true).SetTitle("Dependencies").SetTitleAlign(tview.AlignLeft)

	// set up description area
	description := tview.NewTextView().SetText("Description")
	description.SetBorder(true).SetTitle("Description").SetTitleAlign(tview.AlignLeft)

	// set up selected dependencies area
	selected := tview.NewList()
	selected.SetBorder(true).SetTitle("Selected dependencies").SetTitleAlign(tview.AlignLeft)

	// sort category keys to display them in alphabetical order
	keys := make([]string, 0, len(state.Dependency))
	for k := range state.Dependency {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// set up tree view
	for _, k := range keys {
		v := state.Dependency[k]
		category := tview.NewTreeNode(k)
		category.SetSelectable(false)
		for _, d := range v {
			dependency := tview.NewTreeNode(d.Name)
			dependency.SetReference(d)
			category.AddChild(dependency)
		}
		root.AddChild(category)
	}

	// populate description's text area with selected node description
	tree.SetChangedFunc(func(node *tview.TreeNode) {
		ref, isValueWithDesc := node.GetReference().(model.ValueWithDesc)
		if isValueWithDesc {
			description.SetText(ref.Description)
		} else {
			description.Clear()
		}
	})

	// add or remove dependency from selected list
	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		ref, isValueWithDesc := node.GetReference().(model.ValueWithDesc)
		if !isValueWithDesc {
			return
		}
		if idx := sort.Search(len(data.Dependencies), func(i int) bool { return data.Dependencies[i].ID == ref.ID }); idx < len(data.Dependencies) {
			data.Dependencies = utils.RemoveIndex(data.Dependencies, idx)
			selected.RemoveItem(idx)
			node.SetColor(tcell.ColorWhite)
		} else {
			data.Dependencies = append(data.Dependencies, ref)
			selected.AddItem(ref.Name, ref.Description, '✓', nil)
			node.SetColor(tcell.ColorGreen)
		}
	})

	// add buttons
	buttonGrid := tview.NewGrid().SetRows(0).SetColumns(0, 0, 0).SetGap(0, 1)
	next := tview.NewButton("Next").SetSelectedFunc(func() { state.Pages.SwitchToPage(PAGE_PRJ_PATH) })
	back := tview.NewButton("Back").SetSelectedFunc(func() { state.Pages.SwitchToPage(PAGE_PRJ_META) })
	quit := tview.NewButton("Quit").SetSelectedFunc(func() { state.App.Stop() })
	buttonGrid.AddItem(next, 1, 0, 1, 1, 0, 0, false)
	buttonGrid.AddItem(back, 1, 1, 1, 1, 0, 0, false)
	buttonGrid.AddItem(quit, 1, 2, 1, 1, 0, 0, false)

	// set up focus handling
	primitives := []tview.Primitive{tree, selected, next, back, quit}
	grid.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			cycleFocus(state.App, primitives, false)
		}
		return event
	})

	// add items to the grid
	grid.AddItem(tree, 0, 0, 3, 1, 0, 10, true)
	grid.AddItem(selected, 0, 1, 2, 2, 0, 0, false)
	grid.AddItem(description, 2, 1, 1, 2, 0, 50, false)
	grid.AddItem(buttonGrid, 3, 0, 1, 1, 0, 0, false)
	return grid
}

func buildProjectPathPage(state *model.AppState, data *model.AppData) *tview.Form {
	// get user's home dir
	initialDir, err := os.UserHomeDir()
	if err != nil {
		showError(state, fmt.Errorf("can't get user's home dir: %w", err), nil)
		initialDir = ""
	}
	data.Path = initialDir

	// build project path form
	form := tview.NewForm().
		AddInputField("Project path", initialDir, 200, nil, func(text string) { data.Path = text }).
		AddButton("Next", func() {
			if err := client.Generate(data); err != nil {
				// handle project generation error
				showError(state, err, nil)
			} else {
				// show info message and quit
				showInfo(state, fmt.Sprintf("Project created in \"%s\"", path.Join(data.Path, data.Artifact)), func(buttonIndex int, buttonLabel string) { state.App.Stop() })
			}
		}).
		AddButton("Back", func() { state.Pages.SwitchToPage(PAGE_DEPENDENCIES) }).
		AddButton("Quit", func() { state.App.Stop() })
	form.SetBorder(true).SetTitle("Project").SetTitleAlign(tview.AlignLeft)
	return form
}

// helper that shows an info modal
func showInfo(state *model.AppState, message string, handler func(buttonIndex int, buttonLabel string)) {
	showModal(state, message, tcell.ColorBlue, []string{"Ok"}, handler)
}

// helper that shows an error modal
func showError(state *model.AppState, err error, handler func(buttonIndex int, buttonLabel string)) {
	showModal(state, err.Error(), tcell.ColorRed, []string{"Ok"}, handler)
}

// helper that shows a modal. The handler function is used to set the behaviour when one of the buttons is chosen
func showModal(state *model.AppState, message string, modalColor tcell.Color, buttons []string, handler func(buttonIndex int, buttonLabel string)) {
	// set default handler if none was passed
	if handler == nil {
		handler = func(buttonIndex int, buttonLabel string) {
			state.App.SetRoot(state.Pages, true).SetFocus(state.Pages)
		}
	}

	// build modal
	modal := tview.NewModal().
		SetText(message).
		AddButtons(buttons).
		SetBackgroundColor(modalColor).
		SetDoneFunc(handler)
	state.App.SetRoot(modal, true).SetFocus(modal)
}

// change focus between the given primitives
func cycleFocus(app *tview.Application, elements []tview.Primitive, reverse bool) {
	for i, el := range elements {
		if !el.HasFocus() {
			continue
		}

		if reverse {
			i = i - 1
			if i < 0 {
				i = len(elements) - 1
			}
		} else {
			i = i + 1
			i = i % len(elements)
		}

		app.SetFocus(elements[i])
		return
	}
}
