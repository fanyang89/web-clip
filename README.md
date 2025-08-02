# Web Clip

A simple Go tool that takes screenshots of web pages from URLs or local HTML files using headless Chrome.

## Features

- Take screenshots from any URL
- Take screenshots from local HTML files
- Automatically detects input type (URL vs HTML file)
- Configurable DPI for high-quality screenshots
- Simple command-line interface

## Installation

```bash
git clone <repository-url>
cd web-clip
go build
```

Or run directly with:

```bash
go run web-clip.go <URL_OR_HTML_FILE> -output <OUTPUT_FILE>
```

## Usage

### Basic Usage

```bash
# Screenshot a URL
./web-clip https://example.com -output screenshot.png

# Screenshot a local HTML file
./web-clip ./local.html -output screenshot.png
```

### Advanced Usage

```bash
# Specify custom DPI (default: 200)
./web-clip https://example.com -output screenshot.png -dpi 300
```

## Command Line Options

```
Usage: web-clip <URL_OR_HTML_FILE> -output <OUTPUT_FILE>

Arguments:
  <URL_OR_HTML_FILE>    URL (http:// or https://) or path to local HTML file

Options:
  -output string        Output file path for the screenshot
  -dpi int              DPI for screenshot (default: 200)
  -help                 Show help message
```

## Input Detection

The tool automatically detects the input type:

- **URL**: Must start with `http://` or `https://`
- **HTML File**: Must be an existing file path

## Requirements

- Go 1.16 or later
- Chrome/Chromium browser (installed on system)

## Examples

```bash
# Screenshot a website
./web-clip https://github.com -output github.png

# Screenshot with high DPI
./web-clip https://example.com -output high-res.png -dpi 300

# Screenshot local HTML
./web-clip ./index.html -output local-screenshot.png
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Feel free to submit issues and pull requests.
