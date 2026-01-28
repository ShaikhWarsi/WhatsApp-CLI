package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
	"github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

var (
	db             *sql.DB
	activeBorder   = lipgloss.Color("205")
	inactiveBorder = lipgloss.Color("240")
	styleSidebar   = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
	styleChat      = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).Padding(0, 1)
	styleInput     = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(activeBorder).Padding(0, 1)
	styleSelected  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	styleMedia     = lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Italic(true)
)

type model struct {
	client        *whatsmeow.Client
	width, height int
	input         textinput.Model
	conversations map[string][]string
	contacts      []string
	names         map[string]string
	historyLoaded map[string]bool
	historyEnabled bool
	cursor        int
}

type incomingWAMsg struct {
	ChatJID, Sender, Content string
	IsFromMe                 bool
	Timestamp                time.Time
}

func initialModel(client *whatsmeow.Client, historyEnabled bool) model {
	ti := textinput.New()
	ti.Placeholder = "Type a message..."
	ti.Focus()
	ti.CharLimit = 1000
	ti.Width = 50
	os.Mkdir("downloads", 0755)
	
	m := model{
		client:        client,
		input:         ti,
		conversations: make(map[string][]string),
		contacts:      []string{},
		names:         make(map[string]string),
		historyLoaded: make(map[string]bool),
		historyEnabled: historyEnabled,
		cursor:        0,
	}
	m.loadRecentChats()
	if len(m.contacts) > 0 {
		m.loadHistory(m.contacts[0])
	}
	return m
}

func (m *model) loadRecentChats() {
	if !m.historyEnabled { return }
	rows, err := db.Query("SELECT DISTINCT chat_jid FROM cli_messages ORDER BY timestamp DESC LIMIT 20")
	if err != nil { return }
	defer rows.Close()
	for rows.Next() {
		var jid string
		if err := rows.Scan(&jid); err == nil {
			m.contacts = append(m.contacts, jid)
			parsed, _ := types.ParseJID(jid)
			m.names[jid] = resolveName(m.client, parsed)
		}
	}
}

func (m *model) loadHistory(jid string) {
	if !m.historyEnabled || m.historyLoaded[jid] { return }
	
	rows, err := db.Query("SELECT sender, content, is_from_me, timestamp FROM cli_messages WHERE chat_jid = ? ORDER BY timestamp ASC LIMIT 100", jid)
	if err != nil { return }
	defer rows.Close()

	var history []string
	for rows.Next() {
		var sender, content string
		var isFromMe bool
		var ts time.Time
		if err := rows.Scan(&sender, &content, &isFromMe, &ts); err == nil {
			timeStr := ts.Format("15:04")
			colorSender := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(sender)
			if isFromMe { colorSender = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("Me") }
			
			displayContent := content
			if strings.Contains(displayContent, "[Saved") || strings.Contains(displayContent, "CALL") {
				displayContent = styleMedia.Render(displayContent)
			}
			formatted := fmt.Sprintf("[%s] %s: %s", timeStr, colorSender, displayContent)
			history = append(history, formatted)
		}
	}
	
	// Prepend history to existing (live) messages
	m.conversations[jid] = append(history, m.conversations[jid]...)
	m.historyLoaded[jid] = true
}

func saveMessage(chatJID, sender, content string, isFromMe bool, ts time.Time) {
	_, err := db.Exec("INSERT INTO cli_messages (chat_jid, sender, content, is_from_me, timestamp) VALUES (?, ?, ?, ?, ?)",
		chatJID, sender, content, isFromMe, ts)
	if err != nil {
		fmt.Printf("Error saving message: %v\n", err)
	}
}

func (m model) Init() tea.Cmd { return textinput.Blink }

func resolveName(client *whatsmeow.Client, jid types.JID) string {
	if jid.Server == "g.us" {
		info, err := client.GetGroupInfo(context.Background(), jid)
		if err == nil && info != nil { return info.Name }
	}
	contact, err := client.Store.Contacts.GetContact(context.Background(), jid)
	if err == nil && contact.Found {
		if contact.FullName != "" { return contact.FullName }
		if contact.PushName != "" { return contact.PushName }
	}
	return jid.User
}

func downloadMedia(client *whatsmeow.Client, msg *events.Message) string {
	var data []byte
	var err error
	var ext, prefix string
	ctx := context.Background()
	if msg.Message.ImageMessage != nil {
		data, err = client.Download(ctx, msg.Message.ImageMessage)
		ext, prefix = ".jpg", "Image"
	} else if msg.Message.VideoMessage != nil {
		data, err = client.Download(ctx, msg.Message.VideoMessage)
		ext, prefix = ".mp4", "Video"
	} else if msg.Message.DocumentMessage != nil {
		data, err = client.Download(ctx, msg.Message.DocumentMessage)
		ext, prefix = "", "Doc"
	} else { return "" }
	if err != nil { return fmt.Sprintf("[Error %s]", prefix) }
	filename := fmt.Sprintf("%s%s", msg.Info.ID, ext)
	os.WriteFile(filepath.Join("downloads", filename), data, 0644)
	return fmt.Sprintf("[Saved %s: downloads/%s]", prefix, filename)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 34
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.client.Disconnect()
			return m, tea.Quit
		case "ctrl+h":
			m.historyEnabled = !m.historyEnabled
			val := "false"
			if m.historyEnabled { val = "true" }
			db.Exec("UPDATE cli_settings SET value = ? WHERE key = 'history_enabled'", val)
			
			if m.historyEnabled && len(m.contacts) > 0 {
				m.loadHistory(m.contacts[m.cursor])
			}
			return m, nil
		case "up":
			if m.cursor > 0 { 
				m.cursor--
				m.loadHistory(m.contacts[m.cursor])
			}
		case "down":
			if m.cursor < len(m.contacts)-1 { 
				m.cursor++
				m.loadHistory(m.contacts[m.cursor])
			}
		case "enter":
			if len(m.contacts) > 0 { m.sendMessage() }
		}
	case incomingWAMsg:
		if m.historyEnabled {
			saveMessage(msg.ChatJID, msg.Sender, msg.Content, msg.IsFromMe, msg.Timestamp)
		}
		chatID := msg.ChatJID
		exists := false
		for _, c := range m.contacts {
			if c == chatID { exists = true; break }
		}
		if !exists {
			m.contacts = append([]string{chatID}, m.contacts...)
			jid, _ := types.ParseJID(chatID)
			m.names[chatID] = resolveName(m.client, jid)
			sort.Strings(m.contacts)
		}
		if m.conversations[chatID] == nil { m.conversations[chatID] = []string{} }
		senderDisplay := msg.Sender
		if msg.IsFromMe { senderDisplay = "Me" }
		timeStr := msg.Timestamp.Format("15:04")
		colorSender := lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Render(senderDisplay)
		if msg.IsFromMe { colorSender = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("Me") }
		content := msg.Content
		if strings.Contains(content, "[Saved") || strings.Contains(content, "CALL") { content = styleMedia.Render(content) }
		formatted := fmt.Sprintf("[%s] %s: %s", timeStr, colorSender, content)
		m.conversations[chatID] = append(m.conversations[chatID], formatted)
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) sendMessage() {
	text := m.input.Value()
	if text == "" { return }
	targetJIDStr := m.contacts[m.cursor]
	targetJID, _ := types.ParseJID(targetJIDStr)
	m.input.Reset()
	
	ts := time.Now()
	if m.historyEnabled {
		saveMessage(targetJIDStr, "Me", text, true, ts)
	}
	
	go func() {
		msg := &waE2E.Message{Conversation: proto.String(text)}
		m.client.SendMessage(context.Background(), targetJID, msg)
	}()
	formatted := fmt.Sprintf("[%s] %s: %s", ts.Format("15:04"), lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render("Me"), text)
	m.conversations[targetJIDStr] = append(m.conversations[targetJIDStr], formatted)
}

func (m model) View() string {
	var contactList strings.Builder
	historyStatus := "OFF"
	if m.historyEnabled { historyStatus = "ON" }
	contactList.WriteString(fmt.Sprintf("History: %s (ctrl+h)\n", historyStatus))
	contactList.WriteString("----------------------\n")
	contactList.WriteString("Active Chats:\n\n")
	for i, jid := range m.contacts {
		name := m.names[jid]
		if name == "" { name = strings.Split(jid, "@")[0] }
		cursor := " "
		lineStyle := lipgloss.NewStyle()
		if m.cursor == i {
			cursor = ">"
			lineStyle = styleSelected
		}
		runes := []rune(name)
		if len(runes) > 18 { name = string(runes[:15]) + "..." }
		contactList.WriteString(lineStyle.Render(fmt.Sprintf("%s %s", cursor, name)) + "\n")
	}
	totalH := m.height - 2
	chatH := totalH - 3
	if chatH < 0 { chatH = 0 }
	leftPane := styleSidebar.Height(totalH).Width(30).BorderForeground(activeBorder).Render(contactList.String())
	var chatContent string
	if len(m.contacts) > 0 {
		jid := m.contacts[m.cursor]
		msgs := m.conversations[jid]
		start := 0
		if len(msgs) > chatH { start = len(msgs) - chatH }
		chatContent = strings.Join(msgs[start:], "\n")
	} else { chatContent = "Waiting for messages..." }
	chatPane := styleChat.Width(m.width - 34).Height(chatH).BorderForeground(inactiveBorder).Render(chatContent)
	inputPane := styleInput.Width(m.width - 34).Render(m.input.View())
	rightSide := lipgloss.JoinVertical(lipgloss.Left, chatPane, inputPane)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightSide)
}

func printBanner() {
	cyan := "\033[36m"
	purple := "\033[35m"
	reset := "\033[0m"
	bold := "\033[1m"
	art := `
__        __  _           _                           
\ \      / / | |__   __ _| |_ ___  __ _ _ __  _ __    
 \ \ /\ / /| | '_ \ / _` + "`" + ` | __/ __|/ _` + "`" + ` | '_ \| '_ \   
  \ V  V / | | | | | (_| | |_\__ \ (_| | |_) | |_) |  
   \_/\_/  |_|_| |_|\__,_|\__|___/\__,_| .__/| .__/   
                                       |_|   |_|      `
	fmt.Println(cyan + bold + art + reset)
	fmt.Printf("%s   WHATSAPP CLI %s\n", purple, reset)
	fmt.Printf("   %sBy Parth Bhanti, to save your precious RAM%s\n\n", cyan, reset)
}

func main() {
	printBanner()
	dbLog := waLog.Stdout("Database", "ERROR", true)
	container, err := sqlstore.New(context.Background(), "sqlite", "file:whatsapp_store.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", dbLog)
	if err != nil { panic(err) }
	
	// Open a separate handle for our CLI messages to avoid type issues with whatsmeow's container
	db, err = sql.Open("sqlite", "file:whatsapp_store.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)")
	if err != nil { panic(err) }
	
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cli_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chat_jid TEXT,
		sender TEXT,
		content TEXT,
		is_from_me BOOLEAN,
		timestamp DATETIME
	)`)
	if err != nil { panic(err) }

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cli_settings (
		key TEXT PRIMARY KEY,
		value TEXT
	)`)
	if err != nil { panic(err) }

	historyEnabled := true
	var val string
	err = db.QueryRow("SELECT value FROM cli_settings WHERE key = 'history_enabled'").Scan(&val)
	if err == nil {
		historyEnabled = val == "true"
	} else {
		// Set default
		db.Exec("INSERT INTO cli_settings (key, value) VALUES ('history_enabled', 'true')")
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil { panic(err) }
	client := whatsmeow.NewClient(deviceStore, nil)
	p := tea.NewProgram(initialModel(client, historyEnabled), tea.WithAltScreen())

	client.AddEventHandler(func(evt interface{}) {
		switch v := evt.(type) {
		case *events.CallOffer:
			caller := v.CallCreator.User + "@s.whatsapp.net"
			content := "???? INCOMING CALL!"
			p.Send(incomingWAMsg{ChatJID: caller, Sender: "SYSTEM", Content: content, IsFromMe: false, Timestamp: v.Timestamp})
		case *events.Message:
			chatJID := v.Info.Chat
			senderJID := v.Info.Sender
			senderName := v.Info.PushName
			if senderName == "" {
				contact, _ := client.Store.Contacts.GetContact(context.Background(), senderJID)
				if contact.Found && contact.FullName != "" { senderName = contact.FullName } else { senderName = senderJID.User }
			}
			text := ""
			if v.Message.Conversation != nil { text = *v.Message.Conversation } else if v.Message.ExtendedTextMessage != nil { text = *v.Message.ExtendedTextMessage.Text } else {
				text = downloadMedia(client, v)
				if text == "" { text = "[Media/Other]" }
			}
			p.Send(incomingWAMsg{ChatJID: chatJID.String(), Sender: senderName, Content: text, IsFromMe: v.Info.IsFromMe, Timestamp: v.Info.Timestamp})
		}
	})

	if client.Store.ID == nil {
		qrChan, _ := client.GetQRChannel(context.Background())
		client.Connect()
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				fmt.Println("Scan QR code!")
				// FIX: Removed 'return' here. The loop will continue waiting for you to scan.
			} else {
				// Success!
				fmt.Println("Login successful! Launching...")
				break
			}
		}
	} else {
		client.Connect()
	}

	if _, err := p.Run(); err != nil { fmt.Printf("Error: %v", err) }
}
