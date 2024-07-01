package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	const filename = "newfile.csv"
	const buffSize = 100
	users := make(map[string]User)
	mainBuffer := make([]byte, buffSize)
	var tempBuffer []byte

	if _, err := os.Stat(filename); err == nil {
		readFile, err := os.Open(filename)

		if err != nil {
			log.Fatalf("Could not open file %s for read: %s\n", filename, err)
		}

		defer readFile.Close()

		reader := bufio.NewReader(readFile)

		for {
			n, err := reader.Read(mainBuffer)

			if err != nil {
				if err == io.EOF {
					break
				}

				log.Fatalf("Unable to read file: %s\n", err)
				break
			}

			mainBuffer := append(mainBuffer[:n])

			if tempBuffer != nil {
				firstNewlineIndex := firstIndexByte(mainBuffer, '\n')
				tempBuffer = append(tempBuffer, mainBuffer[:firstNewlineIndex+1]...)
				mainBuffer = mainBuffer[firstNewlineIndex+1:]

				tempBuffer, users = processBuffer(tempBuffer, users)
			}

			lastNewlineIndex := bytes.LastIndexByte(mainBuffer, '\n')

			if lastNewlineIndex != -1 {
				tempBuffer = append(mainBuffer[lastNewlineIndex+1:])
				mainBuffer = mainBuffer[:lastNewlineIndex+1]
			}

			mainBuffer, users = processBuffer(mainBuffer, users)

		}

		usersRaw := unpackMap(users)

		usersJson, err := json.MarshalIndent(usersRaw, "", "  ")
		if err != nil {
			log.Fatalf("Unable to convert map to JSON: %s\n", err)
		}

		fmt.Println(string(usersJson))
	}
}

func unpackMap(m map[string]User) map[string]map[string]string {
	mRaw := make(map[string]map[string]string)
	for key, user := range m {
		mRaw[key] = make(map[string]string)
		mRaw[key]["userName"] = user.userName
		mRaw[key]["firstName"] = user.firstName
		mRaw[key]["secondName"] = user.secondName
		mRaw[key]["email"] = user.email
	}

	return mRaw
}

func processBuffer(buffer []byte, users map[string]User) ([]byte, map[string]User) {
	csvReader := csv.NewReader(bytes.NewReader(buffer))

	rawData, err := csvReader.ReadAll()

	if err != nil {
		log.Fatalf("Unable to read CSV data: %s\n", err)
	}

	for _, row := range rawData {
		users[row[0]] = User{
			userName:   row[0],
			firstName:  row[1],
			secondName: row[2],
			email:      row[3],
		}
	}

	return nil, users
}

type User struct {
	userName   string
	firstName  string
	secondName string
	email      string
}

func displayUsers(m map[string]User) {
	for row := range m {
		fmt.Printf("Username: %s\n", row)
		fmt.Printf("  username: %s\n", m[row].userName)
		fmt.Printf("  name: %s\n", m[row].firstName)
		fmt.Printf("  second name: %s\n", m[row].secondName)
		fmt.Printf("  email: %s\n", m[row].email)
	}
}

func firstIndexByte(dataByte []byte, searchByte byte) int {
	for i, v := range dataByte {
		if v == searchByte {
			return i
		}
	}

	return -1
}
