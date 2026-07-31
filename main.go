package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var port = "8081"

func main() {

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err.Error())
	}

	ln, err := net.Listen("tcp", ":"+ port)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf("Server is listening on port %s\n", port)

	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err.Error())
			break
		}
		log.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())

		go handleConnection(cwd, conn)

	}
}

func handleConnection(cwd string, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		queryMap := make(map[string]string)
		headerMap := make(map[string]string)

		requestLine, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		requestLine = strings.TrimRight(requestLine, "\r\n")
		requestParts := strings.Split(requestLine, " ")

		if len(requestParts) != 3 {
			HTTPMessage(conn, "400", "Bad Request", "Too many")
			return
		}
		if requestParts[2] != "HTTP/1.1" {
			HTTPMessage(conn, "400", "Bad Request", "Only HTTP/1.1 supported")
			return
		}
		if requestParts[0] != "GET" && requestParts[0] != "POST" {
			HTTPMessage(conn, "405", "Method Not Allowed", "Unsupported HTTP Method")
			return
		}

		relativeURI := requestParts[1]

		if strings.Contains(relativeURI, "?") {
			endpoint, queryStripped, found := strings.Cut(relativeURI, "?")
			if found {
				queryMap["endpoint"] = endpoint
			}

			parameters := queryStripped

			for value := range strings.SplitSeq(parameters, "&") {
				resultKV := strings.Split(value, "=")
				if len(resultKV) != 2 || slices.Contains(resultKV, "") {
					HTTPMessage(conn, "400", "Bad Request", "Malformed query")
					return
				}
				queryMap[resultKV[0]] = resultKV[1]
			}

		} else {
			queryMap["endpoint"] = relativeURI
		}

		for {
			headerLine, err := reader.ReadString('\n')
			if err != nil {
				log.Print(err.Error())
				HTTPMessage(conn, "400", "Bad Request")
				return
			}
			headerLine = strings.TrimRight(headerLine, "\r\n")
			if headerLine == "" {
				break
			}
			headerParts := strings.SplitN(headerLine, ": ", 2)
			headerMap[headerParts[0]] = headerParts[1]
		}

		if _, ok := headerMap["Content-Length"]; ok {
			strValue := headerMap["Content-Length"]
			intValue64, err := strconv.ParseInt(strValue, 10, 64)
			if err != nil {
				log.Printf("Could not convert string %q to int: %v", strValue, err)
				HTTPMessage(conn, "500", "Internal Server Error")
				return
			}

			if intValue64 > 0 && intValue64 <= 9999999 && requestParts[0] == "POST" {
				bodyBytes, err := io.ReadAll(io.LimitReader(reader, intValue64))
				if err != nil {
					log.Print(err.Error())
					HTTPMessage(conn, "400", "Bad Request")
					return
				}
				log.Printf("DEBUG: POST body: %s", bodyBytes)
			}

		}

		switch queryMap["endpoint"] {
		case "/":
			HTTPFileServe(cwd, conn, "200", "OK", "index.html")
		case "/styles.css":
			HTTPFileServe(cwd, conn, "200", "OK", "styles.css")
		case "/ping":
			HTTPMessage(conn, "200", "OK", "pong")
		default:
			HTTPMessage(conn, "404", "Not Found", "Endpoint Not Found")
		}

		if strings.ToLower(headerMap["Connection"]) == "close" {
			return
		}
	}
}

func HTTPMessage(myConnection net.Conn, statusCode string, statusPhrase string, messageBody ...string) {
	contentType := "text/plain"

	body := strings.Join(messageBody, "")
	body = body + "\n"
	contentLength := strconv.Itoa(len(body))

	serverResponse := "HTTP/1.1 " + statusCode + " " + statusPhrase + "\r\n" +
		"Date: " + time.Now().UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + "GoLang NixOS TCP/HTTP Engine" + "\r\n" +
		"Content-Length: " + contentLength + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"Connection: " + "keep-alive" + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}

func HTTPFileServe(cwd string, myConnection net.Conn, statusCode string, statusPhrase, filePath string) {
	var body string
	var contentType string

	body, contentType = fileReadingHelper(cwd, filePath)

	contentLength := strconv.Itoa(len(body))

	serverResponse := "HTTP/1.1 " + statusCode + " " + statusPhrase + "\r\n" +
		"Date: " + time.Now().UTC().Format(time.RFC1123) + "\r\n" +
		"Server: " + "GoLang NixOS TCP/HTTP Engine" + "\r\n" +
		"Content-Length: " + contentLength + "\r\n" +
		"Content-Type: " + contentType + "\r\n" +
		"Connection: " + "keep-alive" + "\r\n" +
		"\r\n" +
		body

	myConnection.Write([]byte(serverResponse))
}

func fileReadingHelper(cwd string, filePath string) (string, string) {
	fullPath := filepath.Join(cwd, filePath)
	basePrefix := cwd + string(filepath.Separator)

	if !strings.HasPrefix(fullPath, basePrefix) && fullPath != cwd {
		log.Println("Error: File path is outside the allowed directory")
		return "", ""
	}

	contentSlice := strings.SplitN(filePath, ".", 2)

	bodyBytes, err := os.ReadFile(fullPath)
	if err != nil {
		log.Print(err.Error())
		return "", ""
	}
	body := string(bodyBytes)

	contentTypeMap := map[string]string{
		"html": "text/html",
		"css":  "text/css",
		"js":   "application/javascript",
		"png":  "image/png",
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"gif":  "image/gif",
	}

	contentType := contentTypeMap[contentSlice[1]]
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return body, contentType
}
