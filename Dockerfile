# Użyj oficjalnego obrazu Golang jako bazowego
FROM golang:latest

# Stwórz katalog /build i ustaw go jako główny folder roboczy
RUN mkdir /build
WORKDIR /build

# Sklonuj repozytorium z GitHub
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/lomakopakula/Learning_GO.git .

# Przejdź do katalogu z aplikacją
WORKDIR /build/http_server

# Zainstaluj zależności
RUN go mod tidy

# Skompiluj aplikację
RUN go build -o server server.go

# Uruchom aplikację
CMD ["./server"]