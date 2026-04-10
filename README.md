[![Docker](https://github.com/Ogglord/comp2unraid/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/Ogglord/comp2unraid/actions/workflows/docker-publish.yml) [![Go](https://github.com/Ogglord/comp2unraid/actions/workflows/go.yml/badge.svg)](https://github.com/Ogglord/comp2unraid/actions/workflows/go.yml)

# comp2unraid

Convert Docker Compose files to Unraid container templates (XML).

By default, XML is printed to stdout. Use `-w` to write individual XML files to disk (one per service).

## Usage

### From a URL

```bash
docker run ghcr.io/ogglord/comp2unraid \
  "https://raw.githubusercontent.com/Ogglord/comp2unraid/main/examples/compose/docker-compose.yml"
```

### From a local file (piped via stdin)

```bash
cat docker-compose.yml | docker run -i ghcr.io/ogglord/comp2unraid -
```

### Write XML files to disk

Mount your current directory to `/output` in the container and use `-w`:

```bash
cat docker-compose.yml | docker run -i -v .:/output ghcr.io/ogglord/comp2unraid -w -
```

This creates one `.xml` file per service in your current directory.

## Flags

| Flag | Description |
|------|-------------|
| `-w` | Write XML files to disk (one per service) |
| `-f` | Overwrite existing XML files |
| `-e` | Include host environment variables and `.env` file |
| `-v` | Verbose output |

## Installing templates on Unraid

Unraid picks up user templates from `/boot/config/plugins/dockerMan/templates-user/`. Save your generated XML files there and they will appear in the Unraid Docker UI under **Add Container > Template**.

```bash
cat docker-compose.yml | docker run -i \
  -v /boot/config/plugins/dockerMan/templates-user:/output \
  ghcr.io/ogglord/comp2unraid -w -
```

## Build from source

```bash
git clone https://github.com/Ogglord/comp2unraid.git
cd comp2unraid
make
./bin/comp2unraid_linux_amd64 examples/compose/docker-compose.yml
```
