# PassPort

![logo](img/logo_bigger.png)

A secure, cross-platform password manager built with Go and Fyne. PassPort provides both a graphical user interface (GUI) and command-line interface (CLI) for managing encrypted passwords and sensitive information.

## Features

- **Secure Encryption**: AES encryption with PBKDF2 key derivation for maximum security
- **Cross-Platform**: Works on Windows and macOS
- **Dual Interface**: 
  - Graphical User Interface (GUI) built with Fyne
  - Command-Line Interface (CLI) for terminal users
- **Master Password**: Single master password protects all stored credentials
- **Easy Installation**: Includes installers for Windows
- **Customizable Theme**: Configurable application theme

## System Requirements

- Go 1.26.0 or later (for building from source)
- Windows 10+ or macOS 10.13+
- OpenGL support for GUI mode

## Installation

### Windows

#### Using winget

The easiest way to install PassPort on Windows is using the Windows Package Manager:

```cmd
winget install PassPort
```

#### Manual Installation

1. Download the latest release from the GitHub repository
2. Extract the executable to a folder of your choice
3. (Optional) Add the folder to your PATH environment variable for easy command-line access

#### Building from Source

1. Ensure Go 1.26.0+ is installed
2. Clone the repository:
   ```cmd
   git clone https://github.com/Buct0r/PassPort.git
   cd PassPort
   ```

3. Build the GUI version:
   ```cmd
   go build -o PassPort.exe ./src
   ```

4. Or build the CLI version:
   ```cmd
   go build -o PassPort-cli.exe ./cli
   ```

### Linux

#### Building from Source

1. Ensure Go 1.26.0+ is installed
2. Clone the repository:
   ```bash
   git clone https://github.com/Buct0r/PassPort.git
   cd PassPort
   ```

3. Install dependencies (for GUI):
   ```bash
   sudo apt-get install libgl1-mesa-dev xorg-dev
   ```

4. Build the GUI version:
   ```bash
   go build -o PassPort ./src
   ```

5. Or build the CLI version:
   ```bash
   go build -o PassPort-cli ./cli
   ```

### macOS

#### Using Homebrew

The easiest way to install PassPort on macOS is using Homebrew:

```bash
brew tap Buct0r/PassPort
brew install PassPort
```

To install the CLI version only:

```bash
brew install PassPort --without-gui
```

#### Building from Source

1. Ensure Go 1.26.0+ is installed and Xcode Command Line Tools:
   ```bash
   xcode-select --install
   ```

2. Clone the repository:
   ```bash
   git clone https://github.com/Buct0r/PassPort.git
   cd PassPort
   ```

3. Build the GUI version:
   ```bash
   go build -o PassPort ./src
   ```

4. Or build the CLI version:
   ```bash
   go build -o PassPort-cli ./cli
   ```

5. (Optional) Move the binary to a location in your PATH:
   ```bash
   sudo mv PassPort /usr/local/bin/
   ```

## Usage

### GUI Mode

1. Run `PassPort.exe`
2. On first launch, set up your master password
3. Authenticate with your master password
4. Use the graphical interface to:
   - Add new passwords/secrets
   - View stored credentials
   - Search for passwords
   - Manage your password vault

### CLI Mode

1. Run `PassPort.exe --cli` or `PassPort-cli.exe`
2. Authenticate with your master password
3. Navigate the menu to:
   - Add new passwords
   - Check existing passwords
   - Search for specific credentials
   - Manage your vault

### Command-Line Options

```
PassPort [FLAGS]

FLAGS:
  --cli, -c          Run in command-line interface mode
  --version, -v      Show version information
  --help, -h         Show help message
```

## Security

PassPort takes security seriously. The application features:

- **PBKDF2 Key Derivation**: Strengthens master password against brute-force attacks
- **AES Encryption**: Industry-standard encryption for stored data
- **No Cloud Storage**: All data remains local on your machine
- **Master Password Required**: All access requires authentication



## File Structure

```
PassPort/
├── src/                 # GUI application source code
│   ├── main.go
│   ├── gui.go
│   ├── encrypt.go
│   ├── functions.go
│   └── ...
├── cli/                 # CLI application source code
│   ├── cli.go
│   ├── encrypt.go
│   ├── functions.go
│   └── ...
├── go.mod              # Go module definition
└── config.json         # Configuration file
```

## Configuration

### Theme

Customize the application theme by editing `config.json`:

```json
{
  "theme": "CustomTheme"
}
```

### Data Storage

- Master password: `master.key` (stored in user's home directory)
- Encrypted vault: `SECRET` file (stored in user's home directory)

## Development

### Prerequisites

- Go 1.26.0+
- For GUI: Fyne framework (included in dependencies)
- OpenGL development libraries

### Dependencies

Key dependencies:
- `fyne.io/fyne/v2` - GUI framework
- `golang.org/x/crypto` - Cryptographic functions
- `golang.org/x/term` - Terminal control

### Building for Different Platforms

GUI application:
```bash
go build -o PassPort ./src
```

CLI application:
```bash
go build -o PassPort-cli ./cli
```

## Contributing

Contributions are welcome! Please ensure:
- Code passes security review
- Documentation is updated
- Cross-platform compatibility is maintained

## License

[Add your license information here]

## Support

For issues, questions, or suggestions, please open a Github issue.

## Disclaimer

PassPort is provided as-is. While security has been prioritized, users should regularly review security documentation and updates. Keep your master password safe and unique.

---

**Version**: 0.1  
**Last Updated**: April 2026  
**Maintainer**: Buct0r
