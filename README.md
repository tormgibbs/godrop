# godrop

**godrop** is a command-line utility to securely share a file or directory over a temporary [Cloudflare Tunnel](https://www.cloudflare.com/products/tunnel/). It wraps the shared file in an easily accessible, temporary link.

## Features

  * **Secure & Temporary Sharing**: Uses a temporary Cloudflare Tunnel to expose the file, and the link expires when `godrop` is stopped.
  * **File and Directory Support**: Can share a single file, a whole directory (automatically zipped), or multiple files (also automatically zipped).
  * **Download Limits**: Allows you to set a maximum number of downloads before the server automatically shuts down.
  * **Download Logging**: Prints access logs, including the downloader's IP and User-Agent, to the console.

## Installation

`godrop` is a Go program and requires Go **1.25.1** or later.

```bash
go install github.com/tormgibbs/godrop@v1.0.0
```

*Note: You must also have the `cloudflared` executable installed and accessible in your system's PATH for the tunneling functionality to work.*

## Usage

The basic usage is to pass the path to the file or directory you want to share.

```bash
godrop [file|directory]
```

### Examples

**1. Share a single file:**

```bash
godrop ./my-document.pdf
```

**2. Share a directory (it will be zipped automatically):**

```bash
godrop ~/MyProjectFolder
```

**3. Share multiple files (they will be zipped together):**

```bash
godrop fileA.txt fileB.jpg
```

### Options

| Flag | Shorthand | Default | Description |
| :--- | :--- | :--- | :--- |
| `--once` | `-o` | `false` | Serve once and exit. This is equivalent to setting `--limit 1`. |
| `--port` | `-p` | `8080` | Port to listen on for the local HTTP server. |
| `--limit` | `-l` | `0` | Maximum number of downloads before the server shuts down (0 means no limit). |
| `--zip-name` | `-z` | `""` | Specify the name for the temporary zip file when sharing a directory or multiple items. |

### Example with Options

Serve a file once and then shut down:

```bash
godrop -o ./important-file.tar.gz
# or
godrop -l 1 ./important-file.tar.gz
```

Share a directory with a custom zip name and allow up to 5 downloads:

```bash
godrop -l 5 -z "archive-name.zip" ~/secret-folder
```

-----

*Press `Ctrl+C` to stop sharing at any time.*