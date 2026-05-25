package smtp

import (
	"io"
	"mime"
	"mime/multipart"
	netmail "net/mail"
	"strings"
)

// parsedMessage holds the extracted subject and HTML body from a raw SMTP DATA payload.
type parsedMessage struct {
	subject string
	body    string
}

// parseMessage reads the raw message, extracts the Subject header and HTML body.
func parseMessage(r io.Reader) (parsedMessage, error) {
	msg, err := netmail.ReadMessage(r)
	if err != nil {
		return parsedMessage{}, err
	}

	subject := msg.Header.Get("Subject")
	body, err := extractBody(msg)
	if err != nil {
		return parsedMessage{}, err
	}

	return parsedMessage{subject: subject, body: body}, nil
}

// extractBody returns the HTML body of the message if present,
// otherwise returns the plain text body wrapped in <pre> tags.
func extractBody(msg *netmail.Message) (string, error) {
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		return readPlain(msg.Body)
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return readPlain(msg.Body)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		return extractMultipart(msg.Body, params["boundary"])
	}

	if mediaType == "text/html" {
		return readAll(msg.Body)
	}

	return readPlain(msg.Body)
}

func extractMultipart(r io.Reader, boundary string) (string, error) {
	mr := multipart.NewReader(r, boundary)
	var plain string

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		ct := part.Header.Get("Content-Type")
		mediaType, _, parseErr := mime.ParseMediaType(ct)
		if parseErr != nil {
			continue
		}

		if mediaType == "text/html" {
			return readAll(part)
		}
		if mediaType == "text/plain" && plain == "" {
			var readErr error
			plain, readErr = readAll(part)
			if readErr != nil {
				return "", readErr
			}
		}
	}

	if plain != "" {
		return "<pre>" + plain + "</pre>", nil
	}
	return "", nil
}

func readPlain(r io.Reader) (string, error) {
	text, err := readAll(r)
	if err != nil {
		return "", err
	}
	return "<pre>" + text + "</pre>", nil
}

func readAll(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
