package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/ssh/terminal"
)

// Set the constants for this project
const apiPath string = "https://api.github.com"

func credentials() (string, string) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, _ := reader.ReadString('\n')

	fmt.Print("Enter Password: ")
	bytePassword, _ := terminal.ReadPassword(0)
	password := string(bytePassword)

	return strings.TrimSpace(username), strings.TrimSpace(password)
}

func main() {
	fmt.Println("Github Collaborators Management Console")

	fmt.Println("Press 1 to get all collaborators, 2 to remove a collaborator")
	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		userName, password := credentials()
		b, err := json.MarshalIndent(getAllCollaborators(userName, password), "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Print(string(b))
		break
	case 2:
		userName, password := credentials()
		removeCollaboratorFromAllRepos(userName, password)
		break
	default:
		panic("Incorrect option")
	}
}

// Remove a given collaborator from all repositories
func removeCollaboratorFromAllRepos(userName string, password string) {
	fmt.Println("\nFetching all the private repositories:")
	collabs := getAllCollaborators(userName, password)
	b, err := json.MarshalIndent(collabs, "", "  ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Print(string(b))

	fmt.Println("\nEnter username to delete:")
	var userNameToDelete string
	fmt.Scanln(&userNameToDelete)

	reposToCheck := collabs[userNameToDelete]
	fmt.Println("\nRemoving " + userNameToDelete + " from repos " + strings.Join(reposToCheck, ","))

	for i := 0; i < len(reposToCheck); i++ {
		getAPIPath := fmt.Sprintf("%s/repos/%s/collaborators/%s",
			apiPath, reposToCheck[i], userNameToDelete)
		fmt.Println("Calling delete on " + getAPIPath)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", getAPIPath, nil)
		req.SetBasicAuth(userName, password)
		req.Header.Add("Accept", "application/vnd.github.v3+json")
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
	}
}

// Get all collaborators for public project of this userName
func getAllCollaborators(userName string, password string) map[string][]string {
	const getAPIPath string = apiPath + "/user/repos?visibility=private&affiliation=owner"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", getAPIPath, nil)
	req.SetBasicAuth(userName, password)
	req.Header.Add("Accept", "application/vnd.github.v3+json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data []map[string]interface{}
	decodeErr := decoder.Decode(&data)
	if decodeErr != nil {
		fmt.Println(decodeErr)
	}

	// For all the repos, find the collaborators and store it in a map
	collabs := make(map[string][]string)
	for i := 0; i < len(data); i++ {
		name := data[i]["full_name"]
		collabURL := data[i]["collaborators_url"]
		collabURLStr := strings.Split(collabURL.(string), "{")[0]

		// Call the collaborator's url
		req, _ = http.NewRequest("GET", collabURLStr, nil)
		req.SetBasicAuth(userName, password)
		req.Header.Add("Accept", "application/vnd.github.v3+json")
		resp, _ = client.Do(req)

		// bodyText, _ := ioutil.ReadAll(resp.Body)
		// s := string(bodyText)
		// fmt.Println(s)
		collabDecoder := json.NewDecoder(resp.Body)
		var collabData []map[string]interface{}
		decodeErr = collabDecoder.Decode(&collabData)
		if decodeErr != nil {
			fmt.Println(decodeErr)
		}
		// Loop over all collaborators for this repo
		for c := 0; c < len(collabData); c++ {
			collab := collabData[c]["login"].(string)
			// Get this collaborator from the map, and append this repos name to it
			collabs[collab] = append(collabs[collab], name.(string))
		}
	}
	return collabs
}
