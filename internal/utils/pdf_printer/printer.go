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

const (
	TemplateFolder = "internal/utils/pdf_printer/templates"
	AssetsFolder   = "assets/images"
)

func GeneratePDF(templateName string, data interface{}, outputPath string) error {
	tmplPath := filepath.Join(TemplateFolder, templateName)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("gagal parsing template %s: %v", tmplPath, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("gagal inject data ke template: %v", err)
	}
	htmlContent := body.String()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfBuffer []byte
	
	err = chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			frameTree, err := page.GetFrameTree().Do(ctx)
			if err != nil {
				return err
			}
			return page.SetDocumentContent(frameTree.Frame.ID, htmlContent).Do(ctx)
		}),

		chromedp.Sleep(500*time.Millisecond),

		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().
				WithPaperWidth(8.27).   
				WithPaperHeight(11.69). 
				WithScale(1).
				WithPrintBackground(true).
				WithMarginTop(0.4).
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

	if err := os.WriteFile(outputPath, pdfBuffer, 0644); err != nil {
		return fmt.Errorf("gagal save file output: %v", err)
	}

	fmt.Printf("✅ PDF Berhasil dibuat: %s\n", outputPath)
	return nil
}