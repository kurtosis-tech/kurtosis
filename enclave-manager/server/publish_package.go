package server

import (
	"context"
	"fmt"
	"github.com/google/go-github/v60/github"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"io"
	"net/http"
	"net/url"
)

const (
	clientID     = ""
	clientSecret = ""
)

type BlobData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func PublishPackageRepository(ctx context.Context, authCode, packageName, serializedStarlarkScript string, serializedPackageIcon []byte) error {
	logrus.Infof("Attempting to publish package using github code: %v, package name: %v", authCode, packageName)

	// Step 2: Exchange the code for an access token
	accessToken, _, err := exchangeCodeForToken(authCode)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving access token for publishing pakcage.")
	}
	logrus.Infof("Access Token for GitHub %v", accessToken)

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)
	logrus.Infof("Successfully created authed github client using access token: %v", accessToken)

	// list all repositories for the authenticated user
	repos, resp, err := ghClient.Repositories.ListByOrg(ctx, "kurtosis-tech", nil)
	if err != nil {
		return err
	}
	logrus.Infof("Users repositories: %v\n%v", repos, resp)
	user, resp, err := ghClient.Users.Get(ctx, "tedim52")
	if err != nil {
		return err
	}
	logrus.Infof("Information about our authed user: %v %v\n", user, resp)

	// Step 3: Create a new repository
	repo, err := createRepo(ctx, ghClient, accessToken, packageName)
	if err != nil {
		fmt.Println("Error creating repository:", err)
		return stacktrace.Propagate(err, "An error occurred creating repo for publishing package.")
	}
	logrus.Infof("Repo creation success: %v", repo)

	//// Step 4: Create blob data for files to be committed
	//files := []BlobData{
	//	{Path: "main.star", Content: serializedStarlarkScript},
	//	{Path: "kurtosis.yml", Content: fmt.Sprintf("name: github.com/%v/%v", repo.Owner, packageName)},
	//	{Path: "kurtosis-package-icon.png", Content: string(serializedPackageIcon)}, // TODO: this isn't gonna work figure it out
	//	{Path: "README_K.md", Content: getPackageReadMeContents(repo.Owner.Login, packageName) },
	//}
	//blobMap, err := createBlob(accessToken, repo.Owner.Login, packageName, files)
	//if err != nil {
	//	return nil, stacktrace.Propagate(err, "An error occurred creating blobs")
	//}
	//
	//baseTreeSha := "BASE_TREE_SHA" // Replace with the base tree SHA
	//treeSha, err := createTree(accessToken, repo.Owner.Login, packageName, baseTreeSha, files)
	//if err != nil {
	//	return nil, stacktrace.Propagate(err, "An error occurred creating tree.")
	//}
	//fmt.Println("Tree created. SHA:", treeSha)
	//
	//// Step 6: Create a commit
	//message := "Initial commit" // Replace with the commit message
	//commitSha, err := createCommit(accessToken, repo.Owner.Login, packageName, message, treeSha, baseTreeSha)
	//if err != nil {
	//	return nil, stacktrace.Propagate(err, "An error occurred creating commit to publish package.")
	//}
	//fmt.Println("Commit created. SHA:", commitSha)
	//
	//// Step 7: Update reference
	//err = updateReference(accessToken, repo.Owner.Login, packageName, commitSha)
	//if err != nil {
	//	return nil, stacktrace.Propagate("An error occurred updating reference.")
	//}
	return nil
}

func exchangeCodeForToken(code string) (string, string, error) {
	// Construct the URL for exchanging code for token
	tokenURL := fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", clientID, clientSecret, code)

	// Make a POST request to the token URL
	resp, err := http.PostForm(tokenURL, url.Values{})
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	logrus.Infof("access token request body: %v", string(body))

	// Parse the response to extract the access token
	values, err := url.ParseQuery(string(body))
	if err != nil {
		return "", "", err
	}

	accessToken := values.Get("access_token")
	if accessToken == "" {
		return "", "", fmt.Errorf("access_token not found in response")
	}

	scope := values.Get("scope")

	return accessToken, scope, nil
}

func createRepo(ctx context.Context, client *github.Client, accessToken string, packageName string) (*github.Repository, error) {
	//repo := &github.Repository{
	//	ID:                nil,
	//	NodeID:            nil,
	//	Owner:             nil,
	//	Name:              &packageName,
	//	FullName:          nil,
	//	Description:       nil,
	//	Homepage:          nil,
	//	CodeOfConduct:     nil,
	//	DefaultBranch:     nil,
	//	MasterBranch:      nil,
	//	CreatedAt:         nil,
	//	PushedAt:          nil,
	//	UpdatedAt:         nil,
	//	HTMLURL:           nil,
	//	CloneURL:          nil,
	//	GitURL:            nil,
	//	MirrorURL:         nil,
	//	SSHURL:            nil,
	//	SVNURL:            nil,
	//	Language:          nil,
	//	Fork:              nil,
	//	ForksCount:        nil,
	//	NetworkCount:      nil,
	//	OpenIssuesCount:   nil,
	//	StargazersCount:   nil,
	//	SubscribersCount:  nil,
	//	WatchersCount:     nil,
	//	Size:              nil,
	//	AutoInit:          nil,
	//	Parent:            nil,
	//	Source:            nil,
	//	Organization:      nil,
	//	Permissions:       nil,
	//	AllowRebaseMerge:  nil,
	//	AllowSquashMerge:  nil,
	//	AllowMergeCommit:  nil,
	//	Topics:            nil,
	//	License:           nil,
	//	Private:           nil,
	//	HasIssues:         nil,
	//	HasWiki:           nil,
	//	HasPages:          nil,
	//	HasProjects:       nil,
	//	HasDownloads:      nil,
	//	LicenseTemplate:   nil,
	//	GitignoreTemplate: nil,
	//	Archived:          nil,
	//	TeamID:            nil,
	//	URL:               nil,
	//	ArchiveURL:        nil,
	//	AssigneesURL:      nil,
	//	BlobsURL:          nil,
	//	BranchesURL:       nil,
	//	CollaboratorsURL:  nil,
	//	CommentsURL:       nil,
	//	CommitsURL:        nil,
	//	CompareURL:        nil,
	//	ContentsURL:       nil,
	//	ContributorsURL:   nil,
	//	DeploymentsURL:    nil,
	//	DownloadsURL:      nil,
	//	EventsURL:         nil,
	//	ForksURL:          nil,
	//	GitCommitsURL:     nil,
	//	GitRefsURL:        nil,
	//	GitTagsURL:        nil,
	//	HooksURL:          nil,
	//	IssueCommentURL:   nil,
	//	IssueEventsURL:    nil,
	//	IssuesURL:         nil,
	//	KeysURL:           nil,
	//	LabelsURL:         nil,
	//	LanguagesURL:      nil,
	//	MergesURL:         nil,
	//	MilestonesURL:     nil,
	//	NotificationsURL:  nil,
	//	PullsURL:          nil,
	//	ReleasesURL:       nil,
	//	StargazersURL:     nil,
	//	StatusesURL:       nil,
	//	SubscribersURL:    nil,
	//	SubscriptionURL:   nil,
	//	TagsURL:           nil,
	//	TreesURL:          nil,
	//	TeamsURL:          nil,
	//	TextMatches:       nil,
	//}
	//repoResult, resp, err := client.Repositories.Create(ctx, "", repo)
	//if err != nil {
	//	return nil, stacktrace.Propagate(err, "An error occurred creating repository.")
	//}
	packageTemplateRepoReq := &github.TemplateRepoRequest{
		Name:               &packageName,
		Owner:              nil,
		Description:        nil,
		IncludeAllBranches: nil,
		Private:            nil,
	}
	template, resp, err := client.Repositories.CreateFromTemplate(ctx, "kurtosis-tech", "package-template-repo", packageTemplateRepoReq)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating repository from template.")
	}
	logrus.Infof("Response body from create repo call %v", resp)
	logrus.Infof("Result from create repo call %v", template)
	return template, nil
}

//
//func createBlob(accessToken, owner, repo string, blobs []BlobData) (map[string]string, error) {
//	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/blobs", owner, repo)
//	blobMap := make(map[string]map[string]interface{})
//	for _, blob := range blobs {
//		blobMap[blob.Path] = map[string]interface{}{
//			"content":  base64.StdEncoding.EncodeToString([]byte(blob.Content)),
//			"encoding": "base64",
//		}
//	}
//
//	data := make([]map[string]interface{}, 0, len(blobMap))
//	for path, blobData := range blobMap {
//		data = append(data, map[string]interface{}{
//			"path":     path,
//			"content":  blobData["content"],
//			"encoding": blobData["encoding"],
//		})
//	}
//
//	payload, err := json.Marshal(data)
//	if err != nil {
//		return nil, err
//	}
//
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
//	if err != nil {
//		return nil, err
//	}
//	req.Header.Set("Authorization", "token "+accessToken)
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return nil, err
//	}
//	defer resp.Body.Close()
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return nil, err
//	}
//
//	var results []map[string]string
//	if err := json.Unmarshal(body, &results); err != nil {
//		return nil, err
//	}
//
//	resultMap := make(map[string]string)
//	for _, result := range results {
//		resultMap[result["path"]] = result["sha"]
//	}
//
//	return resultMap, nil
//}
//
//func createTree(accessToken, owner, repo, baseTreeSha string, files []BlobData) (string, error) {
//	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees", owner, repo)
//
//	tree := make([]map[string]interface{}, len(files))
//	for i, file := range files {
//		tree[i] = map[string]interface{}{
//			"path": file.Path,
//			"mode": "100644",
//			"type": "blob",
//			"sha":  file.Content,
//		}
//	}
//
//	data := map[string]interface{}{
//		"base_tree": baseTreeSha,
//		"tree":      tree,
//	}
//	payload, err := json.Marshal(data)
//	if err != nil {
//		return "", err
//	}
//
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
//	if err != nil {
//		return "", err
//	}
//	req.Header.Set("Authorization", "token "+accessToken)
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//	defer resp.Body.Close()
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//
//	var result map[string]string
//	if err := json.Unmarshal(body, &result); err != nil {
//		return "", err
//	}
//
//	sha, ok := result["sha"]
//	if !ok {
//		return "", fmt.Errorf("sha not found in response")
//	}
//
//	return sha, nil
//}
//
//func createCommit(accessToken, owner, repo, message, treeSha, parentSha string) (string, error) {
//	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/commits", owner, repo)
//
//	data := map[string]interface{}{
//		"message": message,
//		"tree":    treeSha,
//		"parents": []string{parentSha},
//	}
//	payload, err := json.Marshal(data)
//	if err != nil {
//		return "", err
//	}
//
//	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
//	if err != nil {
//		return "", err
//	}
//	req.Header.Set("Authorization", "token "+accessToken)
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//	defer resp.Body.Close()
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//
//	var result map[string]string
//	if err := json.Unmarshal(body, &result); err != nil {
//		return "", err
//	}
//
//	sha, ok := result["sha"]
//	if !ok {
//		return "", fmt.Errorf("sha not found in response")
//	}
//
//	return sha, nil
//}
//
//func updateReference(accessToken, owner, repo, commitSha string) error {
//	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/main", owner, repo)
//
//	data := map[string]string{"sha": commitSha}
//	payload, err := json.Marshal(data)
//	if err != nil {
//		return err
//	}
//
//	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
//	if err != nil {
//		return err
//	}
//	req.Header.Set("Authorization", "token "+accessToken)
//	req.Header.Set("Content-Type", "application/json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
//	}
//
//	return nil
//}
//
//func getBaseTreeSha(accessToken, owner, repo, branch string) (string, error) {
//	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/refs/heads/%s", owner, repo, branch)
//
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return "", err
//	}
//	req.Header.Set("Authorization", "token "+accessToken)
//	req.Header.Set("Accept", "application/vnd.github.v3+json")
//
//	client := &http.Client{}
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", err
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK {
//		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
//	}
//
//	var result map[string]interface{}
//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
//		return "", err
//	}
//
//	commitSHA, ok := result["object"].(map[string]interface{})["sha"].(string)
//	if !ok {
//		return "", fmt.Errorf("commit SHA not found in response")
//	}
//
//	return commitSHA, nil
//}
//
//func getPackageReadMeContents(owner, repoName string) string {
//	// TODO: turn this into a go template to parametrize owner and repo name
//	readMeBytes, err := os.ReadFile("./README_.md.tmpl")
//	if err != nil {
//		panic(err)
//	}
//	return string(readMeBytes)
//}
