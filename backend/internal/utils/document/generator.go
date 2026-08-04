package document

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"

	"github.com/nguyenthenguyen/docx"
)

type DocumentRequest struct {
	Variables    map[string]interface{}
	TemplateName string
}

type ImageData struct {
	Data     string // base64-encoded image data
	MimeType string // e.g. "image/jpeg", "image/png"
}

type Generator struct {
	httpClient   *http.Client
	gotenbergURL string
	templateDir  string
}

func NewGenerator(gotenbergURL string, templateDir string) *Generator {
	return &Generator{
		gotenbergURL: gotenbergURL,
		templateDir:  templateDir,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

func (g *Generator) Generate(ctx context.Context, req DocumentRequest) ([]byte, error) {
	docxBytes, err := g.GenerateDocx(ctx, req)
	if err != nil {
		return nil, err
	}
	return g.convertToPDF(ctx, docxBytes)
}

func (g *Generator) GenerateDocx(ctx context.Context, req DocumentRequest) ([]byte, error) {
	templatePath := filepath.Join(g.templateDir, req.TemplateName)
	doc, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open template: %w", err)
	}
	defer doc.Close()

	d := doc.Editable()
	content := fixSplitRuns(d.GetContent())

	var keys []string
	for k := range req.Variables {
		keys = append(keys, k)
	}
	// Sort keys by length descending to prevent {{lampiran_1}} from partially replacing {{lampiran_10}}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	for _, k := range keys {
		v := req.Variables[k]
		val := html.EscapeString(fmt.Sprintf("%v", v))
		placeholder := fmt.Sprintf("{{%s}}", k)
		content = strings.ReplaceAll(content, placeholder, val)
		// Also handle angle brackets if any
		content = replaceAngleBrackets(content, k, val)
	}

	// Remove duplicate hardcoded 2026 after tanggal_no_surat if it exists
	if tgl, ok := req.Variables["tanggal_no_surat"].(string); ok && tgl != "" {
		yearStr := fmt.Sprintf(" %d", time.Now().Year())
		if strings.HasSuffix(tgl, yearStr) {
			content = regexp.MustCompile(`(Tanggal\s+`+regexp.QuoteMeta(tgl)+`(?:</w:t>.*?)?)(?:<w:t(?:.*?)>\s*`+regexp.QuoteMeta(fmt.Sprintf("%d", time.Now().Year()))+`\s*</w:t>)`).ReplaceAllString(content, "$1")
		}
	}

	// Handle 1-day travel format removing the dash
	if tglMulai, ok := req.Variables["tanggal_mulai"].(string); ok && tglMulai == "" {
		// Replace ` - {{tanggal_selesai}}` or just ` - ` that might be left in the template
		content = strings.ReplaceAll(content, "Tanggal  -", "Tanggal")
		content = strings.ReplaceAll(content, "Tanggal -", "Tanggal")
		content = regexp.MustCompile(`(Tanggal(?:</w:t>[^>]*>)*<w:t(?:[^>]*)>[\s]*)(?:\-[\s]*)(<)`).ReplaceAllString(content, "$1$2")
		content = regexp.MustCompile(`(Tanggal[\s]*)\-[\s]*`).ReplaceAllString(content, "$1")
		content = regexp.MustCompile(`(</w:t></w:r><w:r(?:[^>]*)><w:t(?:[^>]*)>[\s]*)\-([\s]*</w:t></w:r><w:r(?:[^>]*)><w:t(?:[^>]*)>(?:[0-9]{1,2}))`).ReplaceAllString(content, "$1$2")
		// Sometimes there is an en-dash
		content = regexp.MustCompile(`(Tanggal(?:</w:t>[^>]*>)*<w:t(?:[^>]*)>[\s]*)(?:\x{2013}[\s]*)(<)`).ReplaceAllString(content, "$1$2")
		content = regexp.MustCompile(`(Tanggal[\s]*)\x{2013}[\s]*`).ReplaceAllString(content, "$1")
	}

	// Remove duplicate year 2026 if it occurs explicitly right after a filled tanggal_no_surat
	// or we can just replace "Tanggal <tanggal_no_surat> 2026" with "Tanggal <tanggal_no_surat>"
	if tgl, ok := req.Variables["tanggal_no_surat"].(string); ok && tgl != "" {
		// Just remove " 2026 " or any year if it immediately follows the substituted text with some tags in between
		reRemoveHardcodedYear := regexp.MustCompile(regexp.QuoteMeta(tgl) + `(</w:t></w:r><w:r [^>]+><w:rPr><w:spacing [^>]+/></w:rPr><w:t xml:space="preserve">\s*</w:t></w:r><w:r [^>]+><w:t>)20\d\d(</w:t>)`)
		content = reRemoveHardcodedYear.ReplaceAllString(content, tgl+"$1$2")
	}

	content = removeEmptyTableRows(content)
	content = removeEmptyParagraphs(content)
	content = adjustPetugasTableWidths(content)
	d.SetContent(content)

	var docxBuf bytes.Buffer
	if err := d.Write(&docxBuf); err != nil {
		return nil, fmt.Errorf("failed to write docx: %w", err)
	}
	return docxBuf.Bytes(), nil
}

func (g *Generator) GenerateWithImages(ctx context.Context, req DocumentRequest, images map[string]ImageData) ([]byte, error) {
	docxBytes, err := g.GenerateDocxWithImages(ctx, req, images)
	if err != nil {
		return nil, err
	}
	return g.convertToPDF(ctx, docxBytes)
}

func (g *Generator) GenerateDocxWithImages(ctx context.Context, req DocumentRequest, images map[string]ImageData) ([]byte, error) {
	docxBytes, err := g.GenerateDocx(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(images) > 0 {
		docxBytes, err = injectImages(docxBytes, images)
		if err != nil {
			return nil, fmt.Errorf("failed to inject images: %w", err)
		}
	}
	return docxBytes, nil
}

func injectImages(docxBytes []byte, images map[string]ImageData) ([]byte, error) {
	reader, err := zip.NewReader(bytes.NewReader(docxBytes), int64(len(docxBytes)))
	if err != nil {
		return nil, err
	}

	var outBuf bytes.Buffer
	writer := zip.NewWriter(&outBuf)

	imageInfoMap := make(map[string]imageInfo)
	imageFiles := make(map[string][]byte)
	rIdCounter := 50000

	for key, img := range images {
		imgBytes, err := decodeBase64Image(img.Data)
		if err != nil {
			continue
		}

		ext := getImageExtension(img.MimeType)
		rId := fmt.Sprintf("rIdImg%d", rIdCounter)
		filename := fmt.Sprintf("media/img_%d%s", rIdCounter, ext)

		// Aspect ratio
		cx, cy := calculateDimensions(imgBytes, key)

		imageInfoMap[key] = imageInfo{rId: rId, filename: filename, cx: cx, cy: cy}
		imageFiles[filename] = imgBytes
		rIdCounter++
	}

	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		if err != nil {
			rc.Close()
			return nil, err
		}
		rc.Close()

		name := file.Name
		switch name {
		case "word/document.xml":
			contentStr := string(content)

			var keys []string
			for k := range imageInfoMap {
				keys = append(keys, k)
			}
			sort.Slice(keys, func(i, j int) bool {
				return len(keys[i]) > len(keys[j])
			})

			for _, key := range keys {
				info := imageInfoMap[key]
				contentStr = replaceImagePlaceholder(contentStr, key, info.rId, info.cx, info.cy)
			}
			content = []byte(contentStr)
		case "word/_rels/document.xml.rels":
			contentStr := string(content)
			var rels []string
			for _, info := range imageInfoMap {
				rels = append(rels, fmt.Sprintf(`<Relationship Id="%s" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="%s"/>`, info.rId, info.filename))
			}
			contentStr = strings.Replace(contentStr, "</Relationships>", strings.Join(rels, "")+"</Relationships>", 1)
			content = []byte(contentStr)
		case "[Content_Types].xml":
			contentStr := string(content)
			if !strings.Contains(contentStr, `Extension="jpeg"`) {
				contentStr = strings.Replace(contentStr, "</Types>", `<Default Extension="jpeg" ContentType="image/jpeg"/>`+"</Types>", 1)
			}
			if !strings.Contains(contentStr, `Extension="png"`) {
				contentStr = strings.Replace(contentStr, "</Types>", `<Default Extension="png" ContentType="image/png"/>`+"</Types>", 1)
			}
			content = []byte(contentStr)
		}

		w, err := writer.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(content); err != nil {
			return nil, err
		}
	}

	for name, data := range imageFiles {
		w, err := writer.Create("word/" + name)
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
	}

	writer.Close()
	return outBuf.Bytes(), nil
}

type imageInfo struct {
	rId      string
	filename string
	cx, cy   int64
}

func calculateDimensions(imgBytes []byte, key string) (int64, int64) {
	// Default 15cm x 10cm in EMUs
	cx, cy := int64(5400000), int64(3600000)
	if cfg, _, err := image.DecodeConfig(bytes.NewReader(imgBytes)); err == nil {
		w, h := float64(cfg.Width), float64(cfg.Height)

		// Handle EXIF orientation
		if x, err := exif.Decode(bytes.NewReader(imgBytes)); err == nil {
			if tag, err := x.Get(exif.Orientation); err == nil {
				if orientation, err := tag.Int(0); err == nil && orientation >= 5 && orientation <= 8 {
					w, h = h, w
				}
			}
		}

		maxW, maxH := 5400000.0, 8000000.0 // Default ~15cm x 22cm
		if strings.Contains(key, "foto_dokumentasi") {
			maxW, maxH = 4050000.0, 6000000.0 // 75% size: ~11cm x 16.5cm
		}

		ratio := maxW / w
		if h*ratio > maxH {
			ratio = maxH / h
		}
		cx, cy = int64(w*ratio), int64(h*ratio)
	}
	return cx, cy
}

func replaceImagePlaceholder(xml, key, rId string, cx, cy int64) string {
	docPrId := hash(key)
	drawing := fmt.Sprintf(`</w:t></w:r><w:r><w:drawing><wp:inline distT="0" distB="0" distL="0" distR="0"><wp:extent cx="%d" cy="%d"/><wp:effectExtent l="0" t="0" r="0" b="0"/><wp:docPr id="%d" name="img_%s"/><wp:cNvGraphicFramePr><a:graphicFrameLocks xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:nvPicPr><pic:cNvPr id="%d" name="img_%s"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="%s" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"/><a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr bwMode="auto"><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom><a:noFill/></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r><w:r><w:t>`, cx, cy, docPrId, key, docPrId, key, rId, cx, cy)

	if key == "lampiran_1" {
		drawing = `</w:t></w:r><w:r><w:br w:type="page"/></w:r><w:r><w:t>` + drawing
	}

	placeholder1 := "{{" + key + "}}"
	placeholder2 := "&lt;&lt;" + key + "&gt;&gt;"
	placeholder3 := "<<" + key + ">>"

	xml = strings.ReplaceAll(xml, placeholder1, drawing)
	xml = strings.ReplaceAll(xml, placeholder2, drawing)
	xml = strings.ReplaceAll(xml, placeholder3, drawing)

	return xml
}

func (g *Generator) convertToPDF(ctx context.Context, docxBytes []byte) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("files", "document.docx")
	if err != nil {
		return nil, err
	}
	if _, errCopy := io.Copy(part, bytes.NewReader(docxBytes)); errCopy != nil {
		return nil, errCopy
	}
	if errClose := writer.Close(); errClose != nil {
		return nil, errClose
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.gotenbergURL+"/forms/libreoffice/convert", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return nil, fmt.Errorf("gotenberg error (%d): failed to read body", resp.StatusCode)
		}
		return nil, fmt.Errorf("gotenberg error (%d): %s", resp.StatusCode, string(b))
	}
	
	b, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}
	return b, nil
}

func (g *Generator) MergePDFs(ctx context.Context, pdfs [][]byte) ([]byte, error) {
	if len(pdfs) < 2 {
		if len(pdfs) == 1 {
			return pdfs[0], nil
		}
		return nil, fmt.Errorf("no pdfs")
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for i, p := range pdfs {
		part, errPart := writer.CreateFormFile("files", fmt.Sprintf("%d.pdf", i))
		if errPart != nil {
			return nil, errPart
		}
		if _, errCopy := io.Copy(part, bytes.NewReader(p)); errCopy != nil {
			return nil, errCopy
		}
	}
	if errClose := writer.Close(); errClose != nil {
		return nil, errClose
	}

	req, err := http.NewRequestWithContext(ctx, "POST", g.gotenbergURL+"/forms/pdfengines/merge", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to merge pdfs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			return nil, fmt.Errorf("gotenberg merge error (%d): failed to read body", resp.StatusCode)
		}
		return nil, fmt.Errorf("gotenberg merge error (%d): %s", resp.StatusCode, string(b))
	}

	b, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		return nil, errRead
	}
	return b, nil
}

func fixSplitRuns(xml string) string {
	re := regexp.MustCompile(`(\{)(<[^>]+>)+(\{)`)
	xml = re.ReplaceAllString(xml, "$1$3")
	re = regexp.MustCompile(`(\})(<[^>]+>)+(\})`)
	xml = re.ReplaceAllString(xml, "$1$3")

	reVar := regexp.MustCompile(`\{\{([^{}]+)\}\}`)
	reTag := regexp.MustCompile(`<[^>]+>`)
	xml = reVar.ReplaceAllStringFunc(xml, func(match string) string {
		return reTag.ReplaceAllString(match, "")
	})

	return xml
}

func replaceAngleBrackets(content, key, val string) string {
	content = strings.ReplaceAll(content, "&lt;&lt;"+key+"&gt;&gt;", val)
	content = strings.ReplaceAll(content, "<<"+key+">>", val)
	return content
}

func removeEmptyTableRows(xml string) string {
	marker := "__REMOVE_ROW__"
	for strings.Contains(xml, marker) {
		idx := strings.Index(xml, marker)

		trStart1 := strings.LastIndex(xml[:idx], "<w:tr>")
		trStart2 := strings.LastIndex(xml[:idx], "<w:tr ")

		trStart := trStart1
		if trStart2 > trStart {
			trStart = trStart2
		}

		trEnd := strings.Index(xml[idx:], "</w:tr>")
		if trStart != -1 && trEnd != -1 {
			xml = xml[:trStart] + xml[idx+trEnd+7:]
		} else {
			xml = strings.Replace(xml, marker, "", 1)
		}
	}
	return xml
}

func removeEmptyParagraphs(xml string) string {
	marker := "__REMOVE_P__"
	for strings.Contains(xml, marker) {
		idx := strings.Index(xml, marker)

		pStart1 := strings.LastIndex(xml[:idx], "<w:p>")
		pStart2 := strings.LastIndex(xml[:idx], "<w:p ")

		pStart := pStart1
		if pStart2 > pStart {
			pStart = pStart2
		}

		pEnd := strings.Index(xml[idx:], "</w:p>")
		if pStart != -1 && pEnd != -1 {
			endIdx := idx + pEnd + 6
			foundPageBreak := false

			// Check forward
			nextPStart1 := strings.Index(xml[endIdx:], "<w:p>")
			nextPStart2 := strings.Index(xml[endIdx:], "<w:p ")
			nextPStart := nextPStart1
			if nextPStart2 != -1 && (nextPStart == -1 || nextPStart2 < nextPStart) {
				nextPStart = nextPStart2
			}
			if nextPStart != -1 && nextPStart < 20 {
				nextPEnd := strings.Index(xml[endIdx+nextPStart:], "</w:p>")
				if nextPEnd != -1 {
					nextPContent := xml[endIdx+nextPStart : endIdx+nextPStart+nextPEnd+6]
					if strings.Contains(nextPContent, `<w:br w:type="page"/>`) {
						endIdx = endIdx + nextPStart + nextPEnd + 6
						foundPageBreak = true
					}
				}
			}

			// If not found forward, check backward
			if !foundPageBreak {
				prevPEnd1 := strings.LastIndex(xml[:pStart], "</w:p>")
				if prevPEnd1 != -1 && (pStart-prevPEnd1) < 20 {
					prevPStart1 := strings.LastIndex(xml[:prevPEnd1], "<w:p>")
					prevPStart2 := strings.LastIndex(xml[:prevPEnd1], "<w:p ")
					prevPStart := prevPStart1
					if prevPStart2 > prevPStart {
						prevPStart = prevPStart2
					}
					if prevPStart != -1 {
						prevPContent := xml[prevPStart : prevPEnd1+6]
						if strings.Contains(prevPContent, `<w:br w:type="page"/>`) {
							pStart = prevPStart
						}
					}
				}
			}

			xml = xml[:pStart] + xml[endIdx:]
		} else {
			xml = strings.Replace(xml, marker, "", 1)
		}
	}
	return xml
}

func adjustPetugasTableWidths(xml string) string {
	xml = strings.ReplaceAll(xml, `<w:tcW w:w="1906" w:type="dxa"/>`, `<w:tcW w:w="700" w:type="dxa"/>`)
	xml = strings.ReplaceAll(xml, `<w:tcW w:w="5391" w:type="dxa"/>`, `<w:tcW w:w="6597" w:type="dxa"/>`)
	return xml
}

func hash(s string) int {
	h := 0
	for _, c := range s {
		h = 31*h + int(c)
	}
	if h < 0 {
		h = -h
	}
	return h % 100000
}

func decodeBase64Image(data string) ([]byte, error) {
	if i := strings.Index(data, ","); i != -1 {
		data = data[i+1:]
	}
	return base64.StdEncoding.DecodeString(data)
}

func getImageExtension(mime string) string {
	if strings.Contains(mime, "png") {
		return ".png"
	}
	return ".jpeg"
}
