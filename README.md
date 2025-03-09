# GorkbunDDNS
GorkbunDDNS is a dynamic DNS (DDNS) updater <a href="http://example.com/" target="_blank">Hello, world!</a> for [Porkbun↗](https://porkbun.com/), written in [Go↗](https://go.dev/). It is designed to automatically update your DNS records with your current WAN IP address, supporting both IPv4 and IPv6 addresses. With a valid configuration, GorkbunDDNS is built to run reliably without crashing, ensuring DNS records are always up-to-date.

## Getting Started
The preferred way to run GorkbunDDNS is via a Docker image. Follow the steps below to get started.

### Prerequisites

- [Docker↗](https://www.docker.com/get-started/) installed on your machine.
- <details>
    <summary>API access is enabled for domains you want to update.</summary>
  
    API access can be managed on [Porkbun's Domain management site↗](https://porkbun.com/account/domainsSpeedy):
    
    ![api_access](https://github.com/user-attachments/assets/fa4cb507-f41c-406a-86dd-cecbc535c8e3)
  </details>
- You have a valid API and secret key pair.
> [!NOTE]
> You can generate a new API key and secret key pair at [Porkbun's API management page↗](https://porkbun.com/account/api).

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

### Configuration
The program is configurable through **environment variables**:
- `DOMAINS`: The domains to update. Format: `example.com,api.example.com,*.example.com` (required).
- `APIKEY`: Your Porkbun API key (required).
- `SECRETKEY`: Your Porkbun secret key (required).
+ `TIMEOUT`: The interval in seconds between DNS updates (optional, default is `600`).
+ `IPV4`: Enable or disable IPv4 updates (`true` or `false`, optional, default is `true`).
+ `IPV6`: Enable or disable IPv6 updates (`true` or `false`, optional, default is `false`).
+ `MULTIPLE_RECORDS`: How to handle multiple existing DNS records (`skip` or `unify`, optional, default is `skip`)
