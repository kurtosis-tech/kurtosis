import React, {ChangeEvent, useState} from 'react';
import {
    ChakraProvider,
    Box,
    Button,
    FormControl,
    FormLabel,
    Input,
    Modal,
    ModalOverlay,
    ModalContent,
    ModalHeader,
    ModalFooter,
    ModalBody,
    ModalCloseButton,
    useDisclosure,
    Image
} from '@chakra-ui/react';
import {string} from "yaml/dist/schema/common/string";

async function exchangeCodeForToken(code: string): Promise<string> {
    const response = await fetch('YOUR_BACKEND_URL/exchange_token', { // Replace with your backend URL
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ code })
    });
    const data = await response.json();
    return data.access_token;
}

interface GitHubRepo {
    owner: { login: string };
    name: string;
}

async function createRepo(accessToken: string): Promise<GitHubRepo> {
    const response = await fetch('https://api.github.com/user/repos', {
        method: 'POST',
        headers: {
            'Authorization': `token ${accessToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            name: 'new-repo-name',
            private: false
        })
    });
    return await response.json();
}

interface BlobData {
    path: string;
    sha: string;
}

async function createBlob(accessToken: string, owner: string, repo: string, content: string): Promise<string> {
    const response = await fetch(`https://api.github.com/repos/${owner}/${repo}/git/blobs`, {
        method: 'POST',
        headers: {
            'Authorization': `token ${accessToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            content: btoa(content),
            encoding: 'base64'
        })
    });
    const data = await response.json();
    return data.sha;
}

async function createTree(accessToken: string, owner: string, repo: string, baseTreeSha: string, files: BlobData[]): Promise<string> {
    const tree = files.map(file => ({
        path: file.path,
        mode: '100644',
        type: 'blob',
        sha: file.sha
    }));

    const response = await fetch(`https://api.github.com/repos/${owner}/${repo}/git/trees`, {
        method: 'POST',
        headers: {
            'Authorization': `token ${accessToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            base_tree: baseTreeSha,
            tree: tree
        })
    });
    const data = await response.json();
    return data.sha;
}

async function createCommit(accessToken: string, owner: string, repo: string, message: string, treeSha: string, parentSha: string): Promise<string> {
    const response = await fetch(`https://api.github.com/repos/${owner}/${repo}/git/commits`, {
        method: 'POST',
        headers: {
            'Authorization': `token ${accessToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            message: message,
            tree: treeSha,
            parents: [parentSha]
        })
    });
    const data = await response.json();
    return data.sha;
}

async function updateReference(accessToken: string, owner: string, repo: string, commitSha: string): Promise<void> {
    await fetch(`https://api.github.com/repos/${owner}/${repo}/git/refs/heads/main`, {
        method: 'PATCH',
        headers: {
            'Authorization': `token ${accessToken}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            sha: commitSha
        })
    });
}

type PublishRepoModalProps = {
    isOpen: boolean;
    onClose: () => void;
    code: string;
    starlark?: string;
};

export const PublishRepoModal = ({ isOpen, onClose, code, starlark }: PublishRepoModalProps) => {
    const [repoName, setRepoName] = useState<string>('basic-package');
    const [image, setImage] = useState<string | null>(null);

    const handleImageChange = (e: ChangeEvent<HTMLInputElement>) => {
        if (e.target.files && e.target.files[0]) {
            setImage(URL.createObjectURL(e.target.files[0]));
        }
    };

    const handlePublishSubmit = async() => {
        // Handle the form submission
        console.log('Repository Name:', repoName);
        console.log('Uploaded Image:', image);

        const accessToken = await exchangeCodeForToken(code);
        const repo = await createRepo(accessToken);
        const owner = repo.owner.login;

        const refResponse = await fetch(`https://api.github.com/repos/${owner}/${repoName}/git/refs/heads/main`, {
            headers: {
                'Authorization': `token ${accessToken}`
            }
        });
        const refData = await refResponse.json();
        const baseTreeSha = refData.object.sha;

        const fileContents: Record<string, string> = {
            'path/to/file1.txt': 'Content of file 1',
            'path/to/file2.txt': 'Content of file 2'
        };

        const blobs: BlobData[] = [];
        for (const path in fileContents) {
            const sha = await createBlob(accessToken, owner, repoName, fileContents[path]);
            blobs.push({ path, sha });
        }
        const treeSha = await createTree(accessToken, owner, repoName, baseTreeSha, blobs);
        const commitSha = await createCommit(accessToken, owner, repoName, 'Initial commit', treeSha, baseTreeSha);
        await updateReference(accessToken, owner, repoName, commitSha);
        console.log('Repository created and files committed successfully.');
    };

    return (
            <Modal isOpen={isOpen} onClose={onClose}>
                <ModalOverlay />
                <ModalContent>
                    <ModalHeader>Enter Repository Details</ModalHeader>
                    <ModalCloseButton />
                    <ModalBody>
                        <FormControl id="repo-name" isRequired>
                            <FormLabel>Repository Name</FormLabel>
                            <Input
                                placeholder="Enter repository name"
                                value={repoName}
                                onChange={(e) => setRepoName(e.target.value)}
                            />
                        </FormControl>
                        <FormControl id="image-upload" mt={4}>
                            <FormLabel>Upload Picture/Icon</FormLabel>
                            <Input type="file" accept="image/*" onChange={handleImageChange} />
                            {image && (
                                <Box mt={4}>
                                    <Image src={image} alt="Uploaded image" boxSize="100px" />
                                </Box>
                            )}
                        </FormControl>
                    </ModalBody>
                    <ModalFooter>
                        <Button colorScheme="blue" mr={3} onClick={handlePublishSubmit}>
                            Publish Package
                        </Button>
                        <Button onClick={onClose}>Close</Button>
                    </ModalFooter>
                </ModalContent>
            </Modal>
    );
}
