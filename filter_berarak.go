package main

import (
  "bufio"
  "bytes"
  "fmt"
  "io"
  "mime/multipart"
  "net/http"
  "os"
  "os/exec"
  "path/filepath"
  "strings"
)

const discordWebhookURL = "https://discord.com/api/webhooks/1378856121522982924/IE0hsInpOCTmlQcrxfcxENruI2W2W1HqXTjpzBlbz78b5Hv7T4L9VC3zVwTGwYfSs05d"

// Fungsi memanggil ripgrep dan mengembalikan hasil pencarian sebagai slice string
func searchWithRg(folder string, keyword string) ([]string, error) {
  cmd := exec.Command("rg", "-F", keyword, folder)
  var out bytes.Buffer
  cmd.Stdout = &out
  err := cmd.Run()
  if err != nil {
    // Jika output kosong atau error, bisa return hasil kosong
    if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
      // Exit code 1 artinya tidak ada hasil pencarian, ini bukan error fatal
      return []string{}, nil
    }
    return nil, err
  }

  scanner := bufio.NewScanner(&out)
  var results []string
  for scanner.Scan() {
    results = append(results, scanner.Text())
  }
  return results, nil
}

// Fungsi menyimpan hasil ke file seperti sebelumnya
func saveToFile(lines []string, outputPath string) error {
  f, err := os.Create(outputPath)
  if err != nil {
    return err
  }
  defer f.Close()
  for _, line := range lines {
    _, err := f.WriteString(line + "\n")
    if err != nil {
      return err
    }
  }
  return nil
}

// Fungsi mengirim file ke Discord seperti sebelumnya
func sendToDiscord(filename string) error {
  file, err := os.Open(filename)
  if err != nil {
    return err
  }
  defer file.Close()

  body := &bytes.Buffer{}
  writer := multipart.NewWriter(body)
  part, err := writer.CreateFormFile("file", filepath.Base(filename))
  if err != nil {
    return err
  }
  _, err = io.Copy(part, file)
  if err != nil {
    return err
  }
  writer.Close()

  req, err := http.NewRequest("POST", discordWebhookURL, body)
  if err != nil {
    return err
  }
  req.Header.Set("Content-Type", writer.FormDataContentType())

  client := &http.Client{}
  resp, err := client.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode >= 200 && resp.StatusCode < 300 {
    fmt.Println("✅ Berhasil mengirim ke Discord!")
  } else {
    fmt.Println("❌ Gagal mengirim ke Discord:", resp.Status)
  }
  return nil
}

func main() {
  var folderPath, keyword string
  fmt.Print("📂 Masukkan path folder yang berisi file .txt: ")
  fmt.Scanln(&folderPath)

  fmt.Print("🔍 Masukkan kata kunci yang ingin dicari: ")
  fmt.Scanln(&keyword)

  fmt.Println("⏳ Memproses...")

  resultLines, err := searchWithRg(folderPath, keyword)
  if err != nil {
    fmt.Println("❌ Error saat menjalankan ripgrep:", err)
    return
  }

  if len(resultLines) == 0 {
    fmt.Println("⚠️  Tidak ditemukan baris yang mengandung kata kunci.")
    return
  }

  outputFile := "hasil_filter.txt"
  err = saveToFile(resultLines, outputFile)
  if err != nil {
    fmt.Println("❌ Error saat menyimpan hasil:", err)
    return
  }

  err = sendToDiscord(outputFile)
  if err != nil {
    fmt.Println("❌ Error saat mengirim ke Discord:", err)
  }
}
