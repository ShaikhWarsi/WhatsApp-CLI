# WhatsApp CLI

> "By Parth Bhanti, to save your precious RAM"

A lightweight, interactive Command Line Interface (CLI) for WhatsApp, written in Go. It uses the [whatsmeow](https://github.com/tulir/whatsmeow) library to communicate with WhatsApp servers and [Bubble Tea](https://github.com/charmbracelet/bubbletea) for the Text User Interface (TUI).

## Screenshots

![WhatsApp CLI TUI View 1](ss1.png)
*Interface*

![WhatsApp CLI TUI View 2](ss2.png)
*CLI View*

## Features

- **Terminal-based UI**: Detailed and interactive TUI navigating chats and messages.
- **Low Resource Usage**: designed as a lighter alternative to the Electron-based WhatsApp Desktop app.
- **Media Support**: Automatically downloads Images, Videos, and Documents to a local `downloads` folder.
- **Real-time**: Receives messages and updates instantly.
- **Cross-Platform**: Binaries available for Linux, macOS, and Windows.

## Installation

### Automatic Install (Linux/macOS)

You can install the latest version using the provided script:

```bash
chmod +x install.sh
./install.sh
```

Or download and run directly (assuming the repo is public):

```bash
curl -sL https://raw.githubusercontent.com/parthbhanti22/WhatsApp-CLI/main/install.sh | bash
```

### Manual Install

1.  Download the binary for your OS from the [Releases](https://github.com/parthbhanti22/WhatsApp-CLI/releases) page.
2.  Make it executable (Linux/macOS):
    ```bash
    chmod +x whatsapp-cli
    ```
3.  Move it to your path or run it directly.

### Windows

Download `whatsapp-win.exe` and run it from your command prompt or PowerShell.

## Building from Source

Requirements: Go 1.24+

```bash
git clone https://github.com/parthbhanti22/WhatsApp-CLI.git
cd WhatsApp-CLI
go mod tidy
go build -o whatsapp-cli main.go
```

## Usage

1.  Run the application:
    ```bash
    ./whatsapp-cli
    ```
2.  **First Run**: You will see a QR code in your terminal.
3.  Open WhatsApp on your phone, go to **Linked Devices** -> **Link a Device**, and scan the QR code.
4.  Once connected, your chat list will appear.

### keybindings

-   **Up / Down Arrow**: Navigate through the contact list.
-   **Enter**: Select a contact to chat (focus input).
-   **Type Message + Enter**: Send a message to the selected contact.
-   **Ctrl+C** or **Esc**: Quit the application.

## Project Structure

-   `main.go`: Application entry point and TUI logic.
-   `go.mod` / `go.sum`: Go dependencies.
-   `install.sh`: Installation script for Unix-like systems.

## Credits

-   **Author**: [Parth Bhanti](https://github.com/parthbhanti22)
-   **Libraries**:
    -   [whatsmeow](https://go.mau.fi/whatsmeow) (WhatsApp Protocol)
    -   [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI Framework)
    -   [Lip Gloss](https://github.com/charmbracelet/lipgloss) (Styling)


