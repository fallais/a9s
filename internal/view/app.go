package view

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"a9s/internal/client"
	"a9s/internal/resources"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App represents the main application
type App struct {
	app       *tview.Application
	pages     *tview.Pages
	table     *tview.Table
	menu      *tview.Flex
	menuList  *tview.List
	menuInput *tview.InputField
	status    *tview.TextView
	header    *tview.TextView
	client    *client.Client
	registry  *resources.Registry
	current   resources.Resource
	ctx       context.Context

	// Resource keys for menu filtering
	resourceKeys []string

	// Auto-refresh
	autoRefresh   bool
	refreshTicker *time.Ticker
	stopRefresh   chan struct{}
	refreshMu     sync.Mutex
}

// Default refresh interval for auto-refresh
const defaultRefreshInterval = 10 * time.Second

// New creates a new App instance
func New(ctx context.Context, c *client.Client) *App {
	a := &App{
		app:         tview.NewApplication(),
		pages:       tview.NewPages(),
		registry:    resources.DefaultRegistry(),
		client:      c,
		ctx:         ctx,
		autoRefresh: true,
		stopRefresh: make(chan struct{}),
	}

	a.setupUI()
	return a
}

// setupUI initializes all UI components
func (a *App) setupUI() {
	// Header
	a.header = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	a.updateHeader()

	// Resource table
	a.table = tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false).
		SetFixed(1, 0)
	a.table.SetBorder(true).SetTitle(" Resources ")

	// Status bar
	a.status = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	a.updateStatus("Press ':' to open menu, 'p' for profile, 'r' for region, 'q' to quit")

	// Resource menu with search
	a.setupResourceMenu()

	// Main layout
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.header, 3, 0, false).
		AddItem(a.table, 0, 1, true).
		AddItem(a.status, 1, 0, false)

	a.pages.AddPage("main", mainFlex, true, true)
	a.pages.AddPage("menu", a.createModal(a.menu, 40, 15), true, false)

	// Key bindings
	a.setupKeyBindings()
}

// createModal creates a centered modal with the given content
func (a *App) createModal(content tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(content, height, 0, true).
			AddItem(nil, 0, 1, false), width, 0, true).
		AddItem(nil, 0, 1, false)
}

// setupResourceMenu creates the resource menu with search functionality
func (a *App) setupResourceMenu() {
	// Get and sort resource keys
	a.resourceKeys = a.registry.List()
	sort.Strings(a.resourceKeys)

	// Create search input field
	a.menuInput = tview.NewInputField().
		SetLabel("Search: ").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray)

	// Create resource list
	a.menuList = tview.NewList().
		SetSelectedBackgroundColor(tcell.ColorDarkCyan).
		SetMainTextColor(tcell.ColorWhite).
		SetHighlightFullLine(true).
		ShowSecondaryText(false)

	// Populate initial list
	a.populateMenuList("")

	// Handle search input changes
	a.menuInput.SetChangedFunc(func(text string) {
		a.populateMenuList(text)
	})

	// Handle input field key events
	a.menuInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyDown, tcell.KeyTab:
			a.app.SetFocus(a.menuList)
			return nil
		case tcell.KeyEnter:
			// Select first item if list has items
			if a.menuList.GetItemCount() > 0 {
				a.menuList.SetCurrentItem(0)
				mainText, _ := a.menuList.GetItemText(0)
				a.selectResource(mainText)
			}
			return nil
		case tcell.KeyEscape:
			a.closeMenu()
			return nil
		}
		return event
	})

	// Handle list key events
	a.menuList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			if a.menuList.GetCurrentItem() == 0 {
				a.app.SetFocus(a.menuInput)
				return nil
			}
		case tcell.KeyEscape:
			a.closeMenu()
			return nil
		case tcell.KeyRune:
			// If typing, focus on input and pass the key
			a.app.SetFocus(a.menuInput)
			a.menuInput.SetText(a.menuInput.GetText() + string(event.Rune()))
			return nil
		}
		return event
	})

	// Create menu layout
	a.menu = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(a.menuInput, 1, 0, true).
		AddItem(a.menuList, 0, 1, false)
	a.menu.SetBorder(true).SetTitle(" Select Resource (Esc to close) ")
}

// populateMenuList populates the menu list based on search filter
func (a *App) populateMenuList(filter string) {
	a.menuList.Clear()
	filter = strings.ToLower(filter)

	for _, key := range a.resourceKeys {
		if filter == "" || strings.Contains(strings.ToLower(key), filter) {
			k := key // capture for closure
			a.menuList.AddItem(key, "", 0, func() {
				a.selectResource(k)
			})
		}
	}
}

// closeMenu closes the resource menu and returns to main view
func (a *App) closeMenu() {
	a.menuInput.SetText("")
	a.populateMenuList("")
	a.pages.SwitchToPage("main")
	a.app.SetFocus(a.table)
}

// setupKeyBindings configures global key bindings
func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Global key bindings
		switch event.Key() {
		case tcell.KeyEscape:
			if a.pages.HasPage("confirm") {
				name, _ := a.pages.GetFrontPage()
				if name == "confirm" {
					a.pages.RemovePage("confirm")
					a.pages.SwitchToPage("main")
					a.app.SetFocus(a.table)
					return nil
				}
			}
			if a.pages.HasPage("profile") || a.pages.HasPage("region") {
				name, _ := a.pages.GetFrontPage()
				if name == "profile" {
					a.pages.RemovePage("profile")
					a.pages.SwitchToPage("main")
					a.app.SetFocus(a.table)
					return nil
				}
				if name == "region" {
					a.pages.RemovePage("region")
					a.pages.SwitchToPage("main")
					a.app.SetFocus(a.table)
					return nil
				}
			}
		case tcell.KeyRune:
			// Only process these keys when on main page
			name, _ := a.pages.GetFrontPage()
			if name != "main" {
				return event
			}
			switch event.Rune() {
			case ':':
				a.pages.SwitchToPage("menu")
				a.app.SetFocus(a.menuInput)
				return nil
			case 'q':
				a.app.Stop()
				return nil
			case 'f':
				// Refresh current resource
				if a.current != nil {
					a.refreshResource()
				}
				return nil
			case 'a':
				// Toggle auto-refresh
				a.toggleAutoRefresh()
				return nil
			case '1':
				a.selectResource("ec2")
				return nil
			case '2':
				a.selectResource("s3")
				return nil
			case 's':
				// Stop EC2 instance
				a.handleEC2Action("stop")
				return nil
			case 'S':
				// Start EC2 instance
				a.handleEC2Action("start")
				return nil
			case 'R':
				// Restart EC2 instance
				a.handleEC2Action("restart")
				return nil
			case 'c':
				// Create S3 bucket
				a.handleS3Create()
				return nil
			case 'd':
				// Delete S3 bucket
				a.handleS3Delete()
				return nil
			case 'e':
				// Empty S3 bucket
				a.handleS3Empty()
				return nil
			case 'p':
				// Switch AWS profile
				a.showProfileInput()
				return nil
			case 'r':
				// Switch AWS region
				a.showRegionInput()
				return nil
			}
		}
		return event
	})
}

// selectResource switches to the specified resource view
func (a *App) selectResource(key string) {
	res, ok := a.registry.Get(key)
	if !ok {
		a.updateStatus(fmt.Sprintf("[red]Unknown resource: %s", key))
		return
	}

	a.current = res
	// Clear search and close menu
	a.menuInput.SetText("")
	a.populateMenuList("")
	a.pages.SwitchToPage("main")
	a.app.SetFocus(a.table)
	a.refreshResource()
	a.startAutoRefresh()
}

// refreshResource fetches and displays the current resource
func (a *App) refreshResource() {
	if a.current == nil {
		return
	}

	a.updateStatus("[yellow]Loading...")
	a.table.Clear()

	go func() {
		err := a.current.Fetch(a.ctx, a.client)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Error: %v", err))
				return
			}

			a.renderTable()
			rows := a.current.Rows()
			autoStatus := "[gray]auto:off"
			if a.autoRefresh {
				autoStatus = "[green]auto:on"
			}
			resourceHelp := ""
			if _, ok := a.current.(*resources.EC2Instances); ok {
				resourceHelp = " | s: stop | S: start | R: restart"
			}
			if _, ok := a.current.(*resources.S3Buckets); ok {
				resourceHelp = " | c: create | d: delete | e: empty"
			}
			a.updateStatus(fmt.Sprintf("%s | [green]%s: %d items | [white]f: refresh | a: auto | p: profile | r: region | :: menu | q: quit%s",
				autoStatus, a.current.Name(), len(rows), resourceHelp))
		})
	}()
}

// renderTable renders the current resource data in the table
func (a *App) renderTable() {
	a.table.Clear()

	if a.current == nil {
		return
	}

	// Header row
	columns := a.current.Columns()
	for i, col := range columns {
		cell := tview.NewTableCell(col.Name).
			SetTextColor(tcell.ColorYellow).
			SetSelectable(false).
			SetExpansion(1)
		a.table.SetCell(0, i, cell)
	}

	// Data rows
	rows := a.current.Rows()
	for i, row := range rows {
		for j, value := range row {
			cell := tview.NewTableCell(value).
				SetTextColor(tcell.ColorWhite).
				SetExpansion(1)
			a.table.SetCell(i+1, j, cell)
		}
	}

	a.table.SetTitle(fmt.Sprintf(" %s ", a.current.Name()))
	a.table.ScrollToBeginning()
}

// updateHeader updates the header text
func (a *App) updateHeader() {
	region := "not configured"
	profile := "not configured"
	if a.client != nil {
		if a.client.Region() != "" {
			region = a.client.Region()
		}
		if a.client.Profile() != "" {
			profile = a.client.Profile()
		}
	}
	a.header.SetText(fmt.Sprintf("[::b]a9s[-:-:-] - AWS Resource Browser\n[gray]Region: %s | Profile: %s", region, profile))
}

// updateStatus updates the status bar text
func (a *App) updateStatus(text string) {
	a.status.SetText(" " + text)
}

// startAutoRefresh starts the background auto-refresh ticker
func (a *App) startAutoRefresh() {
	a.refreshMu.Lock()
	defer a.refreshMu.Unlock()

	// Stop existing ticker if any
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
	}

	if !a.autoRefresh || a.current == nil {
		return
	}

	a.refreshTicker = time.NewTicker(defaultRefreshInterval)

	go func() {
		for {
			select {
			case <-a.refreshTicker.C:
				if a.autoRefresh && a.current != nil {
					a.refreshResource()
				}
			case <-a.stopRefresh:
				return
			case <-a.ctx.Done():
				return
			}
		}
	}()
}

// stopAutoRefresh stops the background auto-refresh ticker
func (a *App) stopAutoRefresh() {
	a.refreshMu.Lock()
	defer a.refreshMu.Unlock()

	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
		a.refreshTicker = nil
	}
}

// toggleAutoRefresh toggles the auto-refresh feature
func (a *App) toggleAutoRefresh() {
	a.autoRefresh = !a.autoRefresh

	if a.autoRefresh {
		a.startAutoRefresh()
		a.updateStatusWithAutoRefresh("[green]Auto-refresh enabled")
	} else {
		a.stopAutoRefresh()
		a.updateStatusWithAutoRefresh("[yellow]Auto-refresh disabled")
	}
}

// updateStatusWithAutoRefresh updates status showing auto-refresh state
func (a *App) updateStatusWithAutoRefresh(prefix string) {
	autoStatus := "[gray]auto:off"
	if a.autoRefresh {
		autoStatus = "[green]auto:on"
	}

	if a.current != nil {
		rows := a.current.Rows()
		a.updateStatus(fmt.Sprintf("%s | %s: %d items | [white]f: refresh | a: auto | p: profile | r: region | :: menu | q: quit",
			autoStatus, a.current.Name(), len(rows)))
	} else {
		a.updateStatus(fmt.Sprintf("%s | [white]%s", autoStatus, prefix))
	}
}

// Run starts the application
func (a *App) Run() error {
	defer func() {
		close(a.stopRefresh)
		a.stopAutoRefresh()
	}()
	return a.app.SetRoot(a.pages, true).EnableMouse(true).Run()
}

// showProfileInput displays an input dialog for switching AWS profile
func (a *App) showProfileInput() {
	input := tview.NewInputField().
		SetLabel("Profile: ").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetText(a.client.Profile())

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			profile := input.GetText()
			if profile != "" {
				a.switchProfile(profile)
			}
		}
		a.pages.RemovePage("profile")
		a.pages.SwitchToPage("main")
		a.app.SetFocus(a.table)
	})

	form := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true)
	form.SetBorder(true).SetTitle(" Switch AWS Profile (Enter to confirm, Esc to cancel) ")

	modal := a.createModal(form, 50, 3)
	a.pages.AddPage("profile", modal, true, true)
	a.app.SetFocus(input)
}

// showRegionInput displays an input dialog for switching AWS region
func (a *App) showRegionInput() {
	input := tview.NewInputField().
		SetLabel("Region: ").
		SetFieldWidth(30).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetText(a.client.Region())

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			region := input.GetText()
			if region != "" {
				a.switchRegion(region)
			}
		}
		a.pages.RemovePage("region")
		a.pages.SwitchToPage("main")
		a.app.SetFocus(a.table)
	})

	form := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true)
	form.SetBorder(true).SetTitle(" Switch AWS Region (Enter to confirm, Esc to cancel) ")

	modal := a.createModal(form, 50, 3)
	a.pages.AddPage("region", modal, true, true)
	a.app.SetFocus(input)
}

// switchProfile changes the AWS profile and refreshes the view
func (a *App) switchProfile(profile string) {
	a.updateStatus(fmt.Sprintf("[yellow]Switching to profile: %s...", profile))

	go func() {
		err := a.client.SetProfile(a.ctx, profile)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Failed to switch profile: %v", err))
				return
			}

			a.updateHeader()
			a.updateStatus(fmt.Sprintf("[green]Switched to profile: %s", profile))

			// Refresh current resource if any
			if a.current != nil {
				a.refreshResource()
			}
		})
	}()
}

// switchRegion changes the AWS region and refreshes the view
func (a *App) switchRegion(region string) {
	a.updateStatus(fmt.Sprintf("[yellow]Switching to region: %s...", region))

	go func() {
		err := a.client.SetRegion(a.ctx, region)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Failed to switch region: %v", err))
				return
			}

			a.updateHeader()
			a.updateStatus(fmt.Sprintf("[green]Switched to region: %s", region))

			// Refresh current resource if any
			if a.current != nil {
				a.refreshResource()
			}
		})
	}()
}

// handleEC2Action handles EC2 instance actions (start, stop, restart)
func (a *App) handleEC2Action(action string) {
	// Check if we're viewing EC2 instances
	ec2Res, ok := a.current.(*resources.EC2Instances)
	if !ok {
		a.updateStatus("[yellow]EC2 actions only available when viewing EC2 instances")
		return
	}

	// Get selected row (subtract 1 for header row)
	row, _ := a.table.GetSelection()
	if row <= 0 {
		a.updateStatus("[yellow]Please select an instance first")
		return
	}

	instanceID := ec2Res.GetID(row - 1)
	if instanceID == "" {
		a.updateStatus("[red]Could not get instance ID")
		return
	}

	// Show confirmation dialog
	a.showEC2ActionConfirm(action, instanceID, ec2Res)
}

// showEC2ActionConfirm displays a confirmation dialog for EC2 actions
func (a *App) showEC2ActionConfirm(action, instanceID string, ec2Res *resources.EC2Instances) {
	actionColors := map[string]string{
		"start":   "green",
		"stop":    "red",
		"restart": "yellow",
	}
	color := actionColors[action]

	modal := tview.NewModal().
		SetText(fmt.Sprintf("[%s]%s[-] instance [white]%s[-]?", color, action, instanceID)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm")
			a.pages.SwitchToPage("main")
			a.app.SetFocus(a.table)

			if buttonLabel == "Yes" {
				a.executeEC2Action(action, instanceID, ec2Res)
			}
		})

	a.pages.AddPage("confirm", modal, true, true)
	a.app.SetFocus(modal)
}

// executeEC2Action executes the EC2 action
func (a *App) executeEC2Action(action, instanceID string, ec2Res *resources.EC2Instances) {
	a.updateStatus(fmt.Sprintf("[yellow]%sing instance %s...", action, instanceID))

	go func() {
		var err error
		switch action {
		case "start":
			err = ec2Res.StartInstance(a.ctx, a.client, instanceID)
		case "stop":
			err = ec2Res.StopInstance(a.ctx, a.client, instanceID)
		case "restart":
			err = ec2Res.RestartInstance(a.ctx, a.client, instanceID)
		}

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Failed to %s instance: %v", action, err))
				return
			}

			a.updateStatus(fmt.Sprintf("[green]Successfully initiated %s for %s", action, instanceID))
			// Refresh to show updated state
			time.Sleep(2 * time.Second)
			a.refreshResource()
		})
	}()
}

// handleS3Create handles S3 bucket creation
func (a *App) handleS3Create() {
	// Check if we're viewing S3 buckets
	_, ok := a.current.(*resources.S3Buckets)
	if !ok {
		a.updateStatus("[yellow]S3 create only available when viewing S3 buckets")
		return
	}

	a.showS3CreateInput()
}

// showS3CreateInput displays an input dialog for creating an S3 bucket
func (a *App) showS3CreateInput() {
	input := tview.NewInputField().
		SetLabel("Bucket Name: ").
		SetFieldWidth(40).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray)

	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			bucketName := input.GetText()
			if bucketName != "" {
				a.pages.RemovePage("s3create")
				a.showS3CreateConfirm(bucketName)
				return
			}
		}
		a.pages.RemovePage("s3create")
		a.pages.SwitchToPage("main")
		a.app.SetFocus(a.table)
	})

	form := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(input, 1, 0, true)
	form.SetBorder(true).SetTitle(" Create S3 Bucket (Enter to confirm, Esc to cancel) ")

	modal := a.createModal(form, 60, 3)
	a.pages.AddPage("s3create", modal, true, true)
	a.app.SetFocus(input)
}

// showS3CreateConfirm displays a confirmation dialog for S3 bucket creation
func (a *App) showS3CreateConfirm(bucketName string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Create bucket [green]%s[-] in region [yellow]%s[-]?", bucketName, a.client.Region())).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm")
			a.pages.SwitchToPage("main")
			a.app.SetFocus(a.table)

			if buttonLabel == "Yes" {
				a.executeS3Create(bucketName)
			}
		})

	a.pages.AddPage("confirm", modal, true, true)
	a.app.SetFocus(modal)
}

// executeS3Create executes the S3 bucket creation
func (a *App) executeS3Create(bucketName string) {
	s3Res, ok := a.current.(*resources.S3Buckets)
	if !ok {
		a.updateStatus("[red]S3 resource not available")
		return
	}

	a.updateStatus(fmt.Sprintf("[yellow]Creating bucket %s...", bucketName))

	go func() {
		err := s3Res.CreateBucket(a.ctx, a.client, bucketName)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Failed to create bucket: %v", err))
				return
			}

			a.updateStatus(fmt.Sprintf("[green]Successfully created bucket %s", bucketName))
			// Refresh to show the new bucket
			time.Sleep(1 * time.Second)
			a.refreshResource()
		})
	}()
}

// handleS3Delete handles S3 bucket deletion
func (a *App) handleS3Delete() {
	// Check if we're viewing S3 buckets
	s3Res, ok := a.current.(*resources.S3Buckets)
	if !ok {
		a.updateStatus("[yellow]S3 delete only available when viewing S3 buckets")
		return
	}

	// Get selected row (subtract 1 for header row)
	row, _ := a.table.GetSelection()
	if row <= 0 {
		a.updateStatus("[yellow]Please select a bucket first")
		return
	}

	bucketName := s3Res.GetID(row - 1)
	if bucketName == "" {
		a.updateStatus("[red]Could not get bucket name")
		return
	}

	a.showS3DeleteConfirm(bucketName)
}

// showS3DeleteConfirm displays a confirmation dialog for S3 bucket deletion
func (a *App) showS3DeleteConfirm(bucketName string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[red]Delete[-] bucket [white]%s[-]?\n\n[yellow]Warning: Bucket must be empty!", bucketName)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm")
			a.pages.SwitchToPage("main")
			a.app.SetFocus(a.table)

			if buttonLabel == "Yes" {
				a.executeS3Delete(bucketName)
			}
		})

	a.pages.AddPage("confirm", modal, true, true)
	a.app.SetFocus(modal)
}

// executeS3Delete executes the S3 bucket deletion
func (a *App) executeS3Delete(bucketName string) {
	s3Res, ok := a.current.(*resources.S3Buckets)
	if !ok {
		a.updateStatus("[red]S3 resource not available")
		return
	}

	a.updateStatus(fmt.Sprintf("[yellow]Deleting bucket %s...", bucketName))

	go func() {
		err := s3Res.DeleteBucket(a.ctx, a.client, bucketName)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Failed to delete bucket: %v", err))
				return
			}

			a.updateStatus(fmt.Sprintf("[green]Successfully deleted bucket %s", bucketName))
			// Refresh to update the list
			time.Sleep(1 * time.Second)
			a.refreshResource()
		})
	}()
}

// handleS3Empty handles S3 bucket emptying
func (a *App) handleS3Empty() {
	// Check if we're viewing S3 buckets
	s3Res, ok := a.current.(*resources.S3Buckets)
	if !ok {
		a.updateStatus("[yellow]S3 empty only available when viewing S3 buckets")
		return
	}

	// Get selected row (subtract 1 for header row)
	row, _ := a.table.GetSelection()
	if row <= 0 {
		a.updateStatus("[yellow]Please select a bucket first")
		return
	}

	bucketName := s3Res.GetID(row - 1)
	if bucketName == "" {
		a.updateStatus("[red]Could not get bucket name")
		return
	}

	a.showS3EmptyConfirm(bucketName)
}

// showS3EmptyConfirm displays a confirmation dialog for S3 bucket emptying
func (a *App) showS3EmptyConfirm(bucketName string) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("[red]Empty[-] bucket [white]%s[-]?\n\n[yellow]WARNING: This will permanently delete ALL objects!\nThis action cannot be undone!", bucketName)).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.pages.RemovePage("confirm")
			a.pages.SwitchToPage("main")
			a.app.SetFocus(a.table)

			if buttonLabel == "Yes" {
				a.executeS3Empty(bucketName)
			}
		})

	a.pages.AddPage("confirm", modal, true, true)
	a.app.SetFocus(modal)
}

// executeS3Empty executes the S3 bucket emptying
func (a *App) executeS3Empty(bucketName string) {
	s3Res, ok := a.current.(*resources.S3Buckets)
	if !ok {
		a.updateStatus("[red]S3 resource not available")
		return
	}

	a.updateStatus(fmt.Sprintf("[yellow]Emptying bucket %s... (this may take a while)", bucketName))

	go func() {
		err := s3Res.EmptyBucket(a.ctx, a.client, bucketName)

		a.app.QueueUpdateDraw(func() {
			if err != nil {
				a.updateStatus(fmt.Sprintf("[red]Failed to empty bucket: %v", err))
				return
			}

			a.updateStatus(fmt.Sprintf("[green]Successfully emptied bucket %s", bucketName))
		})
	}()
}
