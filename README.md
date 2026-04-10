[![Docker](https://github.com/Ogglord/comp2unraid/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/Ogglord/comp2unraid/actions/workflows/docker-publish.yml) [![Go](https://github.com/Ogglord/comp2unraid/actions/workflows/go.yml/badge.svg)](https://github.com/Ogglord/comp2unraid/actions/workflows/go.yml)

# comp2unraid

Convert Docker Compose files to Unraid container templates (XML).

## Quick start

### From a URL

```bash
docker run ghcr.io/ogglord/comp2unraid -n \
  "https://raw.githubusercontent.com/Ogglord/comp2unraid/main/examples/compose/docker-compose.yml"
```

### From a local file (piped via stdin)

```bash
cat docker-compose.yml | docker run -i ghcr.io/ogglord/comp2unraid -n -
```

Both examples use `-n` (dry run) to preview the XML on stdout. Remove `-n` to write XML files to disk (one per service).

## Flags

| Flag | Description |
|------|-------------|
| `-n` | Dry run — output XML to stdout |
| `-f` | Overwrite existing XML files |
| `-e` | Include host environment variables and `.env` file |
| `-v` | Verbose output |

## Installing templates on Unraid

Unraid picks up user templates from `/boot/config/plugins/dockerMan/templates-user/`. Save your generated XML files there and they will appear in the Unraid Docker UI under **Add Container > Template**.

```bash
docker run ghcr.io/ogglord/comp2unraid -n \
  "https://raw.githubusercontent.com/Ogglord/comp2unraid/main/examples/compose/docker-compose.yml" \
  > /boot/config/plugins/dockerMan/templates-user/my-template.xml
```

Note: If your compose file has multiple services, run without `-n` to generate one XML file per service, then copy them into the templates directory.

## Build from source

```bash
git clone https://github.com/Ogglord/comp2unraid.git
cd comp2unraid
make
./bin/comp2unraid_linux_amd64 -n examples/compose/docker-compose.yml
```
