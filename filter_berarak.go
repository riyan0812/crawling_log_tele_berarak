package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var discordWebhookURL = "x" // Ganti dengan webhook kamu

type rgOutput struct {
	Type string `json:"type"`
	Data struct {
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
	} `json:"data"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./keywordFinder <folder_path> <keyword>")
		return
	}
	folderPath := os.Args[1]
	keyword := os.Args[2]

	files, err := filepath.Glob(filepath.Join(folderPath, "*.txt"))
	if err != nil {
		fmt.Println("❌ Error mencari file:", err)
		return
	}
	totalFiles := len(files)
	if totalFiles == 0 {
		fmt.Println("⚠️  Tidak ada file .txt di folder:", folderPath)
		return
	}

	fmt.Printf("📂 Ditemukan %d file .txt di folder %s\n", totalFiles, folderPath)
	fmt.Println("🔍 Mulai proses pencarian keyword:", keyword)

	var matchedLines []string
	processed := 0

	for _, file := range files {
		lines, err := searchKeywordInFileWithRG(file, keyword)
		if err == nil && len(lines) > 0 {
			matchedLines = append(matchedLines, lines...)
		}
		processed++
		animateSnailProgress(processed, totalFiles)
	}

	fmt.Println() // newline setelah animasi progress

	if len(matchedLines) == 0 {
		fmt.Println("⚠️  Tidak ditemukan baris dengan keyword:", keyword)
		return
	}

	outputFile := "hasil_filter.txt"
	err = ioutil.WriteFile(outputFile, []byte(strings.Join(matchedLines, "\n")), 0644)
	if err != nil {
		fmt.Println("❌ Gagal menyimpan hasil:", err)
		return
	}
	fmt.Println("✅ Hasil disimpan di", outputFile)

	err = sendFileToDiscord(outputFile)
	if err != nil {
		fmt.Println("❌ Gagal mengirim ke Discord:", err)
		return
	}
	fmt.Println("✅ Berhasil mengirim hasil ke Discord")
	fmt.Println("🎉 FINISH ✅")
}

func searchKeywordInFileWithRG(filePath, keyword string) ([]string, error) {
	// panggil perintah: rg -F --json -w --json-seq-start <keyword> <filePath>
	cmd := exec.Command("rg", "-F", keyword, "--json", filePath)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(stdout)
	var matches []string

	for scanner.Scan() {
		line := scanner.Text()

		var rgLine rgOutput
		if err := json.Unmarshal([]byte(line), &rgLine); err == nil {
			if rgLine.Type == "match" {
				// Format output: <file_basename>: <matched_line_text>
				matches = append(matches, fmt.Sprintf("%s: %s", filepath.Base(filePath), rgLine.Data.Lines.Text))
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	err = cmd.Wait()
	if err != nil {
		// rg mengembalikan kode != 0 jika tidak ada hasil, kita ignore error ini
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				// no matches found, bukan error
				return matches, nil
			}
		}
		return nil, err
	}
	return matches, nil
}

func animateSnailProgress(current, total int) {
	width := 40
	percent := float64(current) / float64(total)
	pos := int(percent * float64(width))
	if pos > width {
		pos = width
	}

	bar := strings.Repeat("=", pos) + "🐌" + strings.Repeat(" ", width-pos)
	fmt.Printf("\r[%s] %3.0f%%", bar, percent*100)
}

func sendFileToDiscord(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	_ = writer.WriteField("payload_json", `{"content": "Berikut hasil pencarian keyword"}`)

	err = writer.Close()
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", discordWebhookURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Discord API error: %s", string(respBody))
	}
	return nil
}
