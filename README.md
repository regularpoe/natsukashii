# natsukashii

This program lets you search for some keywords across the git history for one file at the time.

## Preparation

Clone repository and build:

```bash
go build -o <your_binary_name>
```

or

```bash
CGO_ENABLED=0 GOOS=linux go build -a -o natsukashii
```

## Usage

Navigate to the directory you want to inspect, and simply run

```bash
natsukashii -s <your_file>
```

This will process file and run server on port 1987 for viewing.

