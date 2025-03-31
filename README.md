# SSH Proxy - Formal Assessment

## Description
This project implements an SSH proxy in Go that forwards communication between clients of the proxy and an upstream SSH server.
It logs client Stdin and delivers an LLM-based summary of each user's session and potential security risks involved.

## Video Demonstration

https://github.com/user-attachments/assets/cf4568f7-1833-468e-8c34-4f2abfa53b93

## Local Setup Guide

### Prerequisites
- Go 1.23+
- Docker 24.0+
- OpenAI API key

The exact keys used in this project's SSH servers are committed to the repository for ease of testing. Since this is a local application, there is little to no risk involved. However, in production, we would communicate these keys over a more secure file transfer/storage mechanism.
**On startup of the OpenSSH image, it will generate a new host key**. It uses the `ecdsa` key as its host, thus you can copy it from the terminal on image startup and overwrite the key at `upstream_auth/.ssh/known_hosts/ssh_host_ecdsa_key.pub`. **This is the only key that needs to be changed**.

### Config File
An example config.yaml file is provided in the project root, with all information filled except the llm API key. I used an OpenAI API key for this project.
Note that for keys, the path to the key is specified, and for passwords, a bcrypted password is used.

### Docker Compose
The application config and session logs are mounted to the upstream host as a volume.
Keys are also mounted into the proxy as a volume from the client machine.

## Application Build
To build the SSH proxy and upstream server, execute the following command in the project root directory:
```docker compose up --build```

This will build the SSH Proxy application on port 2022, and an OpenSSH server on port 2222.

## Design & Implementation Decisions

### Packages Selected
The proxy needs to be able to intercept client stdin and establish its own connection with the upstream server. Thus, it needs to set up an SSH server for the client to connect to, as well as a client to the upstream SSH server.
The `gliderlabs/ssh` package provides a streamlined, high-level API for setting up SSH servers. It greatly reduces the boilerplate code required to achieve this task.

It also came with build-in support for terminal window resizing, which simplified the logic required for this part of the code as well.

Next, the `golang.org/x/crypto` package was used for the proxy client, as the `gliderlabs/ssh` package only allows for setting up servers. This package is well-documented and frequently used for this task.

Lastly, I used the official `openai/openai-go` package for the LLM summary.

### Information Forwarding
I used the io.Copy() function in a goroutine to connect the streams between the client's proxy session and the proxy's upstream session. This approach efficiently forwards data between the streams, only copying when data is available, and runs on its own thread, preventing any blocking.
Additionally, I write to log files on the same io.Copy() thread as the client session's stdin. This ensures the log captures exactly what is being sent by the client.

### Security
In the configuration file, I store the paths to the SSH keys instead of the base64-enccoded values of the keys. 
I also bcrypt the passwords of allowed users. This helps reduce the risk of accidental data exposure and makes it easier to work with the keys, since there is no need to decode them each time they're read.

### LLM Prompt
When initially prompted to point out any security vulnerabilities, the LLM would be overly cautious and point potential flaws with commands like "echo hello" or "exit".
It was describing security vulnerabilities in every session. I tweaked the prompt and asked it to use more careful discretion, which, by inspection, seemed to improve its results.

## Testing
I created integration tests to ensure that the proxy achieves expected behaviour. I test the expected responses under RCE and Interactive Shell with single and multiple clients, to ensure it is able to handle multiple sessions at once.
To run these tests, first start the application by running ```docker compose up --build```. Then, from the root directory, execute `go test ./cmd/test`





