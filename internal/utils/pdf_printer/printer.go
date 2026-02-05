package pdfprinter

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"path/filepath"
	"time"
	"os"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// Config Path (Sesuaikan dengan root project kamu)
const (
	TemplateFolder = "internal/utils/pdf_printer/templates"
	AssetsFolder   = "assets/images"
)

// GeneratePDF adalah fungsi utama yang dipanggil dari luar
func GeneratePDF(templateName string, data interface{}, outputPath string) error {
	// 1. PARSE HTML TEMPLATE
	tmplPath := filepath.Join(TemplateFolder, templateName)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("gagal parsing template %s: %v", tmplPath, err)
	}

	// 2. INJECT DATA KE HTML (Render to String)
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("gagal inject data ke template: %v", err)
	}
	htmlContent := body.String()

	// 3. SETUP CHROMEDP (HEADLESS BROWSER)
	// Opsi agar berjalan lancar di Docker/Server nantinya
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Buat Context Browser
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set Timeout (biar ga hang kalau error)
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 4. LAKUKAN PRINT TO PDF
	var pdfBuffer []byte
	
	err = chromedp.Run(ctx,
		// A. Load HTML String langsung ke Browser
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			// Set konten HTML ke frame utama
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),
		
		// B. Tunggu sebentar (optional, jaga-jaga render font/image selesai)
		chromedp.Sleep(500*time.Millisecond),

		// C. Print ke PDF
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPaperWidth(8.27).   
				WithPaperHeight(11.69). 
				WithScale(1).
				WithPrintBackground(true). 
				// MARGIN DI SET KECIL/NOL BIAR CSS YG ATUR
				WithMarginTop(0.4).     // Sekitar 1cm margin aman printer fisik
				WithMarginBottom(0.4).
				WithMarginLeft(0.4).
				WithMarginRight(0.4).
				Do(ctx)
			
			if err != nil {
				return err
			}
			pdfBuffer = buf
			return nil
		}),
	)

	if err != nil {
		return fmt.Errorf("gagal render PDF via Chrome: %v", err)
	}

	// 5. SIMPAN FILE (Sementara save ke disk, nanti bisa return []byte buat API)
	if err := os.WriteFile(outputPath, pdfBuffer, 0644); err != nil {
		return fmt.Errorf("gagal save file output: %v", err)
	}

	fmt.Printf("✅ PDF Berhasil dibuat: %s\n", outputPath)
	return nil
}