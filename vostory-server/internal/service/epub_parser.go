package service

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html"
)

// EpubChapter 从 epub 中解析出的章节
type EpubChapter struct {
	Title     string
	Content   string
	WordCount int
}

// ParseEpubFile 解析 epub 文件，返回按阅读顺序排列的章节列表
func ParseEpubFile(filePath string) ([]EpubChapter, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开 epub 文件失败: %w", err)
	}
	defer r.Close()

	files := make(map[string]*zip.File)
	for _, f := range r.File {
		files[f.Name] = f
	}

	opfPath, err := findOPFPath(files)
	if err != nil {
		return nil, err
	}
	opfDir := path.Dir(opfPath)

	pkg, err := parseOPF(files[opfPath])
	if err != nil {
		return nil, err
	}

	manifest := make(map[string]opfManifestItem)
	for _, item := range pkg.Manifest.Items {
		manifest[item.ID] = item
	}

	tocTitles := make(map[string]string)
	if tocItem, ok := manifest["ncx"]; ok {
		tocPath := resolvePath(opfDir, tocItem.Href)
		if f, exists := files[tocPath]; exists {
			tocTitles, _ = parseNCXTitles(f, opfDir)
		}
	}
	if len(tocTitles) == 0 {
		for _, item := range pkg.Manifest.Items {
			if item.Properties == "nav" || strings.HasSuffix(item.Href, "nav.xhtml") {
				navPath := resolvePath(opfDir, item.Href)
				if f, exists := files[navPath]; exists {
					tocTitles, _ = parseNavTitles(f, opfDir)
				}
				break
			}
		}
	}

	var chapters []EpubChapter
	chapterNum := 0

	for _, ref := range pkg.Spine.ItemRefs {
		item, ok := manifest[ref.IDRef]
		if !ok {
			continue
		}

		href := resolvePath(opfDir, item.Href)
		f, exists := files[href]
		if !exists {
			continue
		}

		text, err := extractTextFromXHTML(f)
		if err != nil || strings.TrimSpace(text) == "" {
			continue
		}

		wordCount := utf8.RuneCountInString(strings.TrimSpace(text))
		if wordCount < 50 {
			continue
		}

		chapterNum++
		title := tocTitles[item.Href]
		if title == "" {
			title = tocTitles[href]
		}
		if title == "" {
			title = fmt.Sprintf("第%d章", chapterNum)
		}

		chapters = append(chapters, EpubChapter{
			Title:     title,
			Content:   strings.TrimSpace(text),
			WordCount: wordCount,
		})
	}

	if len(chapters) == 0 {
		return nil, fmt.Errorf("epub 中未解析到有效章节")
	}

	return chapters, nil
}

// --- container.xml ---

type epubContainer struct {
	XMLName   xml.Name         `xml:"container"`
	RootFiles []epubRootFile   `xml:"rootfiles>rootfile"`
}

type epubRootFile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

func findOPFPath(files map[string]*zip.File) (string, error) {
	f, ok := files["META-INF/container.xml"]
	if !ok {
		return "", fmt.Errorf("epub 缺少 META-INF/container.xml")
	}

	rc, err := f.Open()
	if err != nil {
		return "", fmt.Errorf("打开 container.xml 失败: %w", err)
	}
	defer rc.Close()

	var container epubContainer
	if err := xml.NewDecoder(rc).Decode(&container); err != nil {
		return "", fmt.Errorf("解析 container.xml 失败: %w", err)
	}

	for _, rf := range container.RootFiles {
		if rf.MediaType == "application/oebps-package+xml" {
			return rf.FullPath, nil
		}
	}

	if len(container.RootFiles) > 0 {
		return container.RootFiles[0].FullPath, nil
	}

	return "", fmt.Errorf("container.xml 中未找到 OPF 文件")
}

// --- OPF ---

type opfPackage struct {
	XMLName  xml.Name    `xml:"package"`
	Manifest opfManifest `xml:"manifest"`
	Spine    opfSpine    `xml:"spine"`
}

type opfManifest struct {
	Items []opfManifestItem `xml:"item"`
}

type opfManifestItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type opfSpine struct {
	TOC      string       `xml:"toc,attr"`
	ItemRefs []opfItemRef `xml:"itemref"`
}

type opfItemRef struct {
	IDRef  string `xml:"idref,attr"`
	Linear string `xml:"linear,attr"`
}

func parseOPF(f *zip.File) (*opfPackage, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("打开 OPF 文件失败: %w", err)
	}
	defer rc.Close()

	var pkg opfPackage
	if err := xml.NewDecoder(rc).Decode(&pkg); err != nil {
		return nil, fmt.Errorf("解析 OPF 文件失败: %w", err)
	}
	return &pkg, nil
}

// --- toc.ncx (EPUB 2) ---

type ncxDocument struct {
	XMLName  xml.Name    `xml:"ncx"`
	NavMap   ncxNavMap   `xml:"navMap"`
}

type ncxNavMap struct {
	NavPoints []ncxNavPoint `xml:"navPoint"`
}

type ncxNavPoint struct {
	Label   ncxNavLabel   `xml:"navLabel"`
	Content ncxContent    `xml:"content"`
}

type ncxNavLabel struct {
	Text string `xml:"text"`
}

type ncxContent struct {
	Src string `xml:"src,attr"`
}

func parseNCXTitles(f *zip.File, opfDir string) (map[string]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	var ncx ncxDocument
	if err := xml.NewDecoder(rc).Decode(&ncx); err != nil {
		return nil, err
	}

	titles := make(map[string]string)
	for _, np := range ncx.NavMap.NavPoints {
		src := np.Content.Src
		// 去掉锚点 (#section1)
		if idx := strings.Index(src, "#"); idx >= 0 {
			src = src[:idx]
		}
		title := strings.TrimSpace(np.Label.Text)
		if title != "" && src != "" {
			titles[src] = title
			titles[resolvePath(opfDir, src)] = title
		}
	}
	return titles, nil
}

// --- nav.xhtml (EPUB 3) ---

func parseNavTitles(f *zip.File, opfDir string) (map[string]string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	titles := make(map[string]string)
	tokenizer := html.NewTokenizer(rc)
	var inNav, inLink bool
	var currentHref, currentText string

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			break
		}

		switch tt {
		case html.StartTagToken:
			tn, hasAttr := tokenizer.TagName()
			tagName := string(tn)
			if tagName == "nav" {
				for hasAttr {
					var key, val []byte
					key, val, hasAttr = tokenizer.TagAttr()
					if string(key) == "epub:type" && string(val) == "toc" {
						inNav = true
					}
				}
			}
			if inNav && tagName == "a" {
				inLink = true
				for hasAttr {
					var key, val []byte
					key, val, hasAttr = tokenizer.TagAttr()
					if string(key) == "href" {
						href := string(val)
						if idx := strings.Index(href, "#"); idx >= 0 {
							href = href[:idx]
						}
						currentHref = href
					}
				}
			}
		case html.TextToken:
			if inLink {
				currentText += string(tokenizer.Text())
			}
		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			if string(tn) == "a" && inLink {
				title := strings.TrimSpace(currentText)
				if title != "" && currentHref != "" {
					titles[currentHref] = title
					titles[resolvePath(opfDir, currentHref)] = title
				}
				inLink = false
				currentHref = ""
				currentText = ""
			}
			if string(tn) == "nav" {
				inNav = false
			}
		}
	}

	return titles, nil
}

// --- XHTML 文本提取 ---

func extractTextFromXHTML(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	var sb strings.Builder
	tokenizer := html.NewTokenizer(rc)
	skipDepth := 0

	for {
		tt := tokenizer.Next()
		if tt == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			}
			return sb.String(), nil
		}

		switch tt {
		case html.StartTagToken:
			tn, _ := tokenizer.TagName()
			tagName := string(tn)
			if tagName == "script" || tagName == "style" {
				skipDepth++
			}
			if tagName == "p" || tagName == "div" || tagName == "br" ||
				tagName == "h1" || tagName == "h2" || tagName == "h3" ||
				tagName == "h4" || tagName == "h5" || tagName == "h6" {
				if sb.Len() > 0 {
					sb.WriteString("\n")
				}
			}
		case html.EndTagToken:
			tn, _ := tokenizer.TagName()
			tagName := string(tn)
			if (tagName == "script" || tagName == "style") && skipDepth > 0 {
				skipDepth--
			}
		case html.TextToken:
			if skipDepth == 0 {
				text := strings.TrimSpace(string(tokenizer.Text()))
				if text != "" {
					sb.WriteString(text)
				}
			}
		}
	}

	return sb.String(), nil
}

// --- 工具函数 ---

func resolvePath(base, href string) string {
	if path.IsAbs(href) {
		return href[1:]
	}
	return path.Join(base, href)
}
