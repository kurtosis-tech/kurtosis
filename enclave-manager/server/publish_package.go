package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type GitHubRepo struct {
	Owner struct {
		Login string `json:"login"`
	} `json:"owner"`
	Name string `json:"name"`
}

type BlobData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func exchangeCodeForToken(code string) (string, error) {
	url := "YOUR_BACKEND_URL/exchange_token" // Replace with your backend URL

	data := map[string]string{"code": code}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	accessToken, ok := result["access_token"]
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}

	return accessToken, nil
}

func createRepo(accessToken string, packageName string) (GitHubRepo, error) {
	url := "https://api.github.com/user/repos"
	data := map[string]interface{}{
		"name":    packageName,
		"private": false,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return GitHubRepo{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return GitHubRepo{}, err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GitHubRepo{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return GitHubRepo{}, err
	}

	var repo GitHubRepo
	if err := json.Unmarshal(body, &repo); err != nil {
		return GitHubRepo{}, err
	}

	return repo, nil
}

func createBlob(accessToken, owner, repo string, blobs []BlobData) (map[string]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/blobs", owner, repo)
	blobMap := make(map[string]map[string]interface{})
	for _, blob := range blobs {
		blobMap[blob.Path] = map[string]interface{}{
			"content":  base64.StdEncoding.EncodeToString([]byte(blob.Content)),
			"encoding": "base64",
		}
	}

	data := make([]map[string]interface{}, 0, len(blobMap))
	for path, blobData := range blobMap {
		data = append(data, map[string]interface{}{
			"path":     path,
			"content":  blobData["content"],
			"encoding": blobData["encoding"],
		})
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []map[string]string
	if err := json.Unmarshal(body, &results); err != nil {
		return nil, err
	}

	resultMap := make(map[string]string)
	for _, result := range results {
		resultMap[result["path"]] = result["sha"]
	}

	return resultMap, nil
}

// Add other functions (createTree, createCommit, updateReference) similarly...
func createTree(accessToken, owner, repo, baseTreeSha string, files []BlobData) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", owner, repo)

	tree := make([]map[string]interface{}, len(files))
	for i, file := range files {
		tree[i] = map[string]interface{}{
			"path": file.Path,
			"mode": "100644",
			"type": "blob",
			"sha":  file.Content,
		}
	}

	data := map[string]interface{}{
		"base_tree": baseTreeSha,
		"tree":      tree,
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	sha, ok := result["sha"]
	if !ok {
		return "", fmt.Errorf("sha not found in response")
	}

	return sha, nil
}

func createCommit(accessToken, owner, repo, message, treeSha, parentSha string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", owner, repo)

	data := map[string]interface{}{
		"message": message,
		"tree":    treeSha,
		"parents": []string{parentSha},
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	sha, ok := result["sha"]
	if !ok {
		return "", fmt.Errorf("sha not found in response")
	}

	return sha, nil
}

func updateReference(accessToken, owner, repo, commitSha string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/main", owner, repo)

	data := map[string]string{"sha": commitSha}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

func getBaseTreeSha(accessToken, owner, repo, branch string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", owner, repo, branch)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	commitSHA, ok := result["object"].(map[string]interface{})["sha"].(string)
	if !ok {
		return "", fmt.Errorf("commit SHA not found in response")
	}

	return commitSHA, nil
}

func getPackageReadMeContents(owner, repoName string) string {
	// TODO: turn this into a go template to parametrize owner and repo name
	readMeBytes, err := os.ReadFile("./README_.md.tmpl")
	if err != nil {
		panic(err)
	}
	return string(readMeBytes)
}
