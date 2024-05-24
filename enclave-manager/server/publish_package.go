package server

import (
	"context"
	"encoding/base64"
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
	clientID     = "Iv23liicwMSrJ7dqrqdO"
	clientSecret = ""
)

type BlobData struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

func PublishPackageRepository(ctx context.Context, authCode, packageName, serializedStarlarkScript string, serializedPackageIcon []byte) error {
	logrus.Infof("Attempting to publish package using github code: %v, package name: %v", authCode, packageName)

	// Step 1: Exchange the code for an access token
	accessToken, _, err := exchangeCodeForToken(authCode)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred retrieving access token for publishing pakcage.")
	}
	logrus.Infof("Access Token for GitHub %v", accessToken)
	//accessToken := ""

	// Step 2: create ghclient authed with user acccess
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	ghClient := github.NewClient(tc)
	logrus.Infof("Successfully created authed github client using access token: %v", accessToken)

	// Step 3: Create a new repository
	repo, err := createRepo(ctx, ghClient, accessToken, packageName)
	if err != nil {
		fmt.Println("Error creating repository:", err)
		return stacktrace.Propagate(err, "An error occurred creating repo for publishing package.")
	}
	logrus.Infof("Repo creation success: %v", repo)

	// Step 4: Create blob data for files to be committed
	files := []BlobData{
		{Path: "main.star", Content: serializedStarlarkScript},
		{Path: "kurtosis.yml", Content: fmt.Sprintf("name: github.com/%v/%v", "tedim52", "basic-package")},
		{Path: "kurtosis-package-icon.png", Content: string(serializedPackageIcon)},
		//{Path: "README_K.md", Content: getPackageReadMeContents(repo.Owner.Login, packageName) },
	}

	baseTreeSha, err := getBaseTreeSha(ctx, ghClient, "tedim52", packageName, "main")
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred getting base tree")
	}
	logrus.Infof("Base tree SHA successfully retrieved: %v", baseTreeSha)

	treeSha, err := createTree(ctx, ghClient, accessToken, "tedim52", packageName, baseTreeSha, files)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating tree.")
	}
	logrus.Infof("Tree successfully created: %v", treeSha)

	// Step 6: Create a commit
	message := "Initial commit" // Replace with the commit message
	//commitSha, err := createCommit(accessToken, repo.Owner.Login, packageName, message, treeSha, baseTreeSha)
	commitSha, err := createCommit(ctx, ghClient, "tedim52", packageName, message, treeSha, baseTreeSha)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred creating commit to publish package.")
	}
	fmt.Println("Commit created. SHA:", commitSha)

	// Step 7: Update reference
	// err = updateReference(ctx, ghClient, repo.Owner.Login, packageName, commitSha)
	err = updateReference(ctx, ghClient, "tedim52", packageName, commitSha)
	if err != nil {
		return stacktrace.Propagate(err, "An error occurred updating reference.")
	}
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
	packageTemplateRepoReq := &github.TemplateRepoRequest{
		Name:               &packageName,
		Owner:              nil,
		Description:        nil,
		IncludeAllBranches: nil,
		Private:            nil,
	}
	template, _, err := client.Repositories.CreateFromTemplate(ctx, "kurtosis-tech", "package-template-repo", packageTemplateRepoReq)
	if err != nil {
		return nil, stacktrace.Propagate(err, "An error occurred creating repository from template.")
	}
	//logrus.Infof("Response body from create repo call %v", resp)
	//logrus.Infof("Result from create repo call %v", template)
	return template, nil
}

func createBlob(ctx context.Context, client *github.Client, accessToken, owner, repo string, blobs []BlobData) (map[string]string, error) {
	// Create a map to store the results
	resultMap := make(map[string]string)

	// Iterate over the blobs and create them
	for _, blob := range blobs {
		encodedContent := base64.StdEncoding.EncodeToString([]byte(blob.Content))

		// Create a new blob object
		gitBlob := &github.Blob{
			Content:  github.String(encodedContent),
			Encoding: github.String("base64"),
		}

		// Create the blob in the repository
		createdBlob, _, err := client.Git.CreateBlob(ctx, owner, repo, gitBlob)
		if err != nil {
			return nil, stacktrace.Propagate(err, "An error occurred creating blocb at path %v", blob.Path)
		}

		// Store the SHA of the created blob in the result map
		resultMap[blob.Path] = createdBlob.GetSHA()
	}

	return resultMap, nil
}

func createTree(ctx context.Context, client *github.Client, accessToken, owner, repo, baseTreeSha string, files []BlobData) (string, error) {
	// Create a slice to hold tree entries
	var entries []*github.TreeEntry

	// Create blobs and add them to the tree
	for _, file := range files {
		// Create a new blob object
		encodedContent := base64.StdEncoding.EncodeToString([]byte(file.Content))
		blob := &github.Blob{
			Content:  github.String(encodedContent),
			Encoding: github.String("base64"),
		}

		// Create the blob in the repository
		createdBlob, _, err := client.Git.CreateBlob(ctx, owner, repo, blob)
		if err != nil {
			return "", err
		}

		// Create a new tree entry for the blob
		entry := &github.TreeEntry{
			Path: github.String(file.Path),
			Mode: github.String("100644"), // File mode
			Type: github.String("blob"),
			SHA:  createdBlob.SHA,
		}

		entries = append(entries, entry)
	}

	// Create the tree with the new entries
	tree, _, err := client.Git.CreateTree(ctx, owner, repo, baseTreeSha, entries)
	if err != nil {
		return "", err
	}

	return tree.GetSHA(), nil
}

func createCommit(ctx context.Context, client *github.Client, owner, repo, message, treeSha, parentSha string) (string, error) {
	commit := &github.Commit{
		Message: github.String(message),
		Tree:    &github.Tree{SHA: github.String(treeSha)},
		Parents: []*github.Commit{{SHA: github.String(parentSha)}},
	}
	createCommitOpts := &github.CreateCommitOptions{}
	createdCommit, _, err := client.Git.CreateCommit(ctx, owner, repo, commit, createCommitOpts)
	if err != nil {
		return "", err
	}
	return createdCommit.GetSHA(), nil
}

func updateReference(ctx context.Context, client *github.Client, owner, repo, commitSha string) error {
	ref := "refs/heads/main"
	force := true

	// Create a new reference object
	gitRef := &github.Reference{
		Ref:    github.String(ref),
		Object: &github.GitObject{SHA: github.String(commitSha)},
	}

	// Update the reference in the repository
	_, _, err := client.Git.UpdateRef(ctx, owner, repo, gitRef, force)
	if err != nil {
		return err
	}

	return nil
}

func getBaseTreeSha(ctx context.Context, client *github.Client, owner, repo, branch string) (string, error) {
	ref := fmt.Sprintf("refs/heads/%s", branch)

	// Get the reference for the branch
	gitRef, _, err := client.Git.GetRef(ctx, owner, repo, ref)
	if err != nil {
		return "", err
	}

	if gitRef.Object == nil || gitRef.Object.SHA == nil {
		return "", fmt.Errorf("commit SHA not found in response")
	}

	return *gitRef.Object.SHA, nil
}

//func getPackageReadMeContents(owner, repoName string) string {
//	// TODO: turn this into a go template to parametrize owner and repo name
//	readMeBytes, err := os.ReadFile("./README_.md.tmpl")
//	if err != nil {
//		panic(err)
//	}
//	return string(readMeBytes)
//}
