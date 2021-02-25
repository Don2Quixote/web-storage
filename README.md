# Storage
Little web service to store you files at

## Installation
Installed Go compiler required to compile or run code
Firstly, download project:
```bash
git clone https://github.com/Don2Quixote/web-storage
cd web-storage
```

Then install dependencies:
```bash
go get github.com/go-sql-driver/mysql
go get github.com/don2quixote/ninja
```
Then build:
```bash
go build -o ./web-storage src/*.go
```

# Configurating
Edit file `config.json`. It's easy to understand what to input

## Launching
```bash
sudo -E ./web-storage
```

