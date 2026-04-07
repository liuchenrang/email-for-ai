package mime

import (
	"encoding/base64"
	"io"
	"mime"
	"mime/quotedprintable"
	"strings"
	"time"
)

// DecodeHeader 解码邮件头（支持RFC 2047编码）
// 格式: =?charset?encoding?encoded_text?=
func DecodeHeader(header string) string {
	// 使用mime包的内置解码
	decoder := new(mime.WordDecoder)
	decoded, err := decoder.DecodeHeader(header)
	if err != nil {
		// 如果解码失败，尝试手动解码
		return decodeHeaderManual(header)
	}
	return decoded
}

// decodeHeaderManual 手动解码邮件头
func decodeHeaderManual(header string) string {
	if !strings.Contains(header, "=?") {
		return header
	}

	result := header

	// 处理所有编码部分
	for {
		start := strings.Index(result, "=?")
		if start == -1 {
			break
		}

		end := strings.Index(result[start+2:], "?=")
		if end == -1 {
			break
		}
		end += start + 2

		encoded := result[start : end+2]
		decoded := decodeEncodedWord(encoded)

		result = result[:start] + decoded + result[end+2:]
	}

	return result
}

// decodeEncodedWord 解码单个编码词
func decodeEncodedWord(word string) string {
	// 格式: =?charset?encoding?encoded_text?=
	if len(word) < 4 || !strings.HasPrefix(word, "=?") || !strings.HasSuffix(word, "?=") {
		return word
	}

	// 去掉前后的=?和?=
	content := word[2 : len(word)-2]

	// 分割: charset?encoding?text
	parts := strings.SplitN(content, "?", 3)
	if len(parts) != 3 {
		return word
	}

	encoding := strings.ToUpper(parts[1])
	encodedText := parts[2]

	switch encoding {
	case "B": // Base64
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedText)
		if err != nil {
			return word
		}
		return string(decodedBytes)

	case "Q": // Quoted-Printable
		decoded, err := decodeQuotedPrintable(encodedText)
		if err != nil {
			return word
		}
		return decoded

	default:
		return word
	}
}

// decodeQuotedPrintable 解码Quoted-Printable编码
func decodeQuotedPrintable(s string) (string, error) {
	s = strings.ReplaceAll(s, "_", " ")
	reader := quotedprintable.NewReader(strings.NewReader(s))
	result, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

// DecodeSubject 解码邮件主题
func DecodeSubject(subject string) string {
	return DecodeHeader(subject)
}

// DecodeAddress 解码邮件地址
func DecodeAddress(addr string) string {
	return DecodeHeader(addr)
}

// DecodeBody 解码邮件正文
func DecodeBody(body string, encoding string) string {
	switch strings.ToUpper(encoding) {
	case "BASE64", "B":
		decoded, err := base64.StdEncoding.DecodeString(body)
		if err != nil {
			return body
		}
		return string(decoded)

	case "QUOTED-PRINTABLE", "Q":
		decoded, err := decodeQuotedPrintable(body)
		if err != nil {
			return body
		}
		return decoded

	default:
		return body
	}
}

// ParseDate 解析邮件日期
func ParseDate(dateStr string) string {
	// 常见邮件日期格式
	formats := []string{
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 -0700",
		"Mon, 02 Jan 2006 15:04:05 MST",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t.Format("2006-01-02 15:04")
		}
	}

	return dateStr
}

// ParseMultipartBody 解析MIME multipart邮件正文，提取纯文本
func ParseMultipartBody(rawBody string) (textBody string, htmlBody string) {
	// 检测是否是multipart格式（以--开头的boundary分隔符）
	if !strings.HasPrefix(rawBody, "--") && !strings.Contains(rawBody, "\n--") {
		// 可能是简单的base64编码正文
		return tryDecodeSimpleBody(rawBody), ""
	}

	// 直接解析boundary分隔符
	return parseMultipartByDelimiter(rawBody)
}

// parseMultipartByDelimiter 通过分隔符解析multipart
func parseMultipartByDelimiter(rawBody string) (textBody string, htmlBody string) {
	// 找到第一个boundary行作为分隔符
	lines := strings.Split(rawBody, "\n")
	var delimiter string

	// 第一行通常是boundary分隔符（格式: --boundary_name）
	// 注意：boundary本身可能包含--前缀，所以整行就是分隔符
	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")
		// 检测以--开头的boundary分隔符（至少2个-）
		if strings.HasPrefix(line, "--") && len(line) > 2 {
			// 去掉可能的结束标记（末尾的--）
			fullDelimiter := strings.TrimSuffix(line, "--")
			// 检查是否是有效的分隔符（不能全是-）
			if !isAllDashes(fullDelimiter) && len(fullDelimiter) > 2 {
				delimiter = fullDelimiter
				break
			}
		}
	}

	if delimiter == "" {
		return tryDecodeSimpleBody(rawBody), ""
	}

	// 使用delimiter分割各部分
	parts := strings.Split(rawBody, delimiter)

	for _, part := range parts {
		part = strings.TrimPrefix(part, "\r\n")
		part = strings.TrimPrefix(part, "\n")

		// 跳过空部分和结束分隔符
		if strings.TrimSpace(part) == "" || strings.HasPrefix(part, "--") {
			continue
		}

		// 解析每个part
		partText, partHTML := parseMimePart(part)
		if partText != "" && textBody == "" {
			textBody = partText
		}
		if partHTML != "" && htmlBody == "" {
			htmlBody = partHTML
		}
	}

	return textBody, htmlBody
}

// isAllDashes 检查字符串是否全是破折号
func isAllDashes(s string) bool {
	for _, c := range s {
		if c != '-' {
			return false
		}
	}
	return true
}

// parseMimePart 解析单个MIME部分
func parseMimePart(part string) (textBody string, htmlBody string) {
	// 分离头部和内容
	var headerEnd int
	part = strings.TrimPrefix(part, "\r\n")
	part = strings.TrimPrefix(part, "\n")

	if idx := strings.Index(part, "\r\n\r\n"); idx != -1 {
		headerEnd = idx
	} else if idx := strings.Index(part, "\n\n"); idx != -1 {
		headerEnd = idx
	}

	if headerEnd <= 0 {
		return "", ""
	}

	header := part[:headerEnd]
	content := part[headerEnd:]
	content = strings.TrimPrefix(content, "\r\n\r\n")
	content = strings.TrimPrefix(content, "\n\n")
	content = strings.TrimSuffix(content, "\r\n")
	content = strings.TrimSuffix(content, "\n")

	// 解析头部获取Content-Type和编码
	headerLines := strings.Split(header, "\n")
	var contentType, encoding string

	for _, line := range headerLines {
		originalLine := line
		line = strings.TrimSpace(line)
		lineLower := strings.ToLower(line)

		if strings.HasPrefix(lineLower, "content-type:") {
			// 使用原始行的正确偏移量
			colonIdx := strings.Index(originalLine, ":")
			if colonIdx != -1 {
				contentType = strings.TrimSpace(originalLine[colonIdx+1:])
			}
		}
		if strings.HasPrefix(lineLower, "content-transfer-encoding:") {
			colonIdx := strings.Index(originalLine, ":")
			if colonIdx != -1 {
				encoding = strings.TrimSpace(originalLine[colonIdx+1:])
			}
		}
	}

	// 解码内容
	decodedContent := DecodeBody(strings.TrimSpace(content), encoding)

	// 根据Content-Type分类
	contentTypeLower := strings.ToLower(contentType)
	if strings.Contains(contentTypeLower, "text/plain") {
		textBody = decodedContent
	} else if strings.Contains(contentTypeLower, "text/html") {
		htmlBody = decodedContent
	}

	return textBody, htmlBody
}

// tryDecodeSimpleBody 尝试解码简单正文
func tryDecodeSimpleBody(body string) string {
	body = strings.TrimSpace(body)

	// 尝试base64解码
	decoded, err := base64.StdEncoding.DecodeString(body)
	if err == nil {
		return string(decoded)
	}

	// 尝试quoted-printable解码
	if strings.Contains(body, "=") {
		reader := quotedprintable.NewReader(strings.NewReader(body))
		result, err := io.ReadAll(reader)
		if err == nil {
			return string(result)
		}
	}

	// 返回原始内容
	return body
}

// ExtractReadableBody 从原始邮件内容提取可读正文
func ExtractReadableBody(rawContent string) string {
	// 先尝试解析multipart
	textBody, htmlBody := ParseMultipartBody(rawContent)

	// 优先返回纯文本
	if textBody != "" {
		return textBody
	}

	// 如果只有HTML，简单处理一下
	if htmlBody != "" {
		// 移除HTML标签，提取纯文本（简单处理）
		return stripHTMLTags(htmlBody)
	}

	// 无法解析，尝试直接解码
	return tryDecodeSimpleBody(rawContent)
}

// stripHTMLTags 简单移除HTML标签
func stripHTMLTags(html string) string {
	// 简单的HTML标签移除
	result := html
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}

	// 处理常见HTML实体
	result = strings.ReplaceAll(result, "&nbsp;", " ")
	result = strings.ReplaceAll(result, "&lt;", "<")
	result = strings.ReplaceAll(result, "&gt;", ">")
	result = strings.ReplaceAll(result, "&amp;", "&")
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "<p>", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n")

	return strings.TrimSpace(result)
}