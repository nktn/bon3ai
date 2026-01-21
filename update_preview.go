package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/qeesung/image2ascii/convert"
)

// Preview mode operations

func (m *Model) openPreview() tea.Cmd {
	node := m.tree.GetNode(m.selected)
	if node == nil {
		return nil
	}

	if node.IsDir {
		m.message = "Cannot preview directory"
		return nil
	}

	// Reset preview state
	m.previewPath = node.Path
	m.previewScroll = 0
	m.previewIsImage = false
	m.imageWidth = 0
	m.imageHeight = 0
	m.imageFormat = ""
	m.imageSize = 0
	m.previewDiffLines = nil
	m.previewDiffMap = nil
	m.previewDiffIndex = -1

	// Check if image file
	if isImageFile(node.Path) {
		// Get image metadata
		imgWidth, imgHeight, imgFormat, imgSize, err := getImageInfo(node.Path)
		if err != nil {
			m.message = fmt.Sprintf("Error: %v", err)
			return nil
		}
		m.imageWidth = imgWidth
		m.imageHeight = imgHeight
		m.imageFormat = imgFormat
		m.imageSize = imgSize

		// In tmux, use tea.ExecProcess to run chafa directly (bypasses Bubble Tea rendering
		// which corrupts tmux passthrough escape sequences)
		if os.Getenv("TMUX") != "" {
			if _, err := exec.LookPath("chafa"); err == nil {
				m.execMode = true // Prevent View() from rendering during exec
				return m.execChafaPreview(node.Path)
			}
		}

		lines, err := m.loadImagePreview(node.Path)
		if err != nil {
			m.message = err.Error()
			return nil
		}
		m.previewContent = lines
		m.previewIsBinary = false
		// previewIsImage is set in loadImagePreview (true for chafa/Kitty, false for ASCII)
		m.inputMode = ModePreview
		// Clear screen for chafa/Kitty images to prevent background text showing through
		if m.previewIsImage {
			return tea.ClearScreen
		}
		return nil
	}

	file, err := os.Open(node.Path)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return nil
	}
	defer file.Close()

	limited := &io.LimitedReader{R: file, N: MaxPreviewBytes + 1}
	content, err := io.ReadAll(limited)
	if err != nil {
		m.message = fmt.Sprintf("Error: %v", err)
		return nil
	}

	truncated := len(content) > MaxPreviewBytes
	if truncated {
		content = content[:MaxPreviewBytes]
	}

	m.previewIsImage = false

	// Check if binary
	if isBinaryContent(content) {
		m.previewIsBinary = true
		m.previewContent = formatHexPreview(content, len(content))
	} else {
		m.previewIsBinary = false
		m.previewContent = strings.Split(string(content), "\n")
		// Load diff for text files
		m.loadFileDiff(node.Path)
	}

	if truncated {
		m.message = "Preview truncated (file > 512KB)"
	}

	m.inputMode = ModePreview
	return nil
}

// loadFileDiff loads VCS diff information for the file
func (m *Model) loadFileDiff(path string) {
	diffLines := m.vcsRepo.GetFileDiff(path)
	if len(diffLines) == 0 {
		return
	}

	// Clamp deletion markers that exceed content length (EOF deletions)
	contentLen := len(m.previewContent)
	if contentLen > 0 {
		for i := range diffLines {
			if diffLines[i].Line > contentLen {
				diffLines[i].Line = contentLen
			}
		}
	}

	m.previewDiffLines = diffLines
	m.previewDiffMap = make(map[int]DiffLine)
	for _, dl := range diffLines {
		m.previewDiffMap[dl.Line] = dl
	}
}

func (m *Model) closePreview() {
	m.inputMode = ModeNormal
	m.previewContent = nil
	m.previewPath = ""
	m.previewScroll = 0
	m.previewIsImage = false
	// Reset image metadata
	m.imageWidth = 0
	m.imageHeight = 0
	m.imageFormat = ""
	m.imageSize = 0
	// Reset diff state
	m.previewDiffLines = nil
	m.previewDiffMap = nil
	m.previewDiffIndex = -1
}

// clearKittyImages sends escape sequence to delete all Kitty graphics
func clearKittyImages() tea.Cmd {
	return tea.Printf("\x1b_Ga=d,d=A\x1b\\")
}

// Image file detection and metadata

func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	imageExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true,
		".gif": true, ".bmp": true, ".webp": true,
		".tiff": true, ".tif": true, ".ico": true,
	}
	return imageExts[ext]
}

// getImageInfo retrieves image metadata with graceful degradation.
// Returns dimensions, format, and file size. On decode failure, falls back to
// extension-based format detection. Only returns error if file doesn't exist.
func getImageInfo(path string) (width, height int, format string, size int64, err error) {
	// Get file size
	info, err := os.Stat(path)
	if err != nil {
		return 0, 0, "", 0, err
	}
	size = info.Size()

	// Get image dimensions and format
	file, err := os.Open(path)
	if err != nil {
		// Fallback: return format from extension if file can't be opened
		return 0, 0, getFormatFromExtension(path), size, nil
	}
	defer file.Close()

	config, formatName, err := image.DecodeConfig(file)
	if err != nil {
		// Fallback: return format from extension if image can't be decoded (e.g., ICO)
		return 0, 0, getFormatFromExtension(path), size, nil
	}

	return config.Width, config.Height, strings.ToUpper(formatName), size, nil
}

// getFormatFromExtension returns image format name based on file extension
func getFormatFromExtension(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	formats := map[string]string{
		".png":  "PNG",
		".jpg":  "JPEG",
		".jpeg": "JPEG",
		".gif":  "GIF",
		".bmp":  "BMP",
		".webp": "WEBP",
		".tiff": "TIFF",
		".tif":  "TIFF",
		".ico":  "ICO",
	}
	if f, ok := formats[ext]; ok {
		return f
	}
	return ""
}

// execChafaPreview runs chafa directly via tea.Exec for tmux compatibility
func (m *Model) execChafaPreview(path string) tea.Cmd {
	width := m.width - 2
	height := m.height - 4
	if width < 10 {
		width = 10
	}
	if height < 5 {
		height = 5
	}

	// Build info line (same format as view.go renderPreview)
	filename := filepath.Base(path)
	var info string
	if m.imageWidth > 0 && m.imageHeight > 0 {
		info = fmt.Sprintf(" %s (%dx%d %s, %s) ",
			filename,
			m.imageWidth, m.imageHeight,
			m.imageFormat,
			formatFileSize(m.imageSize))
	} else if m.imageFormat != "" {
		info = fmt.Sprintf(" %s (%s, %s) ",
			filename,
			m.imageFormat,
			formatFileSize(m.imageSize))
	} else {
		info = fmt.Sprintf(" %s ", filename)
	}

	// Use shell script with environment variables to prevent injection attacks.
	// Data is passed via env vars ($INFO, $IMAGE_PATH) instead of string interpolation.
	// Title uses cyan background (matching previewTitleStyle).
	// Use ANSI escape sequences instead of 'clear' command to reduce flicker:
	// \033[?25l = hide cursor, \033[H = home, \033[2J = clear screen
	script := fmt.Sprintf(`printf '\033[?25l\033[H\033[2J\033[46;30m%%s\033[0m\n\n' "$INFO"; chafa --format kitty --passthrough tmux --animate off --size %dx%d -- "$IMAGE_PATH"; printf '\n\033[90mPress any key to close...\033[0m'; read -n 1; printf '\033[?25h'`,
		width, height)

	cmd := exec.Command("bash", "-c", script)
	cmd.Env = append(os.Environ(), "INFO="+info, "IMAGE_PATH="+path)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return execDoneMsg{} // Signal exec completion to reset execMode
	})
}

// Image preview loading

func (m *Model) loadImagePreview(path string) ([]string, error) {
	// Calculate size based on terminal dimensions
	width := m.width - 2
	height := m.height - 4

	if width < 10 {
		width = 10
	}
	if height < 5 {
		height = 5
	}

	// Try chafa first (high quality with Kitty protocol)
	// Note: tmux uses execChafaPreview instead (View() corrupts tmux passthrough)
	if _, err := exec.LookPath("chafa"); err == nil {
		args := []string{
			"--format", "kitty",
			"--animate", "off",
			"--polite", "on",
			"--size", fmt.Sprintf("%dx%d", width, height),
		}
		args = append(args, "--", path) // Prevent option injection from filenames starting with -
		cmd := exec.Command("chafa", args...)
		output, err := cmd.Output()
		if err == nil {
			m.previewIsImage = true
			return strings.Split(string(output), "\n"), nil
		}
	}

	// Fallback to ASCII art using image2ascii
	return m.loadASCIIPreview(path, width, height)
}

func (m *Model) loadASCIIPreview(path string, width, height int) ([]string, error) {
	converter := convert.NewImageConverter()
	opts := convert.DefaultOptions
	opts.FixedWidth = width
	opts.FixedHeight = height
	opts.Colored = true

	result := converter.ImageFile2ASCIIString(path, &opts)
	if result == "" {
		return nil, fmt.Errorf("failed to convert image to ASCII")
	}

	m.previewIsImage = false // ASCII art doesn't need Kitty cleanup
	return strings.Split(result, "\n"), nil
}

// Binary content detection and hex preview

func isBinaryContent(content []byte) bool {
	// Check first 512 bytes for null bytes or high ratio of non-printable chars
	checkLen := len(content)
	if checkLen > 512 {
		checkLen = 512
	}

	// Empty content is not binary
	if checkLen == 0 {
		return false
	}

	nonPrintable := 0
	for i := 0; i < checkLen; i++ {
		b := content[i]
		if b == 0 {
			return true
		}
		if b < 32 && b != '\n' && b != '\r' && b != '\t' {
			nonPrintable++
		}
	}

	return float64(nonPrintable)/float64(checkLen) > 0.3
}

func formatHexPreview(content []byte, originalSize int) []string {
	var lines []string
	maxBytes := MaxHexPreviewBytes
	truncated := originalSize > maxBytes

	if len(content) > maxBytes {
		content = content[:maxBytes]
	}

	for i := 0; i < len(content); i += 16 {
		end := i + 16
		if end > len(content) {
			end = len(content)
		}
		chunk := content[i:end]

		// Hex part
		hexParts := make([]string, len(chunk))
		for j, b := range chunk {
			hexParts[j] = fmt.Sprintf("%02x", b)
		}
		hexStr := strings.Join(hexParts, " ")

		// Pad hex string
		for len(hexStr) < 47 {
			hexStr += " "
		}

		// ASCII part
		ascii := make([]byte, len(chunk))
		for j, b := range chunk {
			if b >= 32 && b < 127 {
				ascii[j] = b
			} else {
				ascii[j] = '.'
			}
		}

		lines = append(lines, fmt.Sprintf("%08x  %s  %s", i, hexStr, string(ascii)))
	}

	if truncated {
		lines = append(lines, "... (truncated)")
	}

	return lines
}
