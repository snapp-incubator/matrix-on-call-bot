package config

const defaultConfig = `
matrix:
  url: "https://example.com"
  userID: "@some_bot:example.com"
  token: "secret"
  display-name: "user"

server:
  listen: "0.0.0.0:8080"

database:
  driver: mysql
  conn-str: "on-call:secret@tcp(localhost:33060)/on-call?readTimeout=3s&timeout=30s&parseTime=True"
  options:
    connection-lifetime: "10m"
    max-open-connections: 10
    max-idle-connections: 5
`
