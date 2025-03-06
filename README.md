# GorkbunDDNS
GorkbunDDNS is a dynamic DNS (DDNS) updater for [Porkbun↗](https://porkbun.com/), written in [Go↗](https://go.dev/). It is designed to automatically update your DNS records with your current WAN IP address, supporting both IPv4 and IPv6 addresses. With a valid configuration, GorkbunDDNS is built to run reliably without crashing, ensuring DNS records are always up-to-date.

## Getting Started
The preferred way to run GorkbunDDNS is via a Docker image. Follow the steps below to get started.

### Prerequisites

- [Docker↗](https://www.docker.com/get-started/) installed on your machine.

### Installation

1. **Pull the Docker image**
```console
docker pull puma0243/gorkbunddns:latest
```

2. **Run the Docker container**
```console
docker run -d \
  -e DOMAINS=example.com,sub.example.com,sub2.example.com \
  -e APIKEY=pk1_xyz \
  -e SECRETKEY=sk1_xyz \
  puma0243/gorkbunddns:latest
```

### Environment Variables
- `DOMAINS`: The domains to update. Format: `example.com,api.example.com,*.example.com` (required).
- `APIKEY`: Your Porkbun API key (required).
- `SECRETKEY`: Your Porkbun secret key (required).
+ `TIMEOUT`: The interval in seconds between DNS updates (optional, default is `600`).
+ `IPV4`: Enable or disable IPv4 updates (`true` or `false`, optional, default is `true`).
+ `IPV6`: Enable or disable IPv6 updates (`true` or `false`, optional, default is `false`).
+ `MULTIPLE_RECORDS`: How to handle multiple existing DNS records (`skip` or `unify`, optional, default is `skip`)

## Contribution
We welcome contributions to GorkbunDDNS! If you would like to contribute, please follow these steps:

Fork the repository:

Click the "Fork" button at the top right of this page to create a copy of this repository in your GitHub account.

Clone your fork:

Create a new branch:

Make your changes and commit them:

Push to the branch:

Create a pull request:

Open a pull request from your forked repository on GitHub to the main repository.

Please ensure your code follows the project's coding standards and includes appropriate tests.
